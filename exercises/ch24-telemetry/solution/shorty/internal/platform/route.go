package platform

import "net/http"

// RoutePattern sets r.Pattern before the request reaches the mux. Added for
// chapter 24.
//
// ServeMux records the pattern it matched on the request it was handed, but
// every middleware that changes the context (RequestID does) hands on a copy.
// The layers outside that copy, otelhttp among them, would never see the
// pattern, and spans and metrics would lose their route. Asking the mux up
// front costs one extra lookup and keeps the route on every layer.
func RoutePattern(mux *http.ServeMux) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, pattern := mux.Handler(r); pattern != "" {
				r = r.WithContext(r.Context()) // a shallow copy; never modify the caller's request
				r.Pattern = pattern
			}
			next.ServeHTTP(w, r)
		})
	}
}
