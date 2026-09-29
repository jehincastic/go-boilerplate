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
	"github.com/jehincastic/go-boilerplate/internal/restaurent"
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

	r.Route("/api", func(apiR chi.Router) {
		apiR.Get("/health", healthcheck(deps))
		apiR.Route("/auth", func(authR chi.Router) {
			authR.With(ratelimit.Middleware(deps.RedisClient, ratelimit.Register)).Post("/register", deps.Users.Register)
			authR.With(ratelimit.Middleware(deps.RedisClient, ratelimit.Login)).Post("/login", deps.Users.Login)
		})
		apiR.Group(func(authR chi.Router) {
			authR.Use(auth.Middleware(deps.RedisClient, deps.Config.Auth.JWTSecret, deps.Logger))
			authR.Get("/me", deps.Users.Me)
		})

		restHandler := restaurent.NewHandler(deps.Logger)
		apiR.Get("/slots", restHandler.GetAvailableSlots)
		apiR.Post("/book", restHandler.BookSlot)
	})

	return r
}
