package httpserver

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func logError(r *http.Request, log *slog.Logger, err error) {
	if log == nil {
		return
	}
	log.Error("request error",
		"error", err,
		"request_method", r.Method,
		"request_url", r.URL.String(),
		"request_id", middleware.GetReqID(r.Context()),
	)
}

// Error writes a JSON error envelope.
func Error(w http.ResponseWriter, r *http.Request, status int, message any) {
	if err := WriteJSON(w, status, Envelope{"error": message}, nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ServerError logs err and writes a generic 500 response.
func ServerError(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	logError(r, log, err)
	Error(w, r, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

// NotFound writes a 404 response.
func NotFound(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusNotFound, "the requested resource could not be found")
}

// MethodNotAllowed writes a 405 response.
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	Error(w, r, http.StatusMethodNotAllowed, message)
}

// BadRequest writes a 400 response.
func BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	Error(w, r, http.StatusBadRequest, err.Error())
}
