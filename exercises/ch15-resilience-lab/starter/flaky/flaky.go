// Package flaky is an upstream that misbehaves on purpose: every request waits Latency,
// then fails with 503 at ErrorRate. It counts what happened to each request.
package flaky

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Counts is a snapshot of what the server saw.
type Counts struct {
	Received  int64 // requests that arrived
	Succeeded int64 // answered 200
	Failed    int64 // answered 503 on purpose
	Abandoned int64 // the caller went away while the server was still working
}

func (c Counts) String() string {
	return fmt.Sprintf("received=%d ok=%d failed=%d abandoned=%d", c.Received, c.Succeeded, c.Failed, c.Abandoned)
}

// Server is an http.Handler. Change its behaviour at any time with Set.
type Server struct {
	mu        sync.Mutex
	errorRate float64
	latency   time.Duration
	rnd       *rand.Rand

	received, succeeded, failed, abandoned atomic.Int64
}

// New returns a server; seed makes the sequence of failures repeatable.
func New(errorRate float64, latency time.Duration, seed uint64) *Server {
	return &Server{errorRate: errorRate, latency: latency, rnd: rand.New(rand.NewPCG(seed, seed))}
}

// Set changes the error rate and latency for the requests that arrive from now on.
func (s *Server) Set(errorRate float64, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorRate, s.latency = errorRate, latency
}

// Counts returns a snapshot of the counters.
func (s *Server) Counts() Counts {
	return Counts{s.received.Load(), s.succeeded.Load(), s.failed.Load(), s.abandoned.Load()}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.received.Add(1)
	s.mu.Lock()
	latency, fail := s.latency, s.rnd.Float64() < s.errorRate
	s.mu.Unlock()

	if latency > 0 {
		t := time.NewTimer(latency)
		defer t.Stop()
		select {
		case <-t.C:
		case <-r.Context().Done():
			s.abandoned.Add(1) // nobody is waiting for this answer any more
			return
		}
	}
	if fail {
		s.failed.Add(1)
		http.Error(w, "upstream: injected failure", http.StatusServiceUnavailable)
		return
	}
	s.succeeded.Add(1)
	fmt.Fprintln(w, "ok")
}
