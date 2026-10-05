package catalog

import (
	"context"

	"example.com/ch12-cache-aside/internal/cache"
)

// CachedProducts is the same read and write path as ProductCache, built on the
// generic helper (day 32): concurrent misses for one product share one query.
type CachedProducts struct {
	store  *Store
	cache  *cache.Cache[Product]
	prefix string
}

// NewCachedProducts returns products read through c.
func NewCachedProducts(store *Store, c *cache.Cache[Product], prefix string) *CachedProducts {
	return &CachedProducts{store: store, cache: c, prefix: prefix}
}

// Get returns the product, loading it at most once per key at a time on a miss.
func (p *CachedProducts) Get(ctx context.Context, id int64) (Product, error) {
	return p.cache.Get(ctx, ProductKey(p.prefix, id), func(ctx context.Context) (Product, error) {
		return p.store.GetProduct(ctx, id)
	})
}

// UpdatePrice commits the new price, then invalidates the cached product.
func (p *CachedProducts) UpdatePrice(ctx context.Context, id, priceCents int64) error {
	if err := p.store.UpdatePrice(ctx, id, priceCents); err != nil {
		return err
	}
	return p.cache.Invalidate(ctx, ProductKey(p.prefix, id))
}
