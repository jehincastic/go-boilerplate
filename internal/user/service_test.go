package user

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/jehincastic/go-boilerplate/internal/auth"
	"github.com/jehincastic/go-boilerplate/internal/platform/errs"
	"github.com/redis/go-redis/v9"
)

type memRepo struct {
	user *User
}

func (m *memRepo) Insert(_ context.Context, user *User) error {
	user.ID = uuid.MustParse("018f4f4a-7c3a-7b2a-8c1d-6e5f4a3b2c1d")
	user.CreatedAt = time.Date(2026, 9, 29, 5, 0, 0, 0, time.UTC)
	copied := *user
	m.user = &copied
	return nil
}

func (m *memRepo) GetByEmail(_ context.Context, email string) (*User, error) {
	if m.user == nil || m.user.Email != email {
		return nil, errs.ErrNotFound
	}
	return m.user, nil
}

func TestRegisterAndLoginCacheTheUser(t *testing.T) {
	srv := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	repo := &memRepo{}
	service := NewService(repo, client, "test-secret-must-be-at-least-32-bytes", time.Hour)

	created, err := service.Register(context.Background(), "Ada Lovelace", "ada@example.com", "pa55word")
	if err != nil {
		t.Fatal(err)
	}

	cached, err := auth.LoadSession(context.Background(), client, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cached.Name != "Ada Lovelace" || cached.Email != "ada@example.com" {
		t.Fatalf("cached = %+v", cached)
	}

	if _, err := service.Login(context.Background(), "ada@example.com", "pa55word"); err != nil {
		t.Fatal(err)
	}
	cached, err = auth.LoadSession(context.Background(), client, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cached.Name != "Ada Lovelace" {
		t.Fatalf("cached after login = %+v", cached)
	}
}
