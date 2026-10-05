// Package degrade keeps answering when a dependency fails: with the last good value for a
// bounded time, and with a fallback when a feature is switched off or nothing is cached.
package degrade

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// ErrNoValue means the fetch failed and there was no value fresh enough to fall back to.
var ErrNoValue = errors.New("degrade: fetch failed and no usable stale value")

// Stale wraps a fetch with the last good value. It serves that value, marked stale, for
// at most MaxStale after it was fetched. Stale is safe for concurrent use.
type Stale[T any] struct {
	MaxStale time.Duration
	Now      func() time.Time // default time.Now

	mu   sync.Mutex
	last T
	at   time.Time
	ok   bool
}

// Get calls fetch. On success it remembers the value; on failure it returns the last
// good value with stale set, or ErrNoValue joined with the fetch error.
func (s *Stale[T]) Get(ctx context.Context, fetch func(context.Context) (T, error)) (v T, stale bool, err error) {
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	v, err = fetch(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err == nil {
		s.last, s.at, s.ok = v, now(), true
		return v, false, nil
	}
	if s.ok && now().Sub(s.at) <= s.MaxStale {
		return s.last, true, nil
	}
	var zero T
	return zero, false, errors.Join(ErrNoValue, err)
}

// Feature is a kill switch for an optional part of a response.
type Feature struct{ off atomic.Bool }

// Enabled reports whether the feature is on (the zero Feature is on).
func (f *Feature) Enabled() bool { return !f.off.Load() }

// Set turns the feature on or off at run time.
func (f *Feature) Set(on bool) { f.off.Store(!on) }

// Optional calls fetch only when the feature is on, and returns fallback when it is off
// or when fetch fails: the optional part degrades, the response still goes out.
func Optional[T any](ctx context.Context, f *Feature, fallback T, fetch func(context.Context) (T, error)) (T, bool) {
	if !f.Enabled() {
		return fallback, false
	}
	v, err := fetch(ctx)
	if err != nil {
		return fallback, false
	}
	return v, true
}
