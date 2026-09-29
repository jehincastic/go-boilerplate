package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jehincastic/go-boilerplate/internal/config"
	"github.com/jehincastic/go-boilerplate/internal/restaurent"
)

// Deps is everything the router needs from the composition root.
type Deps struct {
	Config config.Config
	Logger *slog.Logger
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
		restHandler := restaurent.NewHandler(deps.Logger)
		apiR.Get("/slots", restHandler.GetAvailableSlots)
		apiR.Post("/book", restHandler.BookSlot)
	})

	return r
}
