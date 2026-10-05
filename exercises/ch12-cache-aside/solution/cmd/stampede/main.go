// Command stampede sends N concurrent reads for one cold key, first through the
// hand-written cache and then through the singleflight helper, and prints how many
// of them reached PostgreSQL. It applies schema.sql to the database it is given.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	catalog "example.com/ch12-cache-aside"
	"example.com/ch12-cache-aside/internal/cache"
)

func main() {
	n := flag.Int("n", 50, "concurrent readers")
	latency := flag.Duration("latency", 200*time.Millisecond, "simulated query time")
	flag.Parse()
	if err := run(*n, *latency); err != nil {
		fmt.Fprintln(os.Stderr, "stampede:", err)
		os.Exit(1)
	}
}

func run(n int, latency time.Duration) error {
	ctx := context.Background()
	dbURL, addr := os.Getenv("DATABASE_URL"), os.Getenv("REDIS_ADDR")
	if dbURL == "" || addr == "" {
		return fmt.Errorf("set DATABASE_URL and REDIS_ADDR; see README")
	}
	cfg, err := catalog.Config(dbURL)
	if err != nil {
		return err
	}
	pool, err := catalog.Connect(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, catalog.Schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()

	store := catalog.NewStore(pool)
	store.Latency = latency
	const prefix, id = "demo:", 42
	naive := catalog.NewProductCache(rdb, store, prefix, time.Minute, 10*time.Second)
	helper := cache.New[catalog.Product](cache.NewRedisStore(rdb), cache.Options{TTL: time.Minute, Jitter: 10 * time.Second})
	collapsed := catalog.NewCachedProducts(store, helper, prefix)

	fmt.Printf("cold key, %d concurrent readers, %v per query\n", n, latency)
	for _, c := range []struct {
		name string
		get  func(context.Context, int64) (catalog.Product, error)
	}{
		{"without singleflight", naive.Get},
		{"with singleflight", collapsed.Get},
	} {
		if err := rdb.Del(ctx, catalog.ProductKey(prefix, id)).Err(); err != nil {
			return err
		}
		store.ResetCalls()
		start := time.Now()
		if err := burst(ctx, n, func(ctx context.Context) error { _, err := c.get(ctx, id); return err }); err != nil {
			return err
		}
		fmt.Printf("%-22s %3d database calls, %v\n", c.name+":", store.Calls(), time.Since(start).Round(time.Millisecond))
	}

	// The key is warm now: the same burst again is all hits.
	if err := burst(ctx, n, func(ctx context.Context) error { _, err := collapsed.Get(ctx, id); return err }); err != nil {
		return err
	}
	s := helper.Stats()
	fmt.Printf("helper stats: hits %d, misses %d, loads %d, hit ratio %.2f\n", s.Hits, s.Misses, s.Loads, s.HitRatio())
	return nil
}

// burst starts n goroutines, releases them together, and waits for all of them.
func burst(ctx context.Context, n int, call func(context.Context) error) error {
	var wg sync.WaitGroup
	errs := make(chan error, n)
	start := make(chan struct{})
	for range n {
		wg.Go(func() {
			<-start
			if err := call(ctx); err != nil {
				errs <- err
			}
		})
	}
	close(start)
	wg.Wait()
	close(errs)
	return <-errs // nil when the channel is empty
}
