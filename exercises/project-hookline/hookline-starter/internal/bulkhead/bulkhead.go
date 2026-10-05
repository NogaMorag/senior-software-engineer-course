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
	panic("TODO: see chapter 38")

}

// Acquire takes a slot, waiting at most maxWait and never past ctx. Every successful
// Acquire must be followed by exactly one Release.
func (b *Bulkhead) Acquire(ctx context.Context) error {
	panic("TODO: see chapter 38")

}

// Release gives a slot back.
func (b *Bulkhead) Release() { panic("TODO: see chapter 38") }

// Do runs fn inside a slot.
func (b *Bulkhead) Do(ctx context.Context, fn func(context.Context) error) error {
	panic("TODO: see chapter 38")

}

// InUse reports how many slots are taken right now.
func (b *Bulkhead) InUse() int { panic("TODO: see chapter 38") }

// Limit reports the number of slots.
func (b *Bulkhead) Limit() int { panic("TODO: see chapter 38") }
