package catalog_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	catalog "example.com/ch12-cache-aside"
	"example.com/ch12-cache-aside/internal/cache"
)

// env is one test's private world: a PostgreSQL schema with the 2,000 products and a
// Redis key prefix. Both are removed at the end.
type env struct {
	store  *catalog.Store
	rdb    *redis.Client
	prefix string
}

func setup(t *testing.T) env {
	t.Helper()
	url, addr := os.Getenv("DATABASE_URL"), os.Getenv("REDIS_ADDR")
	if url == "" || addr == "" {
		t.Skip("integration test: set DATABASE_URL and REDIS_ADDR to the chapter 12 stack; see README")
	}
	ctx := context.Background()
	name := fmt.Sprintf("ch12_test_%d_%d", os.Getpid(), time.Now().UnixNano())

	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+name+" CASCADE"); err != nil {
			t.Errorf("drop schema: %v", err)
		}
	})
	cfg, err := catalog.Config(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = name
	pool, err := catalog.Connect(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(ctx, catalog.Schema); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping redis at %s: %v", addr, err)
	}
	prefix := name + ":"
	t.Cleanup(func() {
		keys, err := rdb.Keys(context.Background(), prefix+"*").Result()
		if err == nil && len(keys) > 0 {
			rdb.Del(context.Background(), keys...)
		}
		rdb.Close()
	})
	return env{store: catalog.NewStore(pool), rdb: rdb, prefix: prefix}
}

// getter is what both caches offer, so each test runs against both.
type getter interface {
	Get(ctx context.Context, id int64) (catalog.Product, error)
	UpdatePrice(ctx context.Context, id, priceCents int64) error
}

func caches(e env) map[string]getter {
	helper := cache.New[catalog.Product](cache.NewRedisStore(e.rdb), cache.Options{TTL: time.Minute, Jitter: 10 * time.Second})
	return map[string]getter{
		"by hand":     catalog.NewProductCache(e.rdb, e.store, e.prefix, time.Minute, 10*time.Second),
		"with helper": catalog.NewCachedProducts(e.store, helper, e.prefix),
	}
}

func TestMissThenHit(t *testing.T) {
	for _, name := range []string{"by hand", "with helper"} {
		t.Run(name, func(t *testing.T) {
			e := setup(t)
			c := caches(e)[name]
			ctx := context.Background()
			for i := range 2 {
				p, err := c.Get(ctx, 42)
				if err != nil {
					t.Fatal(err)
				}
				if p.ID != 42 || p.Name != "Product 42" || p.PriceCents != 500+42*37 {
					t.Fatalf("Get #%d = %+v", i+1, p)
				}
			}
			if got := e.store.Calls(); got != 1 {
				t.Errorf("database calls = %d after a miss and a hit, want 1", got)
			}
			ttl, err := e.rdb.PTTL(ctx, catalog.ProductKey(e.prefix, 42)).Result()
			if err != nil || ttl < 59*time.Second || ttl > 70*time.Second {
				t.Errorf("PTTL = %v, %v; want between 1m and 1m10s", ttl, err)
			}
		})
	}
}

func TestWriteInvalidates(t *testing.T) {
	for _, name := range []string{"by hand", "with helper"} {
		t.Run(name, func(t *testing.T) {
			e := setup(t)
			c := caches(e)[name]
			ctx := context.Background()
			if _, err := c.Get(ctx, 7); err != nil {
				t.Fatal(err)
			}
			if err := c.UpdatePrice(ctx, 7, 999); err != nil {
				t.Fatal(err)
			}
			n, err := e.rdb.Exists(ctx, catalog.ProductKey(e.prefix, 7)).Result()
			if err != nil || n != 0 {
				t.Fatalf("key exists = %d, %v after the write; want deleted", n, err)
			}
			p, err := c.Get(ctx, 7)
			if err != nil || p.PriceCents != 999 {
				t.Errorf("Get after the write = %+v, %v; want price 999", p, err)
			}
			if got := e.store.Calls(); got != 2 {
				t.Errorf("database calls = %d, want 2 (one per miss)", got)
			}
		})
	}
}

func TestNotFoundIsNotCached(t *testing.T) {
	for _, name := range []string{"by hand", "with helper"} {
		t.Run(name, func(t *testing.T) {
			e := setup(t)
			c := caches(e)[name]
			ctx := context.Background()
			if _, err := c.Get(ctx, 99999); !errors.Is(err, catalog.ErrNotFound) {
				t.Fatalf("Get(99999) error = %v, want ErrNotFound", err)
			}
			if n, _ := e.rdb.Exists(ctx, catalog.ProductKey(e.prefix, 99999)).Result(); n != 0 {
				t.Errorf("a missing product was cached")
			}
			if err := c.UpdatePrice(ctx, 99999, 100); !errors.Is(err, catalog.ErrNotFound) {
				t.Errorf("UpdatePrice(99999) error = %v, want ErrNotFound", err)
			}
		})
	}
}

// burst releases n readers of one cold product together and returns the database calls.
func burst(t *testing.T, e env, c getter, n int) int64 {
	t.Helper()
	e.store.Latency = 200 * time.Millisecond
	e.store.ResetCalls()
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range n {
		wg.Go(func() {
			<-start
			if _, err := c.Get(context.Background(), 42); err != nil {
				t.Error(err)
			}
		})
	}
	close(start)
	wg.Wait()
	return e.store.Calls()
}

func TestStampede(t *testing.T) {
	const n = 30
	t.Run("by hand", func(t *testing.T) {
		e := setup(t)
		got := burst(t, e, caches(e)["by hand"], n)
		t.Logf("%d concurrent misses without singleflight: %d database calls", n, got)
		if got < 2 {
			t.Errorf("database calls = %d; a stampede should send more than one", got)
		}
	})
	t.Run("with helper", func(t *testing.T) {
		e := setup(t)
		got := burst(t, e, caches(e)["with helper"], n)
		t.Logf("%d concurrent misses with singleflight: %d database calls", n, got)
		if got != 1 {
			t.Errorf("database calls = %d, want 1", got)
		}
	})
}

func TestIncrViewsIsAtomic(t *testing.T) {
	e := setup(t)
	const n = 200
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			if _, err := catalog.IncrViews(context.Background(), e.rdb, e.prefix, 42); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	got, err := e.rdb.Get(context.Background(), e.prefix+"views:42").Int64()
	if err != nil || got != n {
		t.Errorf("views = %d, %v; want %d", got, err, n)
	}
}

func TestAllowFixedWindow(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	key := e.prefix + "limit:tenant-1"
	var got []bool
	for range 5 {
		ok, err := catalog.Allow(ctx, e.rdb, key, 3, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, ok)
	}
	if fmt.Sprint(got) != "[true true true false false]" {
		t.Errorf("five calls with limit 3 = %v", got)
	}
	if ttl, _ := e.rdb.PTTL(ctx, key).Result(); ttl <= 0 || ttl > time.Second {
		t.Errorf("window PTTL = %v, want in (0, 1s]", ttl)
	}
	time.Sleep(1100 * time.Millisecond)
	if ok, _ := catalog.Allow(ctx, e.rdb, key, 3, time.Second); !ok {
		t.Errorf("first call of the next window was refused")
	}
}
