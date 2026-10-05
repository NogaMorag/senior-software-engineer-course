package catalog

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ProductCache is cache-aside written out by hand (day 31). Read path: try Redis; on
// a miss read PostgreSQL and fill Redis with a TTL. Write path: commit, then delete
// the key. Concurrent misses are not collapsed: that is day 32.
type ProductCache struct {
	rdb    redis.UniversalClient
	store  *Store
	prefix string
	ttl    time.Duration
	jitter time.Duration
}

// NewProductCache returns a cache over store. Keys are prefix + "product:" + id.
func NewProductCache(rdb redis.UniversalClient, store *Store, prefix string, ttl, jitter time.Duration) *ProductCache {
	return &ProductCache{rdb: rdb, store: store, prefix: prefix, ttl: ttl, jitter: jitter}
}

// ProductKey is the cache key of one product.
func ProductKey(prefix string, id int64) string { return fmt.Sprintf("%sproduct:%d", prefix, id) }

// Get returns the product, from Redis when it is there.
func (c *ProductCache) Get(ctx context.Context, id int64) (Product, error) {
	// TODO(day 31): the read path of cache-aside.
	//  1. GET the key ProductKey(c.prefix, id); decode the JSON on a hit.
	//  2. redis.Nil is a miss; any other error means Redis is unhealthy: serve from c.store.
	//  3. On a miss read c.store.GetProduct; do not cache an error (ErrNotFound included).
	//  4. SET the JSON with the TTL cache.Jittered(c.ttl, c.jitter); ignore a failed SET.
	panic("TODO(day 31): ProductCache.Get")
}

// UpdatePrice writes the new price and then invalidates the cached product.
func (c *ProductCache) UpdatePrice(ctx context.Context, id, priceCents int64) error {
	// TODO(day 31): the write path. Commit first (c.store.UpdatePrice), then DEL the key.
	// Wrap a failed DEL in an error that says the price itself was updated.
	panic("TODO(day 31): ProductCache.UpdatePrice")
}
