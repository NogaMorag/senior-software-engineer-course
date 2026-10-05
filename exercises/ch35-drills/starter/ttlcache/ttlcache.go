// Package ttlcache is drill 3: a bounded in-memory cache where every entry has a time to
// live. Expired entries are removed lazily (on access) and actively (by a sweep); when the
// cache is full, the least recently used entry is evicted.
// Time box: 25 minutes. The tests in ttlcache_test.go are the specification.
package ttlcache

import (
	"context"
	"errors"
	"time"

	"example.com/ch35-drills/clock"
)

// ErrConfig is returned by New for a capacity or TTL that cannot work.
var ErrConfig = errors.New("ttlcache: capacity must be >= 1 and ttl > 0")

// Cache must be safe for concurrent use. The zero value is not usable; call New.
type Cache[K comparable, V any] struct {
	// TODO: your fields. Hint: the standard library has container/list.
}

// New returns a cache that holds at most capacity entries, each for ttl unless Set with
// its own TTL. now is the clock.
func New[K comparable, V any](capacity int, ttl time.Duration, now func() time.Time) (*Cache[K, V], error) {
	panic("TODO: drill 3, see chapter 35, section 35.4")
}

// Get returns the value for key if it is present and not expired. A hit makes the entry
// the most recently used; an expired entry is removed on the spot.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	panic("TODO: drill 3, see chapter 35, section 35.4")
}

// Set stores value under key for the default TTL.
func (c *Cache[K, V]) Set(key K, value V) {
	panic("TODO: drill 3, see chapter 35, section 35.4")
}

// SetWithTTL stores value under key for ttl. Setting an existing key replaces its value
// and its expiry and makes it the most recently used. A ttl <= 0 deletes the key.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	panic("TODO: drill 3, see chapter 35, section 35.4")
}

// Delete removes key and reports whether it was present (expired or not).
func (c *Cache[K, V]) Delete(key K) bool {
	panic("TODO: drill 3, see chapter 35, section 35.4")
}

// Len reports how many entries are stored, including expired ones not yet removed.
func (c *Cache[K, V]) Len() int {
	panic("TODO: drill 3, see chapter 35, section 35.4")
}

// DeleteExpired removes every expired entry and returns how many it removed.
func (c *Cache[K, V]) DeleteExpired() int {
	panic("TODO: drill 3, see chapter 35, section 35.4")
}

// RunJanitor calls DeleteExpired every interval until ctx is cancelled.
func (c *Cache[K, V]) RunJanitor(ctx context.Context, clk clock.Clock, every time.Duration) {
	panic("TODO: drill 3, see chapter 35, section 35.4")
}
