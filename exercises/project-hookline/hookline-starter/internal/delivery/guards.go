package delivery

import (
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
	panic("TODO: see chapter 38")

}

// State reports the breaker state of an endpoint, Closed for one never seen.
func (g *Guards) State(endpointID string) breaker.State {
	panic("TODO: see chapter 38")

}

// openWait is how long a delivery waits when its endpoint's breaker is open: about the open timeout,
// spread by up to a fifth so that the postponed deliveries of one endpoint do not return together.
func (g *Guards) openWait() time.Duration {
	panic("TODO: see chapter 38")

}

// busyWait is how long a delivery waits when its endpoint's bulkhead is full.
func busyWait() time.Duration { panic("TODO: see chapter 38") }

// breakerError is what a send means for the endpoint's health: only a failure worth retrying (no
// answer, 408, 429, 5xx) counts. A 4xx or a redirect is a working endpoint giving a final answer.
func breakerError(res Result) error {
	panic("TODO: see chapter 38")

}
