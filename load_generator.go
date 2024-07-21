package partly_open

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type WorkId struct {
	WorkerId  int `json:"worker-id"`
	RequestId int `json:"request-id"`
}

type WorkLogEntry struct {
	WorkId WorkId    `json:"work-id"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

type WorkLogger interface {
	Log(WorkLogEntry) error
}

type DoWorker interface {
	DoWork(context.Context, WorkId) error
}

type DoWorkFunc func(context.Context, WorkId) error

func (dwf DoWorkFunc) DoWork(ctx context.Context, workId WorkId) error {
	return dwf(ctx, workId)
}

type LoadGeneratorConfig struct {
	// Mean arrival rate.
	MeanNewWorkersPerSecond float64
	ArrivalJitter           time.Duration
	MaxWorkers              int
	StayProbability         float64
	ThinkTime               time.Duration
	MaxRequests             uint64
}

func (lgc *LoadGeneratorConfig) validate() error {
	if lgc.MeanNewWorkersPerSecond <= 0 {
		return fmt.Errorf("MeanNewWorkersPerSecond must be greater than zero, got=%v", lgc.MeanNewWorkersPerSecond)
	}
	if lgc.MaxWorkers <= 0 {
		return fmt.Errorf("MaxWorkers must be greater than zero, got=%v", lgc.MaxWorkers)
	}
	if lgc.StayProbability < 0 || lgc.StayProbability > 1 {
		return fmt.Errorf("StayProbability must be in [0, 1], got=%v", lgc.StayProbability)
	}
	return nil
}

type LoadGenerator struct {
	cfg  *LoadGeneratorConfig
	wg   *sync.WaitGroup
	rand *rand.Rand

	arrivalJitter int64
	sleepFor      int64

	nRequests atomic.Uint64

	workLogger WorkLogger

	doWorker DoWorker
	once     sync.Once
	err      error
}

func NewLoadGeneratorFromDoWorkFunc(
	cfg *LoadGeneratorConfig,
	workLogger WorkLogger,
	dwf DoWorkFunc,
) (*LoadGenerator, error) {
	r := rand.New(rand.NewSource(0))
	return NewLoadGeneratorWithRand(cfg, r, workLogger, DoWorker(dwf))
}

func NewLoadGenerator(
	cfg *LoadGeneratorConfig,
	workLogger WorkLogger,
	doWorker DoWorker,
) (*LoadGenerator, error) {
	r := rand.New(rand.NewSource(0))
	return NewLoadGeneratorWithRand(cfg, r, workLogger, doWorker)
}

func NewLoadGeneratorWithRand(
	cfg *LoadGeneratorConfig,
	rand *rand.Rand,
	workLogger WorkLogger,
	doWorker DoWorker,
) (*LoadGenerator, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	var sleepFor int64
	if cfg.MeanNewWorkersPerSecond <= 1_000_000_000 {
		sleepFor = int64(1_000_000_000. / cfg.MeanNewWorkersPerSecond)
	}
	arrivalJitter := cfg.ArrivalJitter.Nanoseconds()

	if arrivalJitter > sleepFor {
		return nil, fmt.Errorf("ArrivalJitter must be less than sleep=%vns required for MeanNewWorkersPerSecond=%v", sleepFor, cfg.MeanNewWorkersPerSecond)
	}

	return &LoadGenerator{
		cfg:           cfg,
		wg:            &sync.WaitGroup{},
		rand:          rand,
		arrivalJitter: arrivalJitter,
		sleepFor:      sleepFor,
		workLogger:    workLogger,
		doWorker:      doWorker,
	}, nil
}

func (lg *LoadGenerator) GenerateLoad(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	for i := 0; i < lg.cfg.MaxWorkers; i++ {
		lg.wg.Add(1)
		go lg.doWork(ctx, cancel, i)

		if lg.arrivalJitter > 0 {
			arrivalJitter := lg.rand.Int63n(lg.arrivalJitter)
			if lg.rand.Intn(2) == 1 {
				time.Sleep(time.Duration(lg.sleepFor + arrivalJitter))
			} else {
				time.Sleep(time.Duration(lg.sleepFor - arrivalJitter))
			}
		} else {
			time.Sleep(time.Duration(lg.sleepFor))
		}
	}

	lg.wg.Wait()

	return lg.err
}

func (lg *LoadGenerator) doWork(ctx context.Context, cancel context.CancelFunc, workerId int) {
	defer lg.wg.Done()
	for i := 0; ; i++ {
		workId := WorkId{
			WorkerId:  workerId,
			RequestId: i,
		}

		if lg.cfg.MaxRequests > 0 {
			if n := lg.nRequests.Add(1); n > lg.cfg.MaxRequests {
				return
			}
		}

		start := time.Now()
		if err := lg.doWorker.DoWork(ctx, workId); err != nil {
			lg.once.Do(func() { lg.err = err })
			cancel()
			return
		}
		end := time.Now()

		workLogEntry := WorkLogEntry{
			WorkId: workId,
			Start:  start,
			End:    end,
		}

		if err := lg.workLogger.Log(workLogEntry); err != nil {
			lg.once.Do(func() { lg.err = err })
			cancel()
			return
		}

		if x := lg.rand.Float64(); x > lg.cfg.StayProbability {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		if lg.cfg.ThinkTime > 0 {
			time.Sleep(lg.cfg.ThinkTime)

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}
}
