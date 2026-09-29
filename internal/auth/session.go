package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jehincastic/go-boilerplate/internal/platform/errs"
	"github.com/redis/go-redis/v9"
)

// Session is the user profile cached for authenticated requests.
type Session struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type contextKey struct{}

// SaveSession stores the user profile until ttl elapses.
func SaveSession(ctx context.Context, rdb *redis.Client, session Session, ttl time.Duration) error {
	payload, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	if err := rdb.Set(ctx, sessionKey(session.ID), payload, ttl).Err(); err != nil {
		return fmt.Errorf("store session: %w", err)
	}
	return nil
}

// LoadSession reads the user profile cached for id.
func LoadSession(ctx context.Context, rdb *redis.Client, id uuid.UUID) (Session, error) {
	payload, err := rdb.Get(ctx, sessionKey(id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return Session{}, errs.ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("load session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(payload, &session); err != nil {
		return Session{}, fmt.Errorf("unmarshal session: %w", err)
	}
	return session, nil
}

// WithSession returns a context that carries session.
func WithSession(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, contextKey{}, session)
}

// SessionFromContext returns the user profile stored by the auth middleware.
func SessionFromContext(ctx context.Context) (Session, bool) {
	session, ok := ctx.Value(contextKey{}).(Session)
	return session, ok
}

func sessionKey(id uuid.UUID) string {
	return "auth:user:" + id.String()
}
