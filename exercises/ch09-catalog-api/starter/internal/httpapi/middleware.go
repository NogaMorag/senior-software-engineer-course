package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"example.com/ch09-catalog-api/internal/problem"
)

type ctxKey struct{}

// requestID reuses a sane incoming X-Request-ID or makes a new one, and echoes it in the response.
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" || len(id) > 64 {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	if s.status == 0 {
		s.status = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusWriter) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

// logging writes one structured line per request.
func logging(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		id, _ := r.Context().Value(ctxKey{}).(string)
		log.Info("request", "method", r.Method, "path", r.URL.Path, "status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(), "request_id", id)
	})
}

// recoverer turns a panic into a 500 problem instead of a dropped connection.
func recoverer(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				log.Error("panic", "path", r.URL.Path, "value", v)
				problem.Write(w, r, problem.Internal.New("unexpected error; the request id identifies it in the logs"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// limitBody caps the request body before anything reads it. Reading past the cap fails with
// *http.MaxBytesError, which decodeJSON and the idempotency middleware turn into 413.
func limitBody(maxBytes int64, next http.Handler) http.Handler {
	// TODO(day 21): reject a declared Content-Length over maxBytes with a TooLarge problem, and wrap
	// r.Body in http.MaxBytesReader so an undeclared (chunked) body is cut off too.
	return next
}

// deprecated marks every response of a route as deprecated (RFC 9745) with a sunset date (RFC 8594)
// and a link to the replacement.
func deprecated(since, sunset time.Time, successor func(*http.Request) string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Deprecation", "@"+strconv.FormatInt(since.Unix(), 10))
		h.Set("Sunset", sunset.UTC().Format(http.TimeFormat))
		h.Set("Link", "<"+successor(r)+`>; rel="successor-version"`)
		next.ServeHTTP(w, r)
	})
}
