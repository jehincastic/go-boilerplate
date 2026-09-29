package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jehincastic/go-boilerplate/internal/platform/errs"
	"github.com/jehincastic/go-boilerplate/internal/platform/httpserver"
	"github.com/redis/go-redis/v9"
)

// Middleware verifies the bearer token, loads the cached user, and stores it on the request context.
func Middleware(rdb *redis.Client, secret string, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := bearerToken(r)
			if err != nil {
				httpserver.InvalidAuthenticationToken(w, r)
				return
			}

			userID, err := UserID(token, secret)
			if err != nil {
				httpserver.InvalidAuthenticationToken(w, r)
				return
			}

			session, err := LoadSession(r.Context(), rdb, userID)
			if err != nil {
				if errors.Is(err, errs.ErrNotFound) {
					httpserver.InvalidAuthenticationToken(w, r)
					return
				}
				httpserver.ServerError(w, r, log, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(WithSession(r.Context(), session)))
		})
	}
}

func bearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	parts := strings.Split(header, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", errors.New("invalid bearer token")
	}
	return parts[1], nil
}
