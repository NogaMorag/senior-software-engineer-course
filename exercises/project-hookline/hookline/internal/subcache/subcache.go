// Package subcache answers the dispatcher's question "which endpoints get this event?" from a cache
// (chapter 12): cache-aside over Redis, one entry per tenant holding all of its active subscriptions,
// with singleflight against stampedes and a jittered TTL. Writers (the API) invalidate the tenant's
// entry after every change; if an invalidation is lost, the TTL bounds how long the entry stays stale.
package subcache

import (
	"context"
	"slices"
	"time"

	"example.com/hookline/internal/cache"
	"example.com/hookline/internal/domain"
)

// Defaults for the entry lifetime. A stale entry can make the dispatcher miss a new subscription or
// still deliver to a removed one for at most TTL+Jitter after a lost invalidation.
const (
	DefaultTTL    = time.Minute
	DefaultJitter = 10 * time.Second
)

// Loader reads the source of truth: store.Subscriptions satisfies it.
type Loader interface {
	ActiveSubscriptions(ctx context.Context, tenantID string) ([]domain.Subscription, error)
}

// Cache is safe for concurrent use.
type Cache struct {
	c      *cache.Cache[[]domain.Subscription]
	loader Loader
}

// New returns a cache over store (a cache.RedisStore, or Disabled() when there is no Redis).
func New(store cache.Store, loader Loader, opts cache.Options) *Cache {
	return &Cache{c: cache.New[[]domain.Subscription](store, opts), loader: loader}
}

// Key is the cache key of a tenant. The version lets a new value format coexist with old entries.
func Key(tenantID string) string { return "hookline:subs:v1:" + tenantID }

// Endpoints returns the ids of the enabled endpoints with a subscription to eventType or to "*",
// each once and sorted, even when an endpoint has both.
func (c *Cache) Endpoints(ctx context.Context, tenantID, eventType string) ([]string, error) {
	subs, err := c.c.Get(ctx, Key(tenantID), func(ctx context.Context) ([]domain.Subscription, error) {
		return c.loader.ActiveSubscriptions(ctx, tenantID)
	})
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, s := range subs {
		if s.EventType == eventType || s.EventType == domain.AllTypes {
			ids = append(ids, s.EndpointID)
		}
	}
	slices.Sort(ids)
	return slices.Compact(ids), nil
}

// Invalidate drops the tenant's entry. Call it after the change is committed, never before: a reader
// between an early delete and the commit would put the old value back.
func (c *Cache) Invalidate(ctx context.Context, tenantID string) error {
	return c.c.Invalidate(ctx, Key(tenantID))
}

// Stats are the hit and miss counts of the underlying cache.
func (c *Cache) Stats() cache.Stats { return c.c.Stats() }

// Disabled is a cache.Store that always misses and stores nothing: every lookup reads the database.
// cmd/api and cmd/dispatcher use it when REDIS_ADDR is not set.
func Disabled() cache.Store { return disabled{} }

type disabled struct{}

func (disabled) Get(context.Context, string) ([]byte, error)              { return nil, cache.ErrMiss }
func (disabled) Set(context.Context, string, []byte, time.Duration) error { return nil }
func (disabled) Del(context.Context, ...string) error                     { return nil }
