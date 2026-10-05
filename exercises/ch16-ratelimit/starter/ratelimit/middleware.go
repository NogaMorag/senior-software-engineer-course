package ratelimit

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

// Options configures Middleware.
type Options struct {
	Limiter Limiter
	Key     func(*http.Request) string // which bucket a request draws from
	// Rate and Burst describe the policy in the RateLimit-Policy header.
	Rate  float64
	Burst int
	// FailOpen serves the request when the limiter fails (Redis down); otherwise 503.
	FailOpen bool
	OnError  func(error) // called on every limiter failure, e.g. to count and log it
}

// Middleware refuses requests over the limit with 429 and Retry-After, and tells every
// client its budget in the RateLimit and RateLimit-Policy fields (an IETF draft: see the
// chapter for its status).
func Middleware(o Options, next http.Handler) http.Handler {
	window := int(math.Ceil(float64(o.Burst) / o.Rate))
	policy := fmt.Sprintf(`"default";q=%d;w=%d`, o.Burst, window)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d, err := o.Limiter.Allow(r.Context(), o.Key(r))
		if err != nil {
			if o.OnError != nil {
				o.OnError(err)
			}
			if o.FailOpen {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limiter unavailable", http.StatusServiceUnavailable)
			return
		}
		h := w.Header()
		h.Set("RateLimit-Policy", policy)
		h.Set("RateLimit", fmt.Sprintf(`"default";r=%d`, d.Remaining))
		if !d.Allowed {
			h.Set("Retry-After", RetryAfterSeconds(d.RetryAfter))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RetryAfterSeconds formats a wait as Retry-After's delta-seconds: whole seconds, rounded
// up so a client never comes back too early, and at least 1.
func RetryAfterSeconds(d time.Duration) string {
	return strconv.Itoa(max(1, int(math.Ceil(d.Seconds()))))
}
