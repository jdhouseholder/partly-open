package partly_open

import (
	"encoding/json"
	"fmt"
	"io"
)

type NopLogger struct{}

func (s *NopLogger) Log(WorkLogEntry) error { return nil }

type StdOutLogger struct{}

func (s *StdOutLogger) Log(w WorkLogEntry) error {
	fmt.Printf("%+v\n", w)
	return nil
}

type MemLogger struct {
	logs []WorkLogEntry
}

func (m *MemLogger) Log(w WorkLogEntry) error {
	m.logs = append(m.logs, w)
	return nil
}

func (m *MemLogger) Logs() []WorkLogEntry {
	return m.logs
}

type JsonLogger struct {
	w io.Writer
}

func NewJsonLogger(w io.Writer) *JsonLogger {
	return &JsonLogger{w}
}

func (m *JsonLogger) Log(w WorkLogEntry) error {
	if err := json.NewEncoder(m.w).Encode(w); err != nil {
		return err
	}
	_, err := m.w.Write([]byte{'\n'})
	return err
}
