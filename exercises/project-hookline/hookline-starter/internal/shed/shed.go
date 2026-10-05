// Package shed is load-shedding middleware: past a limit of requests in flight, it answers
// 503 with Retry-After at once instead of letting work queue up without bound. Part of the
// limit can be kept for high-priority requests, so they are the last to be shed. In Hookline the
// answer is the problem type "overloaded" (spec N6).
package shed

import (
	"net/http"

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
	panic("TODO: see chapter 38")

}

// Wrap returns next behind the shedder.
func (s *Shedder) Wrap(next http.Handler) http.Handler {
	panic("TODO: see chapter 38")

}

// InFlight, Served and Shed report the counters.
func (s *Shedder) InFlight() int64       { panic("TODO: see chapter 38") }
func (s *Shedder) Served() int64         { panic("TODO: see chapter 38") }
func (s *Shedder) Shed(p Priority) int64 { panic("TODO: see chapter 38") }

// ByHeader returns a Priority function that treats requests carrying header with value
// as High. Only trust such a header from your own infrastructure, never from the internet.
func ByHeader(header, value string) func(*http.Request) Priority {
	panic("TODO: see chapter 38")

}
