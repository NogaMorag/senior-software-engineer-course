package bounded_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/ch16-ratelimit/bounded"
)

func TestTryPutRejectsWhenFull(t *testing.T) {
	q := bounded.New[int](2)
	for i := range 2 {
		if err := q.TryPut(i); err != nil {
			t.Fatalf("TryPut %d: %v", i, err)
		}
	}
	if err := q.TryPut(3); !errors.Is(err, bounded.ErrFull) {
		t.Fatalf("TryPut on a full queue = %v, want ErrFull", err)
	}
	if q.Len() != 2 || q.Cap() != 2 {
		t.Fatalf("Len %d Cap %d, want 2 and 2", q.Len(), q.Cap())
	}
}

func TestPutWaitsForTheConsumer(t *testing.T) {
	q := bounded.New[int](1)
	ctx := context.Background()
	if err := q.Put(ctx, 1); err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(30 * time.Millisecond)
		q.Get(ctx)
	}()
	start := time.Now()
	if err := q.Put(ctx, 2); err != nil {
		t.Fatal(err)
	}
	if el := time.Since(start); el < 20*time.Millisecond {
		t.Fatalf("Put on a full queue returned after %v: it must wait for room", el)
	}
	if v, ok := q.Get(ctx); !ok || v != 2 {
		t.Fatalf("Get = %d, %v; want 2", v, ok)
	}
}

func TestPutGivesUpWithTheContext(t *testing.T) {
	q := bounded.New[int](1)
	_ = q.TryPut(1)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := q.Put(ctx, 2); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Put = %v, want context.DeadlineExceeded", err)
	}
	if _, ok := q.Get(context.Background()); !ok { // with a done ctx, select could pick either case
		t.Fatal("the item that was queued is gone")
	}
	if _, ok := q.Get(ctx); ok {
		t.Fatal("Get on an empty queue with a done context returned an item")
	}
}
