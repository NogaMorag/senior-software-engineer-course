// Package bulkhead limits how many calls to one dependency run at the same time, so a slow
// dependency can hold at most its own slots and never every worker of the process.
package bulkhead

import (
	"context"
	"errors"
	"time"
)

// ErrFull is returned when no slot became free within the allowed wait.
var ErrFull = errors.New("bulkhead: full")

// Bulkhead is a counting semaphore built on a buffered channel: a send takes a slot,
// a receive gives it back.
type Bulkhead struct {
	slots   chan struct{}
	maxWait time.Duration
}

// New returns a bulkhead with limit slots. A caller waits at most maxWait for a slot;
// zero means fail immediately when full.
func New(limit int, maxWait time.Duration) *Bulkhead {
	if limit < 1 {
		limit = 1
	}
	return &Bulkhead{slots: make(chan struct{}, limit), maxWait: maxWait}
}

// Acquire takes a slot, waiting at most maxWait and never past ctx. Every successful
// Acquire must be followed by exactly one Release.
func (b *Bulkhead) Acquire(ctx context.Context) error {
	select {
	case b.slots <- struct{}{}:
		return nil
	default:
	}
	if b.maxWait <= 0 {
		return ErrFull
	}
	timer := time.NewTimer(b.maxWait)
	defer timer.Stop()
	select {
	case b.slots <- struct{}{}:
		return nil
	case <-timer.C:
		return ErrFull
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release gives a slot back.
func (b *Bulkhead) Release() { <-b.slots }

// Do runs fn inside a slot.
func (b *Bulkhead) Do(ctx context.Context, fn func(context.Context) error) error {
	if err := b.Acquire(ctx); err != nil {
		return err
	}
	defer b.Release()
	return fn(ctx)
}

// InUse reports how many slots are taken right now.
func (b *Bulkhead) InUse() int { return len(b.slots) }

// Limit reports the number of slots.
func (b *Bulkhead) Limit() int { return cap(b.slots) }
