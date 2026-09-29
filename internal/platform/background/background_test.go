package background

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestGoWaitsForJob(t *testing.T) {
	done := make(chan struct{})
	Go(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), func() {
		close(done)
	})

	WaitGroup.Wait()
	select {
	case <-done:
	default:
		t.Fatal("background job did not finish before Wait")
	}
}

func TestGoRecoversPanic(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, nil))

	Go(log, func() {
		panic("boom")
	})
	WaitGroup.Wait()

	if buf.Len() == 0 {
		t.Fatal("expected panic to be logged")
	}
}
