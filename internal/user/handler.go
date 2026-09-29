package user

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jehincastic/go-boilerplate/internal/auth"
	"github.com/jehincastic/go-boilerplate/internal/platform/errs"
	"github.com/jehincastic/go-boilerplate/internal/platform/httpserver"
	"github.com/jehincastic/go-boilerplate/internal/platform/validator"
)

// ServiceAPI is the user service surface used by HTTP handlers.
type ServiceAPI interface {
	Register(ctx context.Context, name, email, password string) (*User, error)
	Login(ctx context.Context, email, password string) (auth.Token, error)
}

// Handler serves register, login, and the current user.
type Handler struct {
	svc ServiceAPI
	log *slog.Logger
}

// NewHandler returns a user HTTP handler.
func NewHandler(svc ServiceAPI, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httpserver.ReadJSON(w, r, &input); err != nil {
		httpserver.BadRequest(w, r, err)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	v := validator.New()
	validator.ValidateName(v, input.Name)
	validator.ValidateEmail(v, input.Email)
	validator.ValidatePasswordPlaintext(v, input.Password)
	if err := v.Err(); err != nil {
		var vErr *validator.Error
		errors.As(err, &vErr)
		httpserver.FailedValidation(w, r, vErr.Errors)
		return
	}

	user, err := h.svc.Register(r.Context(), input.Name, input.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			httpserver.FailedValidation(w, r, v.Errors)
		default:
			httpserver.ServerError(w, r, h.log, err)
		}
		return
	}

	if err := httpserver.WriteJSON(w, http.StatusCreated, httpserver.Envelope{"user": user}, nil); err != nil {
		httpserver.ServerError(w, r, h.log, err)
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httpserver.ReadJSON(w, r, &input); err != nil {
		httpserver.BadRequest(w, r, err)
		return
	}

	v := validator.New()
	validator.ValidateEmail(v, input.Email)
	validator.ValidatePasswordPlaintext(v, input.Password)
	if err := v.Err(); err != nil {
		var vErr *validator.Error
		errors.As(err, &vErr)
		httpserver.FailedValidation(w, r, vErr.Errors)
		return
	}

	token, err := h.svc.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrInvalidCredentials):
			httpserver.InvalidCredentials(w, r)
		default:
			httpserver.ServerError(w, r, h.log, err)
		}
		return
	}

	if err := httpserver.WriteJSON(w, http.StatusOK, httpserver.Envelope{"authentication_token": token}, nil); err != nil {
		httpserver.ServerError(w, r, h.log, err)
	}
}

// Me returns the authenticated user stored on the request context.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		httpserver.AuthenticationRequired(w, r)
		return
	}

	if err := httpserver.WriteJSON(w, http.StatusOK, httpserver.Envelope{"user": session}, nil); err != nil {
		httpserver.ServerError(w, r, h.log, err)
	}
}
