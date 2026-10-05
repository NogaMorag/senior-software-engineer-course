// Package httpapi is the HTTP face of the catalog. Its contract is api/openapi.yaml: every route,
// status code and error body here is documented there first.
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"example.com/ch09-catalog-api/internal/catalog"
	"example.com/ch09-catalog-api/internal/idem"
	"example.com/ch09-catalog-api/internal/problem"
)

// Options are the knobs of the HTTP API.
type Options struct {
	MaxBodyBytes int64
	Logger       *slog.Logger
}

// Route is one documented operation; the spec test checks that api/openapi.yaml lists each of them.
type Route struct{ Method, Path string }

// Routes lists every operation New registers, in the spec's path syntax.
var Routes = []Route{
	{"GET", "/healthz"},
	{"GET", "/v1/items"},
	{"POST", "/v1/items"},
	{"GET", "/v1/items/{id}"},
	{"GET", "/v1/products/{id}"},
}

// The deprecated alias /v1/products/{id}: deprecated since, and removed at, these instants.
var (
	productsDeprecatedAt = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	productsSunsetAt     = time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC)
)

type server struct {
	store catalog.Store
	log   *slog.Logger
}

// New wires routes and middleware. The order, outermost first: request id, logging, recover; then per
// route the body limit and, for POST, the idempotency middleware.
func New(store catalog.Store, keys *idem.Store, opt Options) http.Handler {
	if opt.Logger == nil {
		opt.Logger = slog.New(slog.DiscardHandler)
	}
	if opt.MaxBodyBytes <= 0 {
		opt.MaxBodyBytes = 16 << 10
	}
	s := &server{store: store, log: opt.Logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /v1/items", s.listItems)
	mux.Handle("POST /v1/items", limitBody(opt.MaxBodyBytes, keys.Middleware(http.HandlerFunc(s.createItem))))
	mux.HandleFunc("GET /v1/items/{id}", s.getItem)
	mux.Handle("GET /v1/products/{id}", deprecated(productsDeprecatedAt, productsSunsetAt,
		func(r *http.Request) string { return "/v1/items/" + r.PathValue("id") }, http.HandlerFunc(s.getItem)))

	// A pattern without a method is less specific than the same pattern with one, so these catch
	// every other method and answer 405 in the problem format instead of the mux's plain text.
	mux.Handle("/v1/items", methodNotAllowed("GET, POST"))
	mux.Handle("/v1/items/{id}", methodNotAllowed("GET"))
	mux.Handle("/v1/products/{id}", methodNotAllowed("GET"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		problem.Write(w, r, problem.NotFound.New("no such route"))
	})
	return requestID(logging(opt.Logger, recoverer(opt.Logger, mux)))
}

func methodNotAllowed(allow string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allow)
		problem.Write(w, r, problem.MethodNotAllowed.New(r.Method+" is not supported here; allowed: "+allow))
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// createItem handles POST /v1/items: 201 with a Location header, or a problem.
func (s *server) createItem(w http.ResponseWriter, r *http.Request) {
	// TODO(day 19): decode the body with decodeJSON, validate it (a validation problem carries the
	// field errors), create the item (a duplicate SKU is a conflict problem, any other error goes to
	// s.internal), then answer 201 with a Location header and the item.
	panic("TODO(day 19): createItem")
}

// getItem handles GET /v1/items/{id} (and its deprecated alias).
func (s *server) getItem(w http.ResponseWriter, r *http.Request) {
	// TODO(day 19): look the item up by r.PathValue("id"); ErrNotFound is a not-found problem,
	// any other error goes to s.internal; success is 200 with the item.
	panic("TODO(day 19): getItem")
}

// listItems handles GET /v1/items?limit=&cursor= with cursor pagination.
func (s *server) listItems(w http.ResponseWriter, r *http.Request) {
	// TODO(day 20): parse limit with page.ParseLimit (a validation problem with pointer "query:limit"),
	// decode the cursor if present (an invalid-cursor problem), call s.store.List, and answer a
	// page.Page whose NextCursor encodes the last item when more exist. An empty page is [], not null.
	panic("TODO(day 20): listItems")
}

// internal logs the real error and returns a problem that does not leak it.
func (s *server) internal(w http.ResponseWriter, r *http.Request, err error) {
	s.log.Error("internal error", "path", r.URL.Path, "err", err)
	problem.Write(w, r, problem.Internal.New("unexpected error; the request id identifies it in the logs"))
}
