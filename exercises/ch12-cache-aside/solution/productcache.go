package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"example.com/ch12-cache-aside/internal/cache"
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
	key := ProductKey(c.prefix, id)
	raw, err := c.rdb.Get(ctx, key).Bytes()
	switch {
	case err == nil:
		var p Product
		if json.Unmarshal(raw, &p) == nil {
			return p, nil // hit
		}
		// An entry we cannot decode is treated as a miss and overwritten below.
	case errors.Is(err, redis.Nil):
		// miss
	default:
		// Redis is down or slow: serve from the database and do not try to fill.
		return c.store.GetProduct(ctx, id)
	}

	p, err := c.store.GetProduct(ctx, id)
	if err != nil {
		return Product{}, err // not-found is not cached: see section 12.6
	}
	if raw, err := json.Marshal(p); err == nil {
		// A failed fill is not a failed read: the next reader simply misses again.
		_ = c.rdb.Set(ctx, key, raw, cache.Jittered(c.ttl, c.jitter)).Err()
	}
	return p, nil
}

// UpdatePrice writes the new price and then invalidates the cached product.
func (c *ProductCache) UpdatePrice(ctx context.Context, id, priceCents int64) error {
	if err := c.store.UpdatePrice(ctx, id, priceCents); err != nil {
		return err
	}
	// After the commit, never before: a delete inside the transaction lets a reader
	// refill the old row before the commit makes the new one visible.
	if err := c.rdb.Del(ctx, ProductKey(c.prefix, id)).Err(); err != nil {
		return fmt.Errorf("price of %d updated, but invalidation failed (the TTL will expire it): %w", id, err)
	}
	return nil
}
