package redisdb

import (
	"context"
	"fmt"
	"time"

	"github.com/jehincastic/go-boilerplate/internal/config"
	"github.com/redis/go-redis/v9"
)

// Open connects to Redis and checks the connection with PING.
func Open(ctx context.Context, cfg config.Redis) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		if closeErr := rdb.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping redis: %w (close: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return rdb, nil
}
