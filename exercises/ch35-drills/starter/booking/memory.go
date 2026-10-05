package booking

import (
	"context"
	"time"
)

// Memory keeps bookings in this process, with one lock per resource, so that bookings of
// different resources never wait for each other.
type Memory struct {
	// TODO: your fields.
}

// NewMemory returns an empty store.
func NewMemory() *Memory {
	panic("TODO: drill 5, see chapter 35, section 35.6")
}

// Book reserves resource for [start, end) or returns ErrConflict or ErrInvalid.
func (m *Memory) Book(ctx context.Context, resource, who string, start, end time.Time) (Booking, error) {
	panic("TODO: drill 5, see chapter 35, section 35.6")
}

// Cancel removes a booking.
func (m *Memory) Cancel(ctx context.Context, id int64) error {
	panic("TODO: drill 5, see chapter 35, section 35.6")
}

// List returns the bookings of resource ordered by start.
func (m *Memory) List(ctx context.Context, resource string) ([]Booking, error) {
	panic("TODO: drill 5, see chapter 35, section 35.6")
}
