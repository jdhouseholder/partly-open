package partly_open

import (
	"bytes"
	"testing"
)

func TestJsonLogger(t *testing.T) {
	var b bytes.Buffer
	lg := NewJsonLogger(&b)

	if err := lg.Log(WorkLogEntry{
		WorkId: WorkId{
			WorkerId:  123,
			RequestId: 456,
		},
	}); err != nil {
		t.Fatal(err)
	}

	got := b.String()
	want := `{"work-id":{"worker-id":123,"request-id":456},"start":"0001-01-01T00:00:00Z","end":"0001-01-01T00:00:00Z"}` + "\n"
	if got != want {
		t.Fatalf("got=%v, want=%v", got, want)
	}

	if err := lg.Log(WorkLogEntry{
		WorkId: WorkId{
			WorkerId:  789,
			RequestId: 987,
		},
	}); err != nil {
		t.Fatal(err)
	}

	got = b.String()
	want = `{"work-id":{"worker-id":123,"request-id":456},"start":"0001-01-01T00:00:00Z","end":"0001-01-01T00:00:00Z"}` + "\n" +
		`{"work-id":{"worker-id":789,"request-id":987},"start":"0001-01-01T00:00:00Z","end":"0001-01-01T00:00:00Z"}` + "\n"
	if got != want {
		t.Fatalf("got=%v, want=%v", got, want)
	}
}
