// Package clock is the time source the drills take as a parameter, so that a test can
// move time forward by an hour without waiting an hour. It is support code: complete in
// the starter too.
package clock

import (
	"sort"
	"sync"
	"time"
)

// Clock is what code needs from time: the current instant and a timer.
type Clock interface {
	Now() time.Time
	NewTimer(d time.Duration) Timer
}

// Timer is the part of *time.Timer the drills use.
type Timer interface {
	C() <-chan time.Time
	Stop() bool
}

// Real returns the system clock.
func Real() Clock { return realClock{} }

type realClock struct{}

func (realClock) Now() time.Time                 { return time.Now() }
func (realClock) NewTimer(d time.Duration) Timer { return realTimer{time.NewTimer(d)} }

type realTimer struct{ t *time.Timer }

func (r realTimer) C() <-chan time.Time { return r.t.C }
func (r realTimer) Stop() bool          { return r.t.Stop() }

// Fake is a clock that moves only when a test calls Advance. It is safe for concurrent use.
type Fake struct {
	mu      sync.Mutex
	now     time.Time
	timers  []*fakeTimer
	changed *sync.Cond
}

// NewFake returns a fake clock that reads start.
func NewFake(start time.Time) *Fake {
	f := &Fake{now: start}
	f.changed = sync.NewCond(&f.mu)
	return f
}

// Now returns the fake time.
func (f *Fake) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

// NewTimer returns a timer that fires when the fake time reaches now+d. A timer with
// d <= 0 has already fired.
func (f *Fake) NewTimer(d time.Duration) Timer {
	f.mu.Lock()
	defer f.mu.Unlock()
	t := &fakeTimer{f: f, at: f.now.Add(d), c: make(chan time.Time, 1)}
	if d <= 0 {
		t.c <- f.now
		return t
	}
	f.timers = append(f.timers, t)
	f.changed.Broadcast()
	return t
}

// Advance moves the time forward by d and fires, in order, every timer that is due.
func (f *Fake) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
	sort.SliceStable(f.timers, func(i, j int) bool { return f.timers[i].at.Before(f.timers[j].at) })
	keep := f.timers[:0]
	for _, t := range f.timers {
		if t.at.After(f.now) {
			keep = append(keep, t)
			continue
		}
		t.c <- t.at // buffered and sent at most once: never blocks
	}
	f.timers = keep
	f.changed.Broadcast()
}

// Waiters reports how many timers are waiting to fire.
func (f *Fake) Waiters() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.timers)
}

// BlockUntil waits until at least n timers are waiting. Tests call it before Advance when
// the code under test sets a relative timer (a ticker, a janitor), so that the timer
// exists before time moves.
func (f *Fake) BlockUntil(n int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for len(f.timers) < n {
		f.changed.Wait()
	}
}

type fakeTimer struct {
	f  *Fake
	at time.Time
	c  chan time.Time
}

func (t *fakeTimer) C() <-chan time.Time { return t.c }

// Stop removes the timer. It reports whether the timer was still waiting.
func (t *fakeTimer) Stop() bool {
	t.f.mu.Lock()
	defer t.f.mu.Unlock()
	for i, other := range t.f.timers {
		if other == t {
			t.f.timers = append(t.f.timers[:i], t.f.timers[i+1:]...)
			t.f.changed.Broadcast()
			return true
		}
	}
	return false
}
