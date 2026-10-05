// Package cache is a small generic cache-aside helper: read through a key-value
// store, load from the source of truth on a miss, collapse concurrent misses for
// one key with singleflight, and count hits and misses. Values are stored as JSON.
package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

// ErrMiss is what a Store returns from Get when the key is absent.
var ErrMiss = errors.New("cache miss")

// Store is the little the helper needs from a key-value server.
type Store interface {
	Get(ctx context.Context, key string) ([]byte, error) // ErrMiss when absent
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// Options configures a Cache.
type Options struct {
	TTL    time.Duration // base time to live of every entry
	Jitter time.Duration // up to this much is added at random to each TTL
}

// Stats are counts since the cache was created.
type Stats struct {
	Hits   int64 // served from the store
	Misses int64 // not in the store (or the store failed)
	Loads  int64 // calls of the load function, after collapsing
	Errors int64 // store reads that failed with something other than a miss
}

// HitRatio is hits divided by all lookups, or 0 before the first lookup.
func (s Stats) HitRatio() float64 {
	if total := s.Hits + s.Misses; total > 0 {
		return float64(s.Hits) / float64(total)
	}
	return 0
}

// Cache is cache-aside over a Store for values of type T. It is safe for
// concurrent use; create one per kind of value and share it.
type Cache[T any] struct {
	store                        Store
	opts                         Options
	group                        singleflight.Group
	hits, misses, loads, errored atomic.Int64
}

// New returns a cache over store.
func New[T any](store Store, opts Options) *Cache[T] {
	return &Cache[T]{store: store, opts: opts}
}

// Get returns the value under key. On a miss it calls load, stores the result with
// a jittered TTL and returns it; concurrent misses for the same key in this process
// share one call of load. Errors from load are returned and never cached. When the
// store itself fails, Get still answers from load, so the cache is never the reason
// a read fails.
func (c *Cache[T]) Get(ctx context.Context, key string, load func(context.Context) (T, error)) (T, error) {
	// TODO(day 32):
	//  1. c.store.Get; on success decode the JSON, count a hit and return it.
	//     ErrMiss is a miss; any other error counts in c.errored and is also a miss.
	//  2. Count the miss, then call c.group.Do(key, ...): inside, count a load, call
	//     load, and on success Set the JSON with Jittered(c.opts.TTL, c.opts.Jitter)
	//     before returning, ignoring a failed Set. Never cache an error.
	//  3. Convert the shared result back with v.(T).
	panic("TODO(day 32): Cache.Get")
}

// Invalidate deletes the entry under key. Call it after the write is committed.
func (c *Cache[T]) Invalidate(ctx context.Context, key string) error {
	return c.store.Del(ctx, key)
}

// Stats returns the counts so far.
func (c *Cache[T]) Stats() Stats {
	return Stats{Hits: c.hits.Load(), Misses: c.misses.Load(), Loads: c.loads.Load(), Errors: c.errored.Load()}
}
