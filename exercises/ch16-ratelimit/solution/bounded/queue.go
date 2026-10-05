// Package bounded is a queue with a fixed capacity and two ways to say "full": Put blocks
// the producer (back-pressure), TryPut refuses at once (rejection).
package bounded

import (
	"context"
	"errors"
)

// ErrFull is TryPut's refusal.
var ErrFull = errors.New("bounded: queue full")

// Queue is a buffered channel with a name for each policy. Safe for concurrent use.
type Queue[T any] struct{ ch chan T }

// New returns a queue that holds at most size items.
func New[T any](size int) *Queue[T] { return &Queue[T]{ch: make(chan T, size)} }

// Put waits for room: a full queue slows the producer down to the consumer's pace.
// It gives up when ctx is done.
func (q *Queue[T]) Put(ctx context.Context, v T) error {
	select {
	case q.ch <- v:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryPut adds v if there is room and returns ErrFull otherwise, without waiting.
func (q *Queue[T]) TryPut(v T) error {
	select {
	case q.ch <- v:
		return nil
	default:
		return ErrFull
	}
}

// Get waits for an item; ok is false when ctx is done first.
func (q *Queue[T]) Get(ctx context.Context) (v T, ok bool) {
	select {
	case v = <-q.ch:
		return v, true
	case <-ctx.Done():
		return v, false
	}
}

// Len is the number of items waiting; Cap is the capacity.
func (q *Queue[T]) Len() int { return len(q.ch) }
func (q *Queue[T]) Cap() int { return cap(q.ch) }
