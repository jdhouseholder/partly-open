package partly_open

import (
	"context"
	"testing"
)

func TestGenerateLoad(t *testing.T) {
	cfg := LoadGeneratorConfig{
		MeanNewWorkersPerSecond: 1,
		MaxWorkers:              1,
	}
	logger := &MemLogger{}
	nopDoWork := func(context.Context, WorkId) error { return nil }
	lg, err := NewLoadGeneratorFromDoWorkFunc(
		&cfg,
		logger,
		nopDoWork,
	)
	if err != nil {
		t.Fatal(err)
	}

	lg.GenerateLoad(context.TODO())

	logs := logger.Logs()
	if got := len(logs); got != 1 {
		t.Fatalf("Want one log entry: got=%v, want=%v", got, 1)
	}
	logEntry := logs[0]
	if got := logEntry.WorkId.WorkerId; got != 0 {
		t.Fatalf("logEntry.WorkerId.WorkerId: got=%v, want=%v", got, 0)
	}
	if got := logEntry.WorkId.RequestId; got != 0 {
		t.Fatalf("logEntry.WorkerId.RequestId: got=%v, want=%v", got, 0)
	}
}

// TODO: We probably want to inject a timesource and sleep implementation to
// make this testable. Since it is pretty simple, I'll omit that for now.
