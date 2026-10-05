// Package ttlcache is drill 3: a bounded in-memory cache where every entry has a time to
// live. Expired entries are removed lazily (on access) and actively (by a sweep); when the
// cache is full, the least recently used entry is evicted.
package ttlcache

import (
	"container/list"
	"context"
	"errors"
	"sync"
	"time"

	"example.com/ch35-drills/clock"
)

// ErrConfig is returned by New for a capacity or TTL that cannot work.
var ErrConfig = errors.New("ttlcache: capacity must be >= 1 and ttl > 0")

// Cache is safe for concurrent use. The zero value is not usable; call New.
type Cache[K comparable, V any] struct {
	capacity int
	ttl      time.Duration
	now      func() time.Time

	mu    sync.Mutex
	items map[K]*list.Element // values are *entry[K, V]
	order *list.List          // front: most recently used
}

type entry[K comparable, V any] struct {
	key     K
	value   V
	expires time.Time
}

// New returns a cache that holds at most capacity entries, each for ttl unless Set with
// its own TTL. now is the clock.
func New[K comparable, V any](capacity int, ttl time.Duration, now func() time.Time) (*Cache[K, V], error) {
	if capacity < 1 || ttl <= 0 || now == nil {
		return nil, ErrConfig
	}
	return &Cache[K, V]{capacity: capacity, ttl: ttl, now: now, items: make(map[K]*list.Element), order: list.New()}, nil
}

// Get returns the value for key if it is present and not expired. A hit makes the entry
// the most recently used; an expired entry is removed on the spot.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	// A full Mutex, not an RWMutex: even a hit writes (it moves the entry to the front).
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	e := el.Value.(*entry[K, V])
	if !c.now().Before(e.expires) {
		c.remove(el)
		var zero V
		return zero, false
	}
	c.order.MoveToFront(el)
	return e.value, true
}

// Set stores value under key for the default TTL.
func (c *Cache[K, V]) Set(key K, value V) { c.SetWithTTL(key, value, c.ttl) }

// SetWithTTL stores value under key for ttl. Setting an existing key replaces its value
// and its expiry and makes it the most recently used. A ttl <= 0 deletes the key.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ttl <= 0 {
		if el, ok := c.items[key]; ok {
			c.remove(el)
		}
		return
	}
	expires := c.now().Add(ttl)
	if el, ok := c.items[key]; ok {
		e := el.Value.(*entry[K, V])
		e.value, e.expires = value, expires
		c.order.MoveToFront(el)
		return
	}
	c.items[key] = c.order.PushFront(&entry[K, V]{key: key, value: value, expires: expires})
	if c.order.Len() > c.capacity {
		c.remove(c.order.Back())
	}
}

// Delete removes key and reports whether it was present (expired or not).
func (c *Cache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if ok {
		c.remove(el)
	}
	return ok
}

// Len reports how many entries are stored, including expired ones not yet removed.
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}

// DeleteExpired removes every expired entry and returns how many it removed. It walks the
// whole cache while holding the lock: O(n). That is acceptable for a cache of thousands of
// entries swept every few seconds; for millions, keep a second index ordered by expiry.
func (c *Cache[K, V]) DeleteExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	n := 0
	for el := c.order.Front(); el != nil; {
		next := el.Next() // read before remove, which clears the links
		if !now.Before(el.Value.(*entry[K, V]).expires) {
			c.remove(el)
			n++
		}
		el = next
	}
	return n
}

// RunJanitor calls DeleteExpired every interval until ctx is cancelled. Without it,
// entries that are never read again stay in memory until LRU pushes them out.
func (c *Cache[K, V]) RunJanitor(ctx context.Context, clk clock.Clock, every time.Duration) {
	for {
		t := clk.NewTimer(every)
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C():
			c.DeleteExpired()
		}
	}
}

// remove deletes el from both indexes. Called with c.mu held.
func (c *Cache[K, V]) remove(el *list.Element) {
	e := c.order.Remove(el).(*entry[K, V])
	delete(c.items, e.key)
}
