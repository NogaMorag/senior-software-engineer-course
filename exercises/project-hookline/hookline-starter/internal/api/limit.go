package api

import (
	"net/http"
)

// rateLimit applies the authenticated tenant's own limit (tenants.rate_per_sec, tenants.burst) to the
// next handler. Over the limit: 429 with Retry-After. Within it: the RateLimit fields tell the client
// its budget. When the limiter fails (Redis down) the request is served: failing open is the decision
// of chapter 37, section 37.6, because failing closed would turn a cache outage into an ingest outage.
func (s *server) rateLimit(next http.Handler) http.Handler {
	panic("TODO: see chapter 38")

}
