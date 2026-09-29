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

// FailedValidation writes a 422 response containing field errors.
func FailedValidation(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	Error(w, r, http.StatusUnprocessableEntity, errors)
}

// EditConflict writes a 409 response.
func EditConflict(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusConflict, "unable to update the record due to an edit conflict, please try again")
}

// RateLimitExceeded writes a 429 response.
func RateLimitExceeded(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusTooManyRequests, "rate limit exceeded")
}

// InvalidCredentials writes a 401 response for a failed login.
func InvalidCredentials(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusUnauthorized, "invalid authentication credentials")
}

// InvalidAuthenticationToken writes a 401 response for a missing or bad bearer token.
func InvalidAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	Error(w, r, http.StatusUnauthorized, "invalid or missing authentication token")
}

// AuthenticationRequired writes a 401 response.
func AuthenticationRequired(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusUnauthorized, "you must be authenticated to access this resource")
}

// InactiveAccount writes a 403 response.
func InactiveAccount(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusForbidden, "your user account must be activated to access this resource")
}

// NotPermitted writes a 403 response.
func NotPermitted(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusForbidden, "your user account doesn't have the necessary permissions to access this resource")
}
