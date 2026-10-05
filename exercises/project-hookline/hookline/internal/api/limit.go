package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/hookline/internal/problem"
	"example.com/hookline/internal/ratelimit"
)

// rateLimit applies the authenticated tenant's own limit (tenants.rate_per_sec, tenants.burst) to the
// next handler. Over the limit: 429 with Retry-After. Within it: the RateLimit fields tell the client
// its budget. When the limiter fails (Redis down) the request is served: failing open is the decision
// of chapter 37, section 37.6, because failing closed would turn a cache outage into an ingest outage.
func (s *server) rateLimit(next http.Handler) http.Handler {
	if s.limiter == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := tenantFrom(r.Context())
		d, err := s.limiter.Allow(r.Context(), t.ID, float64(t.RatePerSec), t.Burst)
		if err != nil {
			if s.onLimit != nil {
				s.onLimit(err)
			}
			s.log.WarnContext(r.Context(), "rate limiter unavailable; serving without a limit",
				slog.String("tenant_id", t.ID), slog.String("error", err.Error()))
			next.ServeHTTP(w, r)
			return
		}
		h := w.Header()
		h.Set("RateLimit-Policy", fmt.Sprintf(`"ingest";q=%d;w=%d`, t.Burst, max(1, (t.Burst+t.RatePerSec-1)/t.RatePerSec)))
		h.Set("RateLimit", `"ingest";r=`+strconv.Itoa(d.Remaining))
		if !d.Allowed {
			h.Set("Retry-After", ratelimit.RetryAfterSeconds(d.RetryAfter))
			problem.Write(w, r, problem.RateLimited.New(fmt.Sprintf(
				"this tenant may send %d events per second (bursts of %d); retry after the Retry-After seconds", t.RatePerSec, t.Burst)))
			return
		}
		next.ServeHTTP(w, r)
	})
}
