package background

import (
	"log/slog"
	"sync"
)

// WaitGroup tracks in-flight background jobs for the process.
// Pass it to httpserver.Options so shutdown waits until every job finishes.
var WaitGroup sync.WaitGroup

// Go runs fn in a new goroutine tracked by WaitGroup.
// A panic in fn is recovered and logged so one job does not stop the process.
func Go(log *slog.Logger, fn func()) {
	WaitGroup.Add(1)
	go func() {
		defer WaitGroup.Done()
		defer func() {
			if rec := recover(); rec != nil && log != nil {
				log.Error("background job panicked", "error", rec)
			}
		}()
		fn()
	}()
}
