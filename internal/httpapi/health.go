package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/jehincastic/go-boilerplate/internal/platform/httpserver"
)

func healthcheck(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "available"
		databaseStatus := "available"
		redisStatus := "available"
		code := http.StatusOK

		if err := ping(r.Context(), deps.Database, "database"); err != nil {
			deps.Logger.Error("database healthcheck failed", "error", err)
			status = "unavailable"
			databaseStatus = "unavailable"
			code = http.StatusServiceUnavailable
		}
		if err := ping(r.Context(), deps.Redis, "redis"); err != nil {
			deps.Logger.Error("redis healthcheck failed", "error", err)
			status = "unavailable"
			redisStatus = "unavailable"
			code = http.StatusServiceUnavailable
		}

		err := httpserver.WriteJSON(w, code, httpserver.Envelope{
			"status":      status,
			"environment": deps.Config.Env,
			"version":     deps.Version,
			"database":    databaseStatus,
			"redis":       redisStatus,
		}, nil)
		if err != nil {
			httpserver.ServerError(w, r, deps.Logger, err)
		}
	}
}

func ping(ctx context.Context, pinger Pinger, name string) error {
	if pinger == nil {
		return errors.New(name + " is not configured")
	}
	return pinger.Ping(ctx)
}
