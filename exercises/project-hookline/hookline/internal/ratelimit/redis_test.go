package ratelimit_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"example.com/hookline/internal/ratelimit"
)

// twoInstances returns two separate clients, standing in for two replicas of a service,
// and a key prefix private to this test; the keys are deleted at the end.
func twoInstances(t *testing.T) (*redis.Client, *redis.Client, string) {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("integration test: set REDIS_ADDR to the Hookline Redis (make up)")
	}
	a := redis.NewClient(&redis.Options{Addr: addr})
	b := redis.NewClient(&redis.Options{Addr: addr})
	if err := a.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping redis at %s: %v", addr, err)
	}
	prefix := fmt.Sprintf("test:%d:%d:", os.Getpid(), time.Now().UnixNano())
	t.Cleanup(func() {
		ctx := context.Background()
		if keys, err := a.Keys(ctx, prefix+"*").Result(); err == nil && len(keys) > 0 {
			a.Del(ctx, keys...)
		}
		a.Close()
		b.Close()
	})
	return a, b, prefix
}

func TestRedisTwoInstancesShareOneLimit(t *testing.T) {
	a, b, prefix := twoInstances(t)
	ctx := context.Background()
	la := limiter{&ratelimit.Redis{Client: a, Prefix: prefix}, 1, 5}
	lb := limiter{&ratelimit.Redis{Client: b, Prefix: prefix}, 1, 5}
	allowed := 0
	var last ratelimit.Decision
	for i := range 10 {
		l := la
		if i%2 == 1 {
			l = lb
		}
		d, err := l.Allow(ctx, "tenant-a")
		if err != nil {
			t.Fatal(err)
		}
		if d.Allowed {
			allowed++
		}
		last = d
	}
	if allowed != 5 {
		t.Fatalf("two instances allowed %d of 10 rapid requests, want 5: they must share one bucket", allowed)
	}
	if last.Allowed || last.RetryAfter <= 0 || last.RetryAfter > time.Second {
		t.Fatalf("last decision = %+v, want refused with a retry-after up to 1s", last)
	}
	if d, _ := la.Allow(ctx, "tenant-b"); !d.Allowed || d.Remaining != 4 {
		t.Fatalf("tenant-b = %+v, want allowed with 4 left: another key, another bucket", d)
	}
}

func TestRedisRefills(t *testing.T) {
	a, _, prefix := twoInstances(t)
	ctx := context.Background()
	l := limiter{&ratelimit.Redis{Client: a, Prefix: prefix}, 20, 1} // a token every 50ms
	if d, _ := l.Allow(ctx, "k"); !d.Allowed {
		t.Fatal("first request refused")
	}
	if d, _ := l.Allow(ctx, "k"); d.Allowed {
		t.Fatal("second request allowed with an empty bucket")
	}
	time.Sleep(80 * time.Millisecond)
	if d, _ := l.Allow(ctx, "k"); !d.Allowed {
		t.Fatal("refused after 80ms at 20 tokens per second")
	}
	ttl, err := a.PTTL(ctx, prefix+"k").Result()
	if err != nil || ttl <= 0 || ttl > 2*time.Second {
		t.Fatalf("key TTL = %v (%v), want an expiry so idle buckets disappear", ttl, err)
	}
}

func TestRedisIsAtomicUnderConcurrency(t *testing.T) {
	a, b, prefix := twoInstances(t)
	ctx := context.Background()
	limiters := []limiter{
		{&ratelimit.Redis{Client: a, Prefix: prefix}, 0.001, 10},
		{&ratelimit.Redis{Client: b, Prefix: prefix}, 0.001, 10},
	}
	var allowed atomic.Int64
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() {
			if d, err := limiters[i%2].Allow(ctx, "hot"); err == nil && d.Allowed {
				allowed.Add(1)
			}
		})
	}
	wg.Wait()
	if got := allowed.Load(); got != 10 {
		t.Fatalf("100 concurrent requests from two instances: %d allowed, want exactly 10", got)
	}
}

// naiveAllow is the version the chapter warns about: read, decide in Go, write back.
// Between the read and the write another instance can read the same value.
func naiveAllow(ctx context.Context, rdb *redis.Client, key string, burst int) bool {
	v, err := rdb.Get(ctx, key).Result()
	used := 0
	if err == nil {
		used, _ = strconv.Atoi(v)
	}
	if used >= burst {
		return false
	}
	rdb.Set(ctx, key, used+1, time.Minute)
	return true
}

// TestNaiveReadThenWriteOvershoots runs the naive version under the same load. How far
// it overshoots depends on your machine: run with -v and record the number.
func TestNaiveReadThenWriteOvershoots(t *testing.T) {
	a, b, prefix := twoInstances(t)
	ctx := context.Background()
	clients := []*redis.Client{a, b}
	var allowed atomic.Int64
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() {
			if naiveAllow(ctx, clients[i%2], prefix+"naive", 10) {
				allowed.Add(1)
			}
		})
	}
	wg.Wait()
	t.Logf("naive read-then-write allowed %d of 100 with a limit of 10", allowed.Load())
	if allowed.Load() < 10 {
		t.Fatalf("allowed %d, fewer than the limit", allowed.Load())
	}
}

func TestRedisSurvivesScriptFlush(t *testing.T) {
	a, _, prefix := twoInstances(t)
	ctx := context.Background()
	l := limiter{&ratelimit.Redis{Client: a, Prefix: prefix}, 1, 3}
	if _, err := l.Allow(ctx, "k"); err != nil {
		t.Fatal(err)
	}
	// A restart or failover empties the script cache; EVALSHA then answers NOSCRIPT.
	if err := a.ScriptFlush(ctx).Err(); err != nil {
		t.Fatal(err)
	}
	d, err := l.Allow(ctx, "k")
	if err != nil || !d.Allowed || d.Remaining != 1 {
		t.Fatalf("after SCRIPT FLUSH: %+v, %v; want allowed with 1 left (the bucket survived, the script reloaded)", d, err)
	}
}

// limiter pins one rate and burst to a ratelimit.Redis, the way these tests were written in
// chapter 16; Hookline passes them per call, per tenant.
type limiter struct {
	r     *ratelimit.Redis
	rate  float64
	burst int
}

func (l limiter) Allow(ctx context.Context, key string) (ratelimit.Decision, error) {
	return l.r.Allow(ctx, key, l.rate, l.burst)
}
