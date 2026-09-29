package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const Version = "1.0.0"

// Config is the process configuration loaded from the environment.
type Config struct {
	Port  int
	Env   string
	DB    Database
	Redis Redis
	Auth  Auth
	CORS  CORS
}

// Database holds PostgreSQL pool settings.
type Database struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
}

// Redis holds the connection used by session cache and per-route rate limiters.
type Redis struct {
	Addr     string
	Password string
	DB       int
}

// Auth holds JWT signing settings.
type Auth struct {
	JWTSecret string
	TokenTTL  time.Duration
}

// CORS holds browser origins allowed to call the API.
type CORS struct {
	TrustedOrigins []string
}

// Load reads a local .env file when one is present, then reads the environment.
// A missing .env file is ignored so the same binary can run in containers.
func Load() (Config, error) {
	_ = godotenv.Load()

	maxIdleTime, err := time.ParseDuration(getenv("DB_CONN_MAX_IDLE_TIME", "15m"))
	if err != nil {
		return Config{}, fmt.Errorf("DB_CONN_MAX_IDLE_TIME: %w", err)
	}
	tokenTTL, err := time.ParseDuration(getenv("JWT_TTL", "24h"))
	if err != nil {
		return Config{}, fmt.Errorf("JWT_TTL: %w", err)
	}

	cfg := Config{
		Port: getenvInt("APP_PORT", 4000),
		Env:  getenv("APP_ENV", "development"),
		DB: Database{
			DSN:          getenv("DB_DSN", ""),
			MaxOpenConns: getenvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getenvInt("DB_MAX_IDLE_CONNS", 25),
			MaxIdleTime:  maxIdleTime,
		},
		Redis: Redis{
			Addr:     getenv("REDIS_ADDR", "localhost:6379"),
			Password: getenv("REDIS_PASSWORD", ""),
			DB:       getenvInt("REDIS_DB", 0),
		},
		Auth: Auth{
			JWTSecret: getenv("JWT_SECRET", ""),
			TokenTTL:  tokenTTL,
		},
		CORS: CORS{
			TrustedOrigins: splitList(getenv("CORS_TRUSTED_ORIGINS", "")),
		},
	}

	if cfg.DB.DSN == "" {
		return Config{}, fmt.Errorf("DB_DSN is required")
	}
	if cfg.Port < 1 {
		return Config{}, fmt.Errorf("APP_PORT must be greater than zero")
	}
	if cfg.Redis.Addr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR is required")
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}
	if cfg.Auth.TokenTTL <= 0 {
		return Config{}, fmt.Errorf("JWT_TTL must be greater than zero")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' '
	})
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			origins = append(origins, part)
		}
	}
	return origins
}
