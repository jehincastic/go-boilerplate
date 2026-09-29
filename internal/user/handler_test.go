package user

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jehincastic/go-boilerplate/internal/auth"
	"github.com/jehincastic/go-boilerplate/internal/platform/errs"
)

type fakeService struct {
	user *User
	err  error
}

func (f *fakeService) Register(_ context.Context, name, email, _ string) (*User, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.user = &User{ID: uuid.MustParse("018f4f4a-7c3a-7b2a-8c1d-6e5f4a3b2c1d"), Name: name, Email: email}
	return f.user, nil
}

func (f *fakeService) Login(context.Context, string, string) (auth.Token, error) {
	if f.err != nil {
		return auth.Token{}, f.err
	}
	return auth.Token{AccessToken: "signed-token"}, nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRegisterValidation(t *testing.T) {
	handler := NewHandler(&fakeService{}, testLogger())
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"email":"not-an-email","password":"short"}`))
	rec := httptest.NewRecorder()
	handler.Register(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRegisterCreated(t *testing.T) {
	handler := NewHandler(&fakeService{}, testLogger())
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"name":"Ada Lovelace","email":"ada@example.com","password":"pa55word"}`))
	rec := httptest.NewRecorder()
	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "ada@example.com") || !strings.Contains(rec.Body.String(), "Ada Lovelace") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestMeReadsContext(t *testing.T) {
	handler := NewHandler(&fakeService{}, testLogger())
	session := auth.Session{
		ID:    uuid.MustParse("018f4f4a-7c3a-7b2a-8c1d-6e5f4a3b2c1d"),
		Name:  "Ada Lovelace",
		Email: "ada@example.com",
	}
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req = req.WithContext(auth.WithSession(req.Context(), session))
	rec := httptest.NewRecorder()
	handler.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Ada Lovelace") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestMeRequiresSession(t *testing.T) {
	handler := NewHandler(&fakeService{}, testLogger())
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	handler.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	handler := NewHandler(&fakeService{err: errs.ErrInvalidCredentials}, testLogger())
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"ada@example.com","password":"pa55word"}`))
	rec := httptest.NewRecorder()
	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
