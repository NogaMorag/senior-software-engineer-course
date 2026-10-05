package redislock

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// client connects to REDIS_ADDR, or skips. Each test gets its own key.
func client(t *testing.T) (*redis.Client, string) {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("integration test: set REDIS_ADDR to the chapter 18 Redis; see README")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = rdb.Close() })
	key := "ch18:lock:" + t.Name()
	if err := rdb.Del(context.Background(), key).Err(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rdb.Del(context.Background(), key) })
	return rdb, key
}

func TestSecondAcquireFailsUntilRelease(t *testing.T) {
	rdb, key := client(t)
	ctx := context.Background()
	a, err := Acquire(ctx, rdb, key, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Acquire(ctx, rdb, key, 10*time.Second); !errors.Is(err, ErrNotAcquired) {
		t.Fatalf("second Acquire: err = %v, want ErrNotAcquired", err)
	}
	if err := a.Release(ctx); err != nil {
		t.Fatal(err)
	}
	b, err := Acquire(ctx, rdb, key, 10*time.Second)
	if err != nil {
		t.Fatalf("Acquire after release: %v", err)
	}
	if err := b.Release(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestOnlyTheHolderCanRelease(t *testing.T) {
	rdb, key := client(t)
	ctx := context.Background()
	a, err := Acquire(ctx, rdb, key, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	impostor := &Lock{rdb: rdb, key: key, token: "not-the-token"}
	if err := impostor.Release(ctx); !errors.Is(err, ErrNotHeld) {
		t.Fatalf("impostor Release: err = %v, want ErrNotHeld", err)
	}
	if n, _ := rdb.Exists(ctx, key).Result(); n != 1 {
		t.Fatal("the impostor deleted the holder's lock")
	}
	if err := a.Release(ctx); err != nil {
		t.Fatal(err)
	}
}

// A holder that pauses longer than the TTL (a GC pause, a slow disk, a stopped VM) still
// believes it holds the lock when it wakes up, while someone else really does.
func TestExpiredHolderAndNewHolderOverlap(t *testing.T) {
	rdb, key := client(t)
	ctx := context.Background()
	a, err := Acquire(ctx, rdb, key, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond) // a pauses; nothing tells it the key expired

	b, err := Acquire(ctx, rdb, key, 10*time.Second)
	if err != nil {
		t.Fatalf("b could not acquire after a's TTL: %v", err)
	}
	// a wakes up. It finds out only if it asks Redis, and only after the fact:
	if err := a.Extend(ctx, 10*time.Second); !errors.Is(err, ErrNotHeld) {
		t.Fatalf("a.Extend: err = %v, want ErrNotHeld", err)
	}
	if err := a.Release(ctx); !errors.Is(err, ErrNotHeld) {
		t.Fatalf("a.Release: err = %v, want ErrNotHeld (and b's lock untouched)", err)
	}
	if err := b.Release(ctx); err != nil {
		t.Fatalf("b.Release: %v; a must not have deleted b's lock", err)
	}
}

func TestExtendKeepsTheLock(t *testing.T) {
	rdb, key := client(t)
	ctx := context.Background()
	a, err := Acquire(ctx, rdb, key, 300*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	for range 3 {
		time.Sleep(150 * time.Millisecond)
		if err := a.Extend(ctx, 300*time.Millisecond); err != nil {
			t.Fatalf("Extend: %v", err)
		}
	}
	if _, err := Acquire(ctx, rdb, key, time.Second); !errors.Is(err, ErrNotAcquired) {
		t.Fatalf("Acquire while extended: err = %v, want ErrNotAcquired", err)
	}
	if err := a.Release(ctx); err != nil {
		t.Fatal(err)
	}
}
