// Package shed is load-shedding middleware: past a limit of requests in flight, it answers
// 503 with Retry-After at once instead of letting work queue up without bound. Part of the
// limit can be kept for high-priority requests, so they are the last to be shed.
package shed

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

// Priority orders requests for shedding: Low goes first.
type Priority int

// The two priorities.
const (
	Low Priority = iota
	High
)

// Options configures a Shedder.
type Options struct {
	Limit           int           // requests in flight, all priorities together
	ReservedForHigh int           // part of Limit that low-priority requests may not use
	RetryAfter      time.Duration // suggested wait for a shed client (whole seconds, at least 1)
	// Priority classifies a request (default: every request is Low).
	Priority func(*http.Request) Priority
}

// Shedder counts requests in flight and sheds the ones over its limit.
type Shedder struct {
	o          Options
	retryAfter string
	inFlight   atomic.Int64
	shed       [2]atomic.Int64 // per priority
	served     atomic.Int64
}

// New returns a Shedder.
func New(o Options) *Shedder {
	if o.Priority == nil {
		o.Priority = func(*http.Request) Priority { return Low }
	}
	o.ReservedForHigh = min(max(o.ReservedForHigh, 0), o.Limit)
	secs := max(1, int(o.RetryAfter.Round(time.Second)/time.Second))
	return &Shedder{o: o, retryAfter: strconv.Itoa(secs)}
}

// Wrap returns next behind the shedder.
func (s *Shedder) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := s.o.Priority(r)
		limit := s.o.Limit
		if p == Low {
			limit -= s.o.ReservedForHigh
		}
		// Count first, then check: two requests racing for the last slot cannot both win.
		if n := s.inFlight.Add(1); n > int64(limit) {
			s.inFlight.Add(-1)
			s.shed[p].Add(1)
			w.Header().Set("Retry-After", s.retryAfter)
			http.Error(w, "overloaded, try again later", http.StatusServiceUnavailable)
			return
		}
		defer s.inFlight.Add(-1)
		s.served.Add(1)
		next.ServeHTTP(w, r)
	})
}

// InFlight, Served and Shed report the counters.
func (s *Shedder) InFlight() int64       { return s.inFlight.Load() }
func (s *Shedder) Served() int64         { return s.served.Load() }
func (s *Shedder) Shed(p Priority) int64 { return s.shed[p].Load() }

// ByHeader returns a Priority function that treats requests carrying header with value
// as High. Only trust such a header from your own infrastructure, never from the internet.
func ByHeader(header, value string) func(*http.Request) Priority {
	return func(r *http.Request) Priority {
		if r.Header.Get(header) == value {
			return High
		}
		return Low
	}
}
