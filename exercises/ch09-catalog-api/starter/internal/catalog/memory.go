// Package catalog holds the catalog's domain type and its in-memory store.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"example.com/ch09-catalog-api/internal/page"
)

// Item is one catalog entry.
type Item struct {
	ID         string    `json:"id"`
	SKU        string    `json:"sku"`
	Name       string    `json:"name"`
	PriceCents int64     `json:"price_cents"`
	CreatedAt  time.Time `json:"created_at"`
}

// Errors the store returns; handlers map them to problem types.
var (
	ErrNotFound     = errors.New("item not found")
	ErrDuplicateSKU = errors.New("sku already exists")
)

// Store is what the handlers need. The memory implementation below satisfies it; a PostgreSQL one
// can replace it without touching a handler.
type Store interface {
	Create(ctx context.Context, sku, name string, priceCents int64) (Item, error)
	Get(ctx context.Context, id string) (Item, error)
	List(ctx context.Context, after *page.Cursor, limit int) (items []Item, more bool, err error)
}

// Memory keeps items sorted newest first by (CreatedAt, ID). Safe for concurrent use.
type Memory struct {
	mu    sync.RWMutex
	items []Item // sorted: CreatedAt descending, then ID descending
	byID  map[string]int
	skus  map[string]bool
	seq   int
	now   func() time.Time
}

// NewMemory returns an empty store. now is injected so tests control CreatedAt (nil means time.Now).
func NewMemory(now func() time.Time) *Memory {
	if now == nil {
		now = time.Now
	}
	return &Memory{byID: map[string]int{}, skus: map[string]bool{}, now: now}
}

// newer reports whether a sorts before b: later CreatedAt first, then larger ID first.
func newer(a, b Item) bool {
	if !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.After(b.CreatedAt)
	}
	return a.ID > b.ID
}

// Create adds an item. IDs are zero-padded so that string order equals creation order.
func (m *Memory) Create(_ context.Context, sku, name string, priceCents int64) (Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.skus[sku] {
		return Item{}, ErrDuplicateSKU
	}
	m.seq++
	it := Item{ID: fmt.Sprintf("itm_%08d", m.seq), SKU: sku, Name: name, PriceCents: priceCents, CreatedAt: m.now().UTC()}
	at, _ := slices.BinarySearchFunc(m.items, it, func(e, t Item) int {
		if newer(e, t) {
			return -1
		}
		return 1
	})
	m.items = slices.Insert(m.items, at, it)
	for i := at; i < len(m.items); i++ {
		m.byID[m.items[i].ID] = i
	}
	m.skus[sku] = true
	return it, nil
}

// Get returns one item or ErrNotFound.
func (m *Memory) Get(_ context.Context, id string) (Item, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	i, ok := m.byID[id]
	if !ok {
		return Item{}, ErrNotFound
	}
	return m.items[i], nil
}

// List returns up to limit items that sort strictly after the cursor (all items when after is nil),
// and whether more exist. In SQL the same query is:
//
//	SELECT ... FROM items WHERE (created_at, id) < ($1, $2)
//	ORDER BY created_at DESC, id DESC LIMIT $3 + 1
func (m *Memory) List(_ context.Context, after *page.Cursor, limit int) ([]Item, bool, error) {
	// TODO(day 20): under the read lock, find the first item that sorts strictly after the cursor
	// (slices.BinarySearchFunc with newer), take up to limit items, and report whether more remain.
	// Return a copy of the slice, never the store's own backing array.
	panic("TODO(day 20): Memory.List")
}

// ListOffset is the naive alternative, kept only so a test can show how it breaks under inserts.
func (m *Memory) ListOffset(_ context.Context, offset, limit int) []Item {
	m.mu.RLock()
	defer m.mu.RUnlock()
	start := min(offset, len(m.items))
	end := min(start+limit, len(m.items))
	return slices.Clone(m.items[start:end])
}
