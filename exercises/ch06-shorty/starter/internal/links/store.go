// Package links is the URL shortener itself: the Link type, the Store
// interface with an in-memory implementation, and the HTTP handlers.
package links

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Link is one short code and the URL it points to.
type Link struct {
	Code      string
	URL       string
	Hits      int64 // how many times the code was resolved
	CreatedAt time.Time
}

var (
	ErrNotFound = errors.New("link not found")
	ErrExists   = errors.New("code already in use")
)

// Store keeps links. The memory implementation below is the only one in this
// chapter; a PostgreSQL one can satisfy the same interface later.
type Store interface {
	// Create stores l; ErrExists if l.Code is taken.
	Create(ctx context.Context, l Link) error
	// Resolve returns the link for code and counts one hit; ErrNotFound if absent.
	Resolve(ctx context.Context, code string) (Link, error)
	// Get returns the link for code without counting a hit; ErrNotFound if absent.
	Get(ctx context.Context, code string) (Link, error)
}

// MemoryStore is a Store in a map. It is safe for concurrent use.
type MemoryStore struct {
	mu    sync.RWMutex
	links map[string]*entry
}

type entry struct {
	link Link         // never changes after Create
	hits atomic.Int64 // changes on every Resolve, without the write lock
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{links: make(map[string]*entry)}
}

func (s *MemoryStore) Create(_ context.Context, l Link) error {
	// TODO 1a (day 11): take the write lock; ErrExists if l.Code is present;
	// otherwise store an *entry whose hits start at l.Hits.
	panic("TODO 1a: MemoryStore.Create")
}

func (s *MemoryStore) Resolve(_ context.Context, code string) (Link, error) {
	// TODO 1b (day 11): look the code up under the read lock (RLock), release it,
	// ErrNotFound if absent; otherwise copy e.link and set Hits = e.hits.Add(1).
	panic("TODO 1b: MemoryStore.Resolve")
}

func (s *MemoryStore) Get(_ context.Context, code string) (Link, error) {
	// TODO 1c (day 11): like Resolve, but Hits = e.hits.Load(): no hit is counted.
	panic("TODO 1c: MemoryStore.Get")
}
