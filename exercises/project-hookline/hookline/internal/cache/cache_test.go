package cache_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"example.com/hookline/internal/cache"
)

// redisClient connects to REDIS_ADDR and returns a key prefix private to this test;
// the test's keys are deleted at the end.
func redisClient(t *testing.T) (*redis.Client, string) {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("integration test: set REDIS_ADDR to the Hookline Redis (make up)")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping redis at %s: %v", addr, err)
	}
	prefix := fmt.Sprintf("test:%d:%d:", os.Getpid(), time.Now().UnixNano())
	t.Cleanup(func() {
		keys, err := rdb.Keys(context.Background(), prefix+"*").Result()
		if err == nil && len(keys) > 0 {
			rdb.Del(context.Background(), keys...)
		}
		rdb.Close()
	})
	return rdb, prefix
}

// counted returns a load function that counts its calls and returns value.
func counted(value string, delay time.Duration) (func(context.Context) (string, error), *atomic.Int64) {
	var n atomic.Int64
	return func(context.Context) (string, error) {
		n.Add(1)
		time.Sleep(delay)
		return value, nil
	}, &n
}

func TestJittered(t *testing.T) {
	tests := []struct {
		name         string
		base, jitter time.Duration
	}{
		{"no jitter", time.Minute, 0},
		{"negative jitter", time.Minute, -time.Second},
		{"ten seconds", time.Minute, 10 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen := map[time.Duration]bool{}
			for range 1000 {
				got := cache.Jittered(tt.base, tt.jitter)
				hi := tt.base + max(tt.jitter, 0)
				if got < tt.base || (tt.jitter > 0 && got >= hi) || (tt.jitter <= 0 && got != tt.base) {
					t.Fatalf("Jittered(%v, %v) = %v, outside [%v, %v)", tt.base, tt.jitter, got, tt.base, hi)
				}
				seen[got] = true
			}
			if tt.jitter > 0 && len(seen) < 2 {
				t.Errorf("1000 draws gave one value: no jitter at all")
			}
		})
	}
}

func TestMissThenHit(t *testing.T) {
	rdb, prefix := redisClient(t)
	ctx := context.Background()
	c := cache.New[string](cache.NewRedisStore(rdb), cache.Options{TTL: time.Minute})
	load, loads := counted("v1", 0)

	for i, want := range []cache.Stats{{Misses: 1, Loads: 1}, {Hits: 1, Misses: 1, Loads: 1}} {
		got, err := c.Get(ctx, prefix+"k", load)
		if err != nil || got != "v1" {
			t.Fatalf("Get #%d = %q, %v; want v1", i+1, got, err)
		}
		if s := c.Stats(); s != want {
			t.Errorf("after Get #%d stats = %+v, want %+v", i+1, s, want)
		}
	}
	if loads.Load() != 1 {
		t.Errorf("load called %d times, want 1", loads.Load())
	}
	if r := c.Stats().HitRatio(); r != 0.5 {
		t.Errorf("hit ratio = %v, want 0.5", r)
	}
}

func TestTTLIsSetWithJitter(t *testing.T) {
	rdb, prefix := redisClient(t)
	ctx := context.Background()
	opts := cache.Options{TTL: time.Minute, Jitter: 10 * time.Second}
	c := cache.New[string](cache.NewRedisStore(rdb), opts)
	load, _ := counted("v1", 0)
	if _, err := c.Get(ctx, prefix+"k", load); err != nil {
		t.Fatal(err)
	}
	ttl, err := rdb.PTTL(ctx, prefix+"k").Result()
	if err != nil {
		t.Fatal(err)
	}
	if ttl < opts.TTL-time.Second || ttl > opts.TTL+opts.Jitter {
		t.Errorf("PTTL = %v, want between %v and %v", ttl, opts.TTL, opts.TTL+opts.Jitter)
	}
}

func TestInvalidateForcesReload(t *testing.T) {
	rdb, prefix := redisClient(t)
	ctx := context.Background()
	c := cache.New[string](cache.NewRedisStore(rdb), cache.Options{TTL: time.Minute})
	if _, err := c.Get(ctx, prefix+"k", func(context.Context) (string, error) { return "old", nil }); err != nil {
		t.Fatal(err)
	}
	if err := c.Invalidate(ctx, prefix+"k"); err != nil {
		t.Fatal(err)
	}
	got, err := c.Get(ctx, prefix+"k", func(context.Context) (string, error) { return "new", nil })
	if err != nil || got != "new" {
		t.Errorf("Get after Invalidate = %q, %v; want new", got, err)
	}
}

