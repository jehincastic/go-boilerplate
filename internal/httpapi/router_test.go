package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jehincastic/go-boilerplate/internal/config"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testRouter() http.Handler {
	return NewRouter(Deps{
		Config: config.Config{Env: "development"},
		Logger: testLogger(),
	})
}

func TestGetSlots(t *testing.T) {
	router := testRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/slots", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"avaiableTables"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestBookSlot(t *testing.T) {
	router := testRouter()
	body := `{"tableId":"` + uuid.NewString() + `","slotTime":"10:00","capacity":1,"bookingFor":"2026-09-29T12:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/book", strings.NewReader(body))
	req.RemoteAddr = "203.0.113.11:1234"
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"booking"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestBookSlotRejectsOverCapacity(t *testing.T) {
	router := testRouter()
	body := `{"tableId":"` + uuid.NewString() + `","slotTime":"10:00","capacity":2,"bookingFor":"2026-09-29T12:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/book", strings.NewReader(body))
	req.RemoteAddr = "203.0.113.12:1234"
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "max capacity") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}
