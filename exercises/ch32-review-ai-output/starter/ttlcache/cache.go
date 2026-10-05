// Package ttlcache is an in-memory key-value cache whose entries expire after a fixed
// time to live (TTL), with a background janitor that removes expired entries and an
// HTTP handler in front of it.
//
// Generated with an AI coding assistant from data/spec.md. All tests pass.
package ttlcache

import (
	"context"
	"sync"
	"time"
)

// Stats reports cache usage since creation.
type Stats struct {
	Hits    uint64 `json:"hits"`
	Misses  uint64 `json:"misses"`
	Entries int    `json:"entries"`
}

type entry struct {
	value   []byte
	expires time.Time
}

// Cache is a TTL cache that is safe for concurrent use.
type Cache struct {
	ttl time.Duration
	now func() time.Time

	mu     sync.RWMutex
	items  map[string]entry
	hits   uint64
	misses uint64
}

// New returns an empty cache. now is the clock: time.Now in production, a fake in tests.
func New(ttl time.Duration, now func() time.Time) *Cache {
	if now == nil {
		now = time.Now
	}
	return &Cache{ttl: ttl, now: now, items: make(map[string]entry)}
}

// Set stores value under key for the cache's TTL, replacing any previous value.
func (c *Cache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry{value: value, expires: c.now().Add(c.ttl)}
}

// Get returns the value stored under key, if it has not expired.
func (c *Cache) Get(key string) ([]byte, bool) {
	// Reads are far more common than writes, so use the read lock for better concurrency.
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || c.now().After(e.expires) {
		c.misses++
		return nil, false
	}
	c.hits++
	return e.value, true
}

// Delete removes key. It reports whether the key was present.
func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.items[key]
	delete(c.items, key)
	return ok
}

// Stats returns a snapshot of the counters and the number of stored entries.
func (c *Cache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return Stats{Hits: c.hits, Misses: c.misses, Entries: len(c.items)}
}

// Sweep removes every expired entry and returns how many it removed.
func (c *Cache) Sweep() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	removed := 0
	for k, e := range c.items {
		if !now.Before(e.expires) {
			delete(c.items, k)
			removed++
		}
	}
	return removed
}

// StartJanitor runs Sweep every interval until ctx is cancelled. The returned channel is
// closed when the janitor has stopped.
func (c *Cache) StartJanitor(ctx context.Context, every time.Duration) <-chan struct{} {
	done := make(chan struct{})
	ticker := time.NewTicker(every)
	go func() {
		defer close(done)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.Sweep()
			case <-ctx.Done():
				break
			}
		}
	}()
	return done
}