func TestLoadErrorIsNotCached(t *testing.T) {
	rdb, prefix := redisClient(t)
	ctx := context.Background()
	c := cache.New[string](cache.NewRedisStore(rdb), cache.Options{TTL: time.Minute})
	boom := errors.New("database down")
	if _, err := c.Get(ctx, prefix+"k", func(context.Context) (string, error) { return "", boom }); !errors.Is(err, boom) {
		t.Fatalf("Get = %v, want %v", err, boom)
	}
	got, err := c.Get(ctx, prefix+"k", func(context.Context) (string, error) { return "v1", nil })
	if err != nil || got != "v1" {
		t.Errorf("Get after a failed load = %q, %v; want v1", got, err)
	}
}

// TestConcurrentMissesLoadOnce is the stampede: 50 readers miss the same cold key at
// once. With singleflight the source sees one call, and every reader gets the value.
func TestConcurrentMissesLoadOnce(t *testing.T) {
	rdb, prefix := redisClient(t)
	ctx := context.Background()
	c := cache.New[string](cache.NewRedisStore(rdb), cache.Options{TTL: time.Minute})
	load, loads := counted("v1", 100*time.Millisecond)

	const n = 50
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)
	for range n {
		wg.Go(func() {
			<-start
			v, err := c.Get(ctx, prefix+"cold", load)
			if err != nil {
				t.Error(err)
			}
			results <- v
		})
	}
	close(start)
	wg.Wait()
	close(results)
	for v := range results {
		if v != "v1" {
			t.Errorf("a reader got %q, want v1", v)
		}
	}
	if got := loads.Load(); got != 1 {
		t.Errorf("load called %d times for %d concurrent misses, want 1", got, n)
	}
}

// TestStaleFillIsBoundedByTTL replays the race of section 12.3 step by step: a reader
// misses and reads the old value; a writer commits and invalidates; the reader then
// fills the cache with the old value. Nothing deletes it again: only the TTL does.
func TestStaleFillIsBoundedByTTL(t *testing.T) {
	rdb, prefix := redisClient(t)
	ctx := context.Background()
	const ttl = 300 * time.Millisecond
	c := cache.New[string](cache.NewRedisStore(rdb), cache.Options{TTL: ttl})
	key := prefix + "raced"

	var source atomic.Value
	source.Store("old")
	read, release := make(chan struct{}), make(chan struct{})
	done := make(chan string)
	go func() { // the slow reader
		v, err := c.Get(ctx, key, func(context.Context) (string, error) {
			v := source.Load().(string) // 1. reads the old row
			close(read)
			<-release // 3. is descheduled while the writer runs
			return v, nil
		})
		if err != nil {
			t.Error(err)
		}
		done <- v // 4. has filled the cache with "old"
	}()

	<-read
	source.Store("new")                            // 2a. the writer commits
	if err := c.Invalidate(ctx, key); err != nil { // 2b. and invalidates: nothing to delete yet
		t.Fatal(err)
	}
	close(release)
	if v := <-done; v != "old" {
		t.Fatalf("slow reader returned %q, want old", v)
	}

	fresh := func(context.Context) (string, error) { return source.Load().(string), nil }
	if v, _ := c.Get(ctx, key, fresh); v != "old" {
		t.Fatalf("right after the race Get = %q, want the stale old", v)
	}
	time.Sleep(ttl + 100*time.Millisecond)
	if v, _ := c.Get(ctx, key, fresh); v != "new" {
		t.Errorf("after the TTL Get = %q, want new", v)
	}
}

// TestStoreDownFallsBackToSource needs no Redis: the client points at a closed port.
func TestStoreDownFallsBackToSource(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1, DialTimeout: 200 * time.Millisecond})
	t.Cleanup(func() { rdb.Close() })
	c := cache.New[string](cache.NewRedisStore(rdb), cache.Options{TTL: time.Minute})
	got, err := c.Get(context.Background(), "k", func(context.Context) (string, error) { return "from-db", nil })
	if err != nil || got != "from-db" {
		t.Fatalf("Get with Redis down = %q, %v; want from-db", got, err)
	}
	if s := c.Stats(); s.Errors != 1 || s.Loads != 1 {
		t.Errorf("stats = %+v, want 1 error and 1 load", s)
	}
}
