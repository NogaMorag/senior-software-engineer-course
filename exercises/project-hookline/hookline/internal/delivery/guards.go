package delivery

import (
	"math/rand/v2"
	"sync"
	"time"

	"example.com/hookline/internal/breaker"
	"example.com/hookline/internal/bulkhead"
)

// Guards are the per-endpoint protections of one worker process (milestone M6, chapter 15): a circuit
// breaker that stops sending to an endpoint that keeps failing, and a bulkhead that caps how many of
// the process's workers one endpoint can hold at once. State is local to the process: simpler, fails
// safe, and each worker learns within FailureThreshold attempts (chapter 37, section 37.5).
type Guards struct {
	Breaker     breaker.Settings // per endpoint; zero fields take the breaker's defaults
	Concurrency int              // workers one endpoint may hold; 0 means no bulkhead
	// OnStateChange, when set, is told about every breaker transition, for logs and metrics.
	OnStateChange func(endpointID string, from, to breaker.State)

	mu sync.Mutex
	m  map[string]*guard
}

type guard struct {
	breaker  *breaker.Breaker
	bulkhead *bulkhead.Bulkhead // nil when Concurrency is 0
}

// get returns the endpoint's guard, creating it on first use. Endpoints are never forgotten in this
// version; milestone M8's metrics show how many there are.
func (g *Guards) get(endpointID string) *guard {
	g.mu.Lock()
	defer g.mu.Unlock()
	if gd, ok := g.m[endpointID]; ok {
		return gd
	}
	if g.m == nil {
		g.m = map[string]*guard{}
	}
	s := g.Breaker
	if cb := g.OnStateChange; cb != nil {
		s.OnStateChange = func(from, to breaker.State) { cb(endpointID, from, to) }
	}
	gd := &guard{breaker: breaker.New(s)}
	if g.Concurrency > 0 {
		gd.bulkhead = bulkhead.New(g.Concurrency, 0) // never wait: a busy endpoint's delivery is postponed
	}
	g.m[endpointID] = gd
	return gd
}

// State reports the breaker state of an endpoint, Closed for one never seen.
func (g *Guards) State(endpointID string) breaker.State {
	g.mu.Lock()
	gd, ok := g.m[endpointID]
	g.mu.Unlock()
	if !ok {
		return breaker.Closed
	}
	return gd.breaker.State()
}

// openWait is how long a delivery waits when its endpoint's breaker is open: about the open timeout,
// spread by up to a fifth so that the postponed deliveries of one endpoint do not return together.
func (g *Guards) openWait() time.Duration {
	open := g.Breaker.OpenTimeout
	if open <= 0 {
		open = 10 * time.Second // the breaker's default
	}
	return open + rand.N(open/5+1)
}

// busyWait is how long a delivery waits when its endpoint's bulkhead is full.
func busyWait() time.Duration { return 200*time.Millisecond + rand.N(800*time.Millisecond) }

// breakerError is what a send means for the endpoint's health: only a failure worth retrying (no
// answer, 408, 429, 5xx) counts. A 4xx or a redirect is a working endpoint giving a final answer.
func breakerError(res Result) error {
	if res.OK() || !res.Retryable {
		return nil
	}
	return res.Err
}
