package logger

import (
	"io"
	"log/slog"
)

// New returns a JSON logger writing to w.
func New(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}

// LevelForEnv uses debug logs in development and info logs everywhere else.
func LevelForEnv(env string) slog.Level {
	if env == "development" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}
