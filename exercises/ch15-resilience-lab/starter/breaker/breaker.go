// Package breaker is a circuit breaker with three states and an injected clock.
//
// Closed: calls pass; consecutive failures are counted. Open: calls fail fast with
// ErrOpen until OpenTimeout has passed. HalfOpen: a few trial calls pass; enough
// successes close the breaker, one failure opens it again.
package breaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// State is the breaker's state.
type State int

// The three states.
const (
	Closed State = iota
	Open
	HalfOpen
)

func (s State) String() string {
	switch s {
	case Closed:
		return "closed"
	case Open:
		return "open"
	case HalfOpen:
		return "half-open"
	}
	return fmt.Sprintf("State(%d)", int(s))
}

var (
	// ErrOpen is returned without calling anything while the breaker is open.
	ErrOpen = errors.New("breaker: open")
	// ErrTooManyProbes is returned while half-open when every trial slot is taken.
	ErrTooManyProbes = errors.New("breaker: half-open, probe limit reached")
)

// Settings configures a Breaker. Zero fields take the defaults in New.
type Settings struct {
	FailureThreshold int           // consecutive failures that open a closed breaker (default 5)
	OpenTimeout      time.Duration // time spent open before the first probe (default 10s)
	HalfOpenProbes   int           // trial calls allowed while half-open; that many successes close it (default 1)
	// IsFailure decides which errors count against the dependency (default: any error
	// except context.Canceled, which means our caller gave up, not that the dependency failed).
	IsFailure func(error) bool
	Now       func() time.Time // the clock (default time.Now); tests inject a fake one
	// OnStateChange is called with the lock held: keep it short and never call the breaker from it.
	OnStateChange func(from, to State)
}

// Breaker is safe for concurrent use.
type Breaker struct {
	s Settings

	mu         sync.Mutex
	state      State
	generation uint64 // bumped on every state change; results from an older generation are ignored
	failures   int    // consecutive failures while closed
	openedAt   time.Time
	probes     int // trial calls started while half-open
	successes  int // successful trial calls while half-open
}

// New returns a closed breaker.
func New(s Settings) *Breaker {
	if s.FailureThreshold <= 0 {
		s.FailureThreshold = 5
	}
	if s.OpenTimeout <= 0 {
		s.OpenTimeout = 10 * time.Second
	}
	if s.HalfOpenProbes <= 0 {
		s.HalfOpenProbes = 1
	}
	if s.IsFailure == nil {
		s.IsFailure = func(err error) bool { return err != nil && !errors.Is(err, context.Canceled) }
	}
	if s.Now == nil {
		s.Now = time.Now
	}
	return &Breaker{s: s}
}

// State reports the current state; an open breaker whose timeout has passed reports HalfOpen.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.current()
}

// Allow asks permission for one call. When it returns a nil error, the caller makes the
// call and then calls done exactly once with the call's error.
func (b *Breaker) Allow() (done func(error), err error) {
	// TODO(day 47): under the lock, switch on b.current():
	//   Open: return ErrOpen.  HalfOpen: ErrTooManyProbes when probes >= HalfOpenProbes, else probes++.
	//  Remember gen := b.generation and return func(err error) { b.record(gen, err) }.
	panic("TODO(day 47): Breaker.Allow")
}

// Execute runs fn if the breaker allows it and records the outcome.
func (b *Breaker) Execute(fn func() error) (err error) {
	done, err := b.Allow()
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			done(fmt.Errorf("panic: %v", r))
			panic(r)
		}
	}()
	err = fn()
	done(err)
	return err
}

// current moves an open breaker to half-open once its timeout has passed. Lock held.
func (b *Breaker) current() State {
	// TODO(day 47): if open and Now() is not before openedAt+OpenTimeout, setState(HalfOpen).
	// Return b.state.
	panic("TODO(day 47): Breaker.current")
}

// record applies one call's outcome. Lock not held.
func (b *Breaker) record(gen uint64, err error) {
	// TODO(day 47): lock; state := b.current(); ignore a result whose gen is not b.generation.
	//  failed := b.s.IsFailure(err).
	//  Closed: success resets failures; a failure increments it and opens at FailureThreshold.
	//  HalfOpen: a failure opens; a non-failure error (canceled) gives the probe slot back
	//   (probes--); a success counts, and HalfOpenProbes successes close the breaker.
	panic("TODO(day 47): Breaker.record")
}

// setState starts a new generation with fresh counters. Lock held.
func (b *Breaker) setState(to State) {
	// TODO(day 47): remember from, set state, bump generation, zero the three counters,
	//  set openedAt = Now() when opening, and call OnStateChange when the state really changed.
	panic("TODO(day 47): Breaker.setState")
}
