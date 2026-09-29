package main

import (
	"log/slog"
	"os"

	"github.com/jehincastic/go-boilerplate/internal/config"
	"github.com/jehincastic/go-boilerplate/internal/httpapi"
	"github.com/jehincastic/go-boilerplate/internal/platform/httpserver"
	applog "github.com/jehincastic/go-boilerplate/internal/platform/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stdout, nil)).Error("configuration error", "error", err)
		os.Exit(1)
	}

	log := applog.New(os.Stdout, applog.LevelForEnv(cfg.Env))

	handler := httpapi.NewRouter(httpapi.Deps{
		Config: cfg,
		Logger: log,
	})

	err = httpserver.Serve(httpserver.Options{
		Addr:    httpserver.Addr(cfg.Port),
		Env:     cfg.Env,
		Handler: handler,
		Logger:  log,
	})
	if err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
