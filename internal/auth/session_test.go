package auth

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/jehincastic/go-boilerplate/internal/platform/errs"
	"github.com/redis/go-redis/v9"
)

func TestSaveAndLoadSession(t *testing.T) {
	srv := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	session := Session{
		ID:        uuid.MustParse("018f4f4a-7c3a-7b2a-8c1d-6e5f4a3b2c1d"),
		Name:      "Ada Lovelace",
		Email:     "ada@example.com",
		CreatedAt: time.Date(2026, 9, 29, 5, 0, 0, 0, time.UTC),
	}
	if err := SaveSession(context.Background(), client, session, time.Hour); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadSession(context.Background(), client, session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != session.Name || loaded.Email != session.Email || !loaded.CreatedAt.Equal(session.CreatedAt) {
		t.Fatalf("loaded = %+v", loaded)
	}
}

func TestLoadSessionMissing(t *testing.T) {
	srv := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	_, err := LoadSession(context.Background(), client, uuid.MustParse("018f4f4a-7c3a-7b2a-8c1d-6e5f4a3b2c1d"))
	if err != errs.ErrNotFound {
		t.Fatalf("err = %v", err)
	}
}
