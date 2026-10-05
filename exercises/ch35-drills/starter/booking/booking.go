// Package booking is drill 5: book a resource (a room, a desk, a test rig) for a time
// range, and never let two bookings of the same resource overlap. Two stores implement
// the same contract: Memory, with a lock per resource, and Postgres, with an exclusion
// constraint.
// Time box: 35 minutes for both. The tests in booking_test.go are the specification.
package booking

import (
	"context"
	"errors"
	"time"
)

// Booking is one reservation of Resource for the half-open range [Start, End): a booking
// that ends at 10:00 and one that starts at 10:00 do not overlap.
type Booking struct {
	ID       int64
	Resource string
	Who      string
	Start    time.Time
	End      time.Time
}

var (
	ErrConflict = errors.New("booking: overlaps an existing booking")
	ErrInvalid  = errors.New("booking: invalid request")
	ErrNotFound = errors.New("booking: no such booking")
)

// Store is the contract both implementations meet.
type Store interface {
	// Book reserves resource for [start, end) or returns ErrConflict or ErrInvalid.
	Book(ctx context.Context, resource, who string, start, end time.Time) (Booking, error)
	// Cancel removes a booking, freeing its range.
	Cancel(ctx context.Context, id int64) error
	// List returns the bookings of resource ordered by start.
	List(ctx context.Context, resource string) ([]Booking, error)
}
