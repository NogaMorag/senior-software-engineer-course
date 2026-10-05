// Package api is Hookline's HTTP API. Its contract is api/openapi.yaml: every route, status code and
// error body here is documented there first, and spec_test.go checks that the two agree.
package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/platform"
	"example.com/hookline/internal/problem"
	"example.com/hookline/internal/ratelimit"
	"example.com/hookline/internal/shed"
	"example.com/hookline/internal/store"
)

// Options are the knobs of the API.
type Options struct {
	MaxBodyBytes int64            // limit for every JSON request body
	DevMode      bool             // accept http:// endpoint URLs
	Logger       *slog.Logger     // nil discards
	Now          func() time.Time // nil means time.Now; tests inject a clock
	// Subscriptions is told about every committed change to a tenant's subscriptions or endpoints, so
	// the dispatcher's cache drops the tenant's entry (milestone M3). nil means no cache.
	Subscriptions Invalidator
	// Limiter limits each tenant's ingest to its own rate_per_sec and burst (milestone M6, spec N5).
	// nil means no limit. A limiter error fails open: the request is served, logged, and reported to
	// OnLimiterError (milestone M8 counts it as a metric).
	Limiter        ratelimit.Limiter
	OnLimiterError func(error)
	// Shedder sheds /v1 requests over its in-flight limit with 503 (spec N6). nil means no shedding.
	Shedder *shed.Shedder
}

// Invalidator drops cached state of a tenant; *subcache.Cache satisfies it.
type Invalidator interface {
	Invalidate(ctx context.Context, tenantID string) error
}

// Route is one documented operation; the spec test checks that api/openapi.yaml lists each of them.
type Route struct{ Method, Path string }

// Routes lists every operation New registers, in the spec's path syntax.
var Routes = []Route{
	{"GET", "/healthz"},
	{"GET", "/v1/tenant"},
	{"GET", "/v1/endpoints"},
	{"POST", "/v1/endpoints"},
	{"GET", "/v1/endpoints/{id}"},
	{"POST", "/v1/endpoints/{id}/disable"},
	{"GET", "/v1/subscriptions"},
	{"POST", "/v1/subscriptions"},
	{"GET", "/v1/subscriptions/{id}"},
	{"DELETE", "/v1/subscriptions/{id}"},
	{"POST", "/v1/events"},
	{"GET", "/v1/events/{id}"},
	{"GET", "/v1/events/{id}/deliveries"},
	{"GET", "/v1/deliveries"},
	{"POST", "/v1/deliveries/{id}/replay"},
}

type server struct {
	store   store.Store
	log     *slog.Logger
	devMode bool
	now     func() time.Time
	subs    Invalidator
	limiter ratelimit.Limiter
	onLimit func(error)
}

