package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jehincastic/go-boilerplate/internal/auth"
	"github.com/jehincastic/go-boilerplate/internal/config"
	"github.com/jehincastic/go-boilerplate/internal/platform/ratelimit"
	"github.com/jehincastic/go-boilerplate/internal/user"
	"github.com/redis/go-redis/v9"
)

// Pinger checks that a dependency can serve traffic.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Deps is everything the router needs from the composition root.
type Deps struct {
	Config      config.Config
	Logger      *slog.Logger
	Users       *user.Handler
	Database    Pinger
	Redis       Pinger
	RedisClient *redis.Client
	Version     string
}

// NewRouter builds the chi router, middleware stack, and API routes.
func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()
	r.NotFound(notFound)
	r.MethodNotAllowed(methodNotAllowed)

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(requestLogger(deps.Logger))
	r.Use(recoverPanic(deps.Logger))
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.CleanPath)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   deps.Config.CORS.TrustedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", healthcheck(deps))
	r.With(ratelimit.Middleware(deps.RedisClient, ratelimit.Register)).Post("/register", deps.Users.Register)
	r.With(ratelimit.Middleware(deps.RedisClient, ratelimit.Login)).Post("/login", deps.Users.Login)
	r.Group(func(authR chi.Router) {
		authR.Use(auth.Middleware(deps.RedisClient, deps.Config.Auth.JWTSecret, deps.Logger))
		authR.Get("/me", deps.Users.Me)
	})

	return r
}
