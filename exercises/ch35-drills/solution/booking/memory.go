package booking

import (
	"context"
	"slices"
	"sync"
	"time"
)

// Memory keeps bookings in this process. Each resource has its own lock, so bookings of
// different rooms never wait for each other; the map of resources has a lock of its own
// that is held only long enough to find or create a resource.
type Memory struct {
	mu        sync.Mutex
	resources map[string]*resource
	byID      map[int64]string // booking id -> resource, for Cancel
	nextID    int64
}

type resource struct {
	mu       sync.Mutex
	bookings []Booking // sorted by Start
}

// NewMemory returns an empty store.
func NewMemory() *Memory {
	return &Memory{resources: make(map[string]*resource), byID: make(map[int64]string)}
}

func (m *Memory) resource(name string) *resource {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.resources[name]
	if !ok {
		r = &resource{}
		m.resources[name] = r
	}
	return r
}

// Book checks for an overlap and inserts while holding the resource's lock: the check and
// the insert are one step, so two callers cannot both see a free slot.
func (m *Memory) Book(_ context.Context, name, who string, start, end time.Time) (Booking, error) {
	if err := validate(name, who, start, end); err != nil {
		return Booking{}, err
	}
	r := m.resource(name)
	r.mu.Lock()
	defer r.mu.Unlock()
	// Bookings are sorted and never overlap, so their ends are sorted too: the first
	// booking that ends after start is the only one that can collide.
	i, _ := slices.BinarySearchFunc(r.bookings, start, func(b Booking, t time.Time) int {
		if b.End.After(t) {
			return 1
		}
		return -1
	})
	if i < len(r.bookings) && overlaps(r.bookings[i].Start, r.bookings[i].End, start, end) {
		return Booking{}, ErrConflict
	}
	m.mu.Lock() // lock order: resource, then store. Cancel takes them in the same order.
	m.nextID++
	b := Booking{ID: m.nextID, Resource: name, Who: who, Start: start, End: end}
	m.byID[b.ID] = name
	m.mu.Unlock()
	r.bookings = slices.Insert(r.bookings, i, b)
	return b, nil
}

// Cancel removes a booking.
func (m *Memory) Cancel(_ context.Context, id int64) error {
	m.mu.Lock()
	name, ok := m.byID[id]
	r := m.resources[name]
	m.mu.Unlock()
	if !ok {
		return ErrNotFound
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	i := slices.IndexFunc(r.bookings, func(b Booking) bool { return b.ID == id })
	if i < 0 {
		return ErrNotFound // cancelled concurrently
	}
	r.bookings = slices.Delete(r.bookings, i, i+1)
	m.mu.Lock()
	delete(m.byID, id)
	m.mu.Unlock()
	return nil
}

// List returns a copy, so the caller cannot change the store's slice.
func (m *Memory) List(_ context.Context, name string) ([]Booking, error) {
	m.mu.Lock()
	r, ok := m.resources[name] // a read must not create a resource
	m.mu.Unlock()
	if !ok {
		return nil, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.bookings), nil
}
