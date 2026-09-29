package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadJSONRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Ada","extra":true}`))
	rec := httptest.NewRecorder()

	var dst struct {
		Name string `json:"name"`
	}
	err := ReadJSON(rec, req, &dst)
	if err == nil || !strings.Contains(err.Error(), "unknown key") {
		t.Fatalf("error = %v", err)
	}
}

func TestReadJSONRejectsMultipleValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Ada"}{"name":"Grace"}`))
	rec := httptest.NewRecorder()

	var dst struct {
		Name string `json:"name"`
	}
	err := ReadJSON(rec, req, &dst)
	if err == nil || !strings.Contains(err.Error(), "single JSON value") {
		t.Fatalf("error = %v", err)
	}
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	headers := make(http.Header)
	headers.Set("Location", "/v1/movies/1")

	err := WriteJSON(rec, http.StatusCreated, Envelope{"movie": map[string]any{"id": 1}}, headers)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content type = %s", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("Location") != "/v1/movies/1" {
		t.Fatalf("location = %s", rec.Header().Get("Location"))
	}
	if !strings.Contains(rec.Body.String(), `"id":1`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}
