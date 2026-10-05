package platform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
)

// Middleware wraps a handler with behaviour that runs before and after it.
type Middleware func(http.Handler) http.Handler

// Chain wraps h with mws; the first middleware is the outermost, so it sees
// the request first and the response last.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// RequestIDHeader carries the request ID in both directions.
const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

// WithRequestID returns a copy of ctx that carries id.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFrom returns the request ID stored in ctx, or "".
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// RequestID keeps a well-formed incoming X-Request-ID (so one ID can follow a
// request across services) or makes a new one, puts it in the context and
// echoes it in the response.
func RequestID(next http.Handler) http.Handler {
	// TODO 5 (day 12, section 6.2): keep a valid incoming X-Request-ID or make a
	// new one (newRequestID), set it on the response header, and call next with
	// r.WithContext(WithRequestID(r.Context(), id)).
	return next // until then: no request ID at all
}

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // never fails; see crypto/rand.Read
	return hex.EncodeToString(b)
}

// validRequestID accepts 1 to 64 characters from [A-Za-z0-9._-]; anything else
// from the outside is replaced, so it cannot inject text into the logs.
func validRequestID(id string) bool {
	if len(id) == 0 || len(id) > 64 {
		return false
	}
	for _, c := range id {
		ok := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-'
		if !ok {
			return false
		}
	}
	return true
}

// statusWriter remembers the status code and the number of body bytes.
type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
}

// wrap returns w itself when it is already a *statusWriter, so stacked
// middleware share one view of the response.
func wrap(w http.ResponseWriter) *statusWriter {
	if sw, ok := w.(*statusWriter); ok {
		return sw
	}
	return &statusWriter{ResponseWriter: w}
}

func (w *statusWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += int64(n)
	return n, err
}

// Unwrap lets http.ResponseController reach the underlying writer (Flush, deadlines).
func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// AccessLog writes one structured line per request after it has been served.
func AccessLog(log *slog.Logger) Middleware {
	// TODO 6 (day 12, section 6.4): return a Middleware that wraps w (wrap), calls
	// next, then logs one "request" record with log.LogAttrs(r.Context(), ...):
	// method, path, pattern (r.Pattern), status (200 if nothing was written),
	// bytes and duration.
	return func(next http.Handler) http.Handler { return next } // until then: no access log
}

// Recover turns a panic in a handler into a logged error and a 500 response,
// so one bad request does not take the connection, or the process, with it.
func Recover(log *slog.Logger) Middleware {
	// TODO 7 (day 12, section 6.2): return a Middleware whose handler defers a
	// function that calls recover(). On a panic: re-panic http.ErrAbortHandler,
	// otherwise log "panic in handler" with the value and debug.Stack(), and send
	// WriteError(sw, 500, "internal error") if nothing was written yet.
	return func(next http.Handler) http.Handler { return next } // until then: a panic drops the connection
}
