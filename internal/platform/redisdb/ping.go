package redisdb

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Pinger checks Redis.
type Pinger struct {
	Client *redis.Client
}

// Ping checks that Redis is reachable.
func (p Pinger) Ping(ctx context.Context) error {
	return p.Client.Ping(ctx).Err()
}
