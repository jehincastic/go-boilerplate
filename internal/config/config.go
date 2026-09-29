package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config is the process configuration loaded from the environment.
type Config struct {
	Port int
	Env  string
	CORS CORS
}

// CORS holds browser origins allowed to call the API.
type CORS struct {
	TrustedOrigins []string
}

// Load reads a local .env file when one is present, then reads the environment.
// A missing .env file is ignored so the same binary can run in containers.
func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Port: getenvInt("APP_PORT", 4000),
		Env:  getenv("APP_ENV", "development"),
		CORS: CORS{
			TrustedOrigins: splitList(getenv("CORS_TRUSTED_ORIGINS", "")),
		},
	}

	if cfg.Port < 1 {
		return Config{}, fmt.Errorf("APP_PORT must be greater than zero")
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
