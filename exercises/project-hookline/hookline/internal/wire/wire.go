// Package wire builds the shared connections every binary needs, so the main functions stay short:
// the PostgreSQL pool, the Redis client, the subscription cache and the ingest limiter over it (or their
// in-process stand-ins without Redis), and the signal context.
package wire

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"example.com/hookline/internal/cache"
	"example.com/hookline/internal/db"
	"example.com/hookline/internal/ratelimit"
	"example.com/hookline/internal/subcache"
)

// SignalContext is cancelled by the first SIGINT or SIGTERM, which starts the graceful shutdown; a
// second signal kills the process at once (chapter 6, section 6.5).
func SignalContext() (context.Context, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	context.AfterFunc(ctx, stop)
	return ctx, stop
}

// Pool opens the PostgreSQL pool with the limits of internal/db.
func Pool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := db.Config(url)
	if err != nil {
		return nil, err
	}
	return db.Connect(ctx, cfg)
}

// Redis returns a client for addr, or nil when addr is empty (no Redis configured). The client
// connects lazily; the first command that cannot reach Redis fails, and its caller decides.
func Redis(addr string) *redis.Client {
	if addr == "" {
		return nil
	}
	return redis.NewClient(&redis.Options{Addr: addr})
}

// Subscriptions returns the subscription cache: over rdb, or without a cache when rdb is nil.
func Subscriptions(rdb *redis.Client, loader subcache.Loader) *subcache.Cache {
	opts := cache.Options{TTL: subcache.DefaultTTL, Jitter: subcache.DefaultJitter}
	if rdb == nil {
		return subcache.New(subcache.Disabled(), loader, opts)
	}
	return subcache.New(cache.NewRedisStore(rdb), loader, opts)
}

// Limiter returns the per-tenant ingest limiter: one token bucket per tenant in Redis, shared by every
// API replica; or, without Redis (dev mode), one in this process, which replicas would multiply.
func Limiter(rdb *redis.Client) ratelimit.Limiter {
	if rdb == nil {
		return &ratelimit.Local{}
	}
	return &ratelimit.Redis{Client: rdb, Prefix: "hookline:rl:ingest:"}
}
