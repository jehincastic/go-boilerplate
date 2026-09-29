package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Options configures the HTTP server process.
type Options struct {
	Addr      string
	Env       string
	Handler   http.Handler
	Logger    *slog.Logger
	WaitGroup *sync.WaitGroup
}

// Serve listens until SIGINT or SIGTERM, then waits for in-flight work to finish.
func Serve(opts Options) error {
	errLog := slog.NewLogLogger(opts.Logger.Handler(), slog.LevelError)
	srv := &http.Server{
		Addr:         opts.Addr,
		Handler:      opts.Handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     errLog,
	}

	shutdownError := make(chan error, 1)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		caught := <-quit

		opts.Logger.Info("caught signal", "signal", caught.String())

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			shutdownError <- err
			return
		}

		opts.Logger.Info("completing background tasks", "addr", srv.Addr)
		if opts.WaitGroup != nil {
			opts.WaitGroup.Wait()
		}
		shutdownError <- nil
	}()

	opts.Logger.Info("starting server", "addr", srv.Addr, "env", opts.Env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownError; err != nil {
		return err
	}

	opts.Logger.Info("stopped server", "addr", srv.Addr)
	return nil
}

// Addr formats a TCP address for the given port.
func Addr(port int) string {
	return fmt.Sprintf(":%d", port)
}
