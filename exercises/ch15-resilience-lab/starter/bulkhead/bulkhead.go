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
	// TODO(day 47):
	//  1. Try a non-blocking send on b.slots (select with default); success means a slot.
	//  2. With maxWait <= 0 return ErrFull.
	//  3. Otherwise wait in a select on: the send, a timer of maxWait (ErrFull), ctx.Done() (ctx.Err()).
	//     Stop the timer with defer.
	panic("TODO(day 47): Bulkhead.Acquire")
}

// Release gives a slot back.
func (b *Bulkhead) Release() {
	// TODO(day 47): give a slot back by receiving from b.slots.
	panic("TODO(day 47): Bulkhead.Release")
}

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
