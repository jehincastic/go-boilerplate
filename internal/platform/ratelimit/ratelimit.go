package ratelimit

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/jehincastic/go-boilerplate/internal/platform/httpserver"
	"github.com/redis/go-redis/v9"
)

const script = `
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current
`

// Policy is the request budget for one route, counted per client IP.
type Policy struct {
	Name     string
	Requests int
	Window   time.Duration
}

var (
	// Login is tight because it checks credentials.
	Login = Policy{Name: "login", Requests: 5, Window: time.Minute}
	// Register is tighter because it creates accounts.
	Register = Policy{Name: "register", Requests: 3, Window: time.Hour}
)

// Middleware limits requests for one route. The limit is always enforced.
func Middleware(rdb *redis.Client, policy Policy) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := httprate.CanonicalizeIP(middleware.GetClientIP(r.Context()))
			if ip == "" {
				ip = r.RemoteAddr
			}
			key := fmt.Sprintf("ratelimit:%s:%s", policy.Name, ip)

			count, err := rdb.Eval(r.Context(), script, []string{key}, policy.Window.Milliseconds()).Int64()
			if err != nil {
				httpserver.ServerError(w, r, nil, err)
				return
			}
			if count > int64(policy.Requests) {
				httpserver.RateLimitExceeded(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
