package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/jehincastic/go-boilerplate/internal/auth"
	"github.com/jehincastic/go-boilerplate/internal/config"
	"github.com/jehincastic/go-boilerplate/internal/user"
	"github.com/redis/go-redis/v9"
)

type stubPinger struct {
	err error
}

func (s stubPinger) Ping(context.Context) error { return s.err }

type stubUsers struct{}

func (stubUsers) Register(context.Context, string, string, string) (*user.User, error) {
	return &user.User{Email: "ada@example.com"}, nil
}

func (stubUsers) Login(context.Context, string, string) (auth.Token, error) {
	return auth.Token{AccessToken: "signed-token"}, nil
}

const testJWTSecret = "test-secret-must-be-at-least-32-bytes"

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testRouter(t *testing.T, database, cache Pinger) (http.Handler, *redis.Client) {
	t.Helper()
	srv := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	log := testLogger()
	handler := NewRouter(Deps{
		Config:      config.Config{Env: "development", Auth: config.Auth{JWTSecret: testJWTSecret}},
		Logger:      log,
		Users:       user.NewHandler(stubUsers{}, log),
		Database:    database,
		Redis:       cache,
		RedisClient: client,
		Version:     "test",
	})
	return handler, client
}

func TestHealthcheckAvailable(t *testing.T) {
	router, _ := testRouter(t, stubPinger{}, stubPinger{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestHealthcheckDatabaseUnavailable(t *testing.T) {
	router, _ := testRouter(t, stubPinger{err: errors.New("down")}, stubPinger{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRoutesHaveSeparateRateLimits(t *testing.T) {
	router, _ := testRouter(t, stubPinger{}, stubPinger{})

	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "203.0.113.20:1234"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("login %d was limited", i)
		}
	}

	blocked := httptest.NewRequest(http.MethodPost, "/login", nil)
	blocked.RemoteAddr = "203.0.113.20:1234"
	blockedRec := httptest.NewRecorder()
	router.ServeHTTP(blockedRec, blocked)
	if blockedRec.Code != http.StatusTooManyRequests {
		t.Fatalf("login status = %d, body = %s", blockedRec.Code, blockedRec.Body.String())
	}

	register := httptest.NewRequest(http.MethodPost, "/register", nil)
	register.RemoteAddr = "203.0.113.20:1234"
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, register)
	if registerRec.Code == http.StatusTooManyRequests {
		t.Fatalf("register status = %d, body = %s", registerRec.Code, registerRec.Body.String())
	}

	health := httptest.NewRequest(http.MethodGet, "/health", nil)
	health.RemoteAddr = "203.0.113.20:1234"
	healthRec := httptest.NewRecorder()
	router.ServeHTTP(healthRec, health)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("health status = %d, body = %s", healthRec.Code, healthRec.Body.String())
	}
}

func TestMeReturnsCachedUser(t *testing.T) {
	router, client := testRouter(t, stubPinger{}, stubPinger{})
	userID := uuid.MustParse("018f4f4a-7c3a-7b2a-8c1d-6e5f4a3b2c1d")
	err := auth.SaveSession(context.Background(), client, auth.Session{
		ID:        userID,
		Name:      "Ada Lovelace",
		Email:     "ada@example.com",
		CreatedAt: time.Now().UTC(),
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := auth.Sign(userID, "ada@example.com", testJWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.RemoteAddr = "203.0.113.30:1234"
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Ada Lovelace") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestMeRejectsMissingSession(t *testing.T) {
	router, _ := testRouter(t, stubPinger{}, stubPinger{})
	userID := uuid.MustParse("018f4f4a-7c3a-7b2a-8c1d-6e5f4a3b2c1d")
	token, err := auth.Sign(userID, "ada@example.com", testJWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.RemoteAddr = "203.0.113.31:1234"
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
