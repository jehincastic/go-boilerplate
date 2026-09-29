package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jehincastic/go-boilerplate/internal/config"
	"github.com/jehincastic/go-boilerplate/internal/httpapi"
	"github.com/jehincastic/go-boilerplate/internal/platform/background"
	"github.com/jehincastic/go-boilerplate/internal/platform/database"
	"github.com/jehincastic/go-boilerplate/internal/platform/httpserver"
	applog "github.com/jehincastic/go-boilerplate/internal/platform/logger"
	redisdb "github.com/jehincastic/go-boilerplate/internal/platform/redisdb"
	"github.com/jehincastic/go-boilerplate/internal/user"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stdout, nil)).Error("configuration error", "error", err)
		os.Exit(1)
	}

	log := applog.New(os.Stdout, applog.LevelForEnv(cfg.Env))

	db, err := database.Open(context.Background(), cfg.DB, cfg.Env)
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("database pool failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Error("close database", "error", err)
		}
	}()
	log.Info("database connection pool established")

	redisClient, err := redisdb.Open(context.Background(), cfg.Redis)
	if err != nil {
		log.Error("redis connection failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Error("close redis", "error", err)
		}
	}()
	log.Info("redis connection established")

	userService := user.NewService(user.NewRepository(db), redisClient, cfg.Auth.JWTSecret, cfg.Auth.TokenTTL)

	handler := httpapi.NewRouter(httpapi.Deps{
		Config:      cfg,
		Logger:      log,
		Users:       user.NewHandler(userService, log),
		Database:    dbPinger{db: db},
		Redis:       redisdb.Pinger{Client: redisClient},
		RedisClient: redisClient,
		Version:     config.Version,
	})

	err = httpserver.Serve(httpserver.Options{
		Addr:      httpserver.Addr(cfg.Port),
		Env:       cfg.Env,
		Handler:   handler,
		Logger:    log,
		WaitGroup: &background.WaitGroup,
	})
	if err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

type dbPinger struct {
	db *gorm.DB
}

func (p dbPinger) Ping(ctx context.Context) error {
	return database.Ping(ctx, p.db)
}