// New wires the routes and the middleware. Outermost first: request id, access log, recover; then, per
// route, authentication and the body limit.
func New(st store.Store, opt Options) http.Handler {
	if opt.Logger == nil {
		opt.Logger = slog.New(slog.DiscardHandler)
	}
	if opt.MaxBodyBytes <= 0 {
		opt.MaxBodyBytes = 64 << 10
	}
	if opt.Now == nil {
		opt.Now = time.Now
	}
	s := &server{store: st, log: opt.Logger, devMode: opt.DevMode, now: opt.Now, subs: opt.Subscriptions,
		limiter: opt.Limiter, onLimit: opt.OnLimiterError}
	auth := s.authenticate
	body := func(h http.HandlerFunc) http.Handler { return auth(limitBody(opt.MaxBodyBytes, h)) }

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		platform.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.Handle("GET /v1/tenant", auth(http.HandlerFunc(s.getTenant)))
	mux.Handle("GET /v1/endpoints", auth(http.HandlerFunc(s.listEndpoints)))
	mux.Handle("POST /v1/endpoints", body(s.createEndpoint))
	mux.Handle("GET /v1/endpoints/{id}", auth(http.HandlerFunc(s.getEndpoint)))
	mux.Handle("POST /v1/endpoints/{id}/disable", auth(http.HandlerFunc(s.disableEndpoint)))
	mux.Handle("GET /v1/subscriptions", auth(http.HandlerFunc(s.listSubscriptions)))
	mux.Handle("POST /v1/subscriptions", body(s.createSubscription))
	mux.Handle("GET /v1/subscriptions/{id}", auth(http.HandlerFunc(s.getSubscription)))
	mux.Handle("DELETE /v1/subscriptions/{id}", auth(http.HandlerFunc(s.deleteSubscription)))
	mux.Handle("POST /v1/events", auth(s.rateLimit(limitBody(opt.MaxBodyBytes, http.HandlerFunc(s.ingestEvent)))))
	mux.Handle("GET /v1/events/{id}", auth(http.HandlerFunc(s.getEvent)))
	mux.Handle("GET /v1/events/{id}/deliveries", auth(http.HandlerFunc(s.eventDeliveries)))
	mux.Handle("GET /v1/deliveries", auth(http.HandlerFunc(s.deadDeliveries)))
	mux.Handle("POST /v1/deliveries/{id}/replay", auth(http.HandlerFunc(s.replayDelivery)))

	// A pattern without a method is less specific than the same pattern with one, so these catch every
	// other method and answer 405 in the problem format instead of the mux's plain text.
	for path, allow := range map[string]string{
		"/healthz": "GET", "/v1/tenant": "GET",
		"/v1/endpoints": "GET, POST", "/v1/endpoints/{id}": "GET", "/v1/endpoints/{id}/disable": "POST",
		"/v1/subscriptions": "GET, POST", "/v1/subscriptions/{id}": "GET, DELETE",
		"/v1/events": "POST", "/v1/events/{id}": "GET", "/v1/events/{id}/deliveries": "GET",
		"/v1/deliveries": "GET", "/v1/deliveries/{id}/replay": "POST",
	} {
		mux.Handle(path, methodNotAllowed(allow))
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		problem.Write(w, r, problem.NotFound.New("no such route"))
	})
	var h http.Handler = mux
	if opt.Shedder != nil {
		h = shedV1(opt.Shedder, mux)
	}
	return platform.Chain(h, platform.RequestID, platform.AccessLog(opt.Logger), platform.Recover(opt.Logger))
}

// shedV1 puts the shedder in front of every /v1 route and leaves /healthz alone: a probe must answer
// even, and especially, when the API is overloaded. Shedding comes before authentication and the rate
// limiter, so a shed request costs no database or Redis call (chapter 16, section 16.5).
func shedV1(s *shed.Shedder, next http.Handler) http.Handler {
	guarded := s.Wrap(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/") {
			guarded.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func methodNotAllowed(allow string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allow)
		problem.Write(w, r, problem.MethodNotAllowed.New(r.Method+" is not supported here; allowed: "+allow))
	})
}

type tenantKey struct{}

// tenantFrom returns the tenant the auth middleware put in the context. Every handler behind it uses
// this tenant's id in every store call; nothing in a request can choose another tenant.
func tenantFrom(ctx context.Context) domain.Tenant {
	t, _ := ctx.Value(tenantKey{}).(domain.Tenant)
	return t
}

// internal logs the real error and answers with a problem that does not leak it.
func (s *server) internal(w http.ResponseWriter, r *http.Request, err error) {
	s.log.ErrorContext(r.Context(), "internal error", slog.String("path", r.URL.Path), slog.String("error", err.Error()))
	problem.Write(w, r, problem.Internal.New("unexpected error; the request id identifies it in the logs"))
}

// notFound answers 404 for a resource of kind with id. Another tenant's id gets exactly this answer.
func notFound(w http.ResponseWriter, r *http.Request, kind, id string) {
	problem.Write(w, r, problem.NotFound.New("no "+kind+" with id "+id))
}

// changed tells the subscription cache that the tenant's subscriptions or endpoints changed. The change
// is already committed, so a failure here is logged, not returned: the entry's TTL bounds the staleness.
func (s *server) changed(r *http.Request, tenantID string) {
	if s.subs == nil {
		return
	}
	if err := s.subs.Invalidate(r.Context(), tenantID); err != nil {
		s.log.WarnContext(r.Context(), "subscription cache invalidation failed; the entry expires by TTL",
			slog.String("tenant_id", tenantID), slog.String("error", err.Error()))
	}
}

func validation(w http.ResponseWriter, r *http.Request, errs []problem.FieldError) {
	p := problem.Validation.New("one or more inputs are invalid")
	p.Errors = errs
	problem.Write(w, r, p)
}

// writeJSON is platform.WriteJSON, named locally because every handler uses it.
func writeJSON(w http.ResponseWriter, status int, v any) { platform.WriteJSON(w, status, v) }
