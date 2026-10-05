package bulkhead_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/hookline/internal/bulkhead"
)

func TestLimitIsNeverExceeded(t *testing.T) {
	b := bulkhead.New(3, time.Second)
	var inside, peak atomic.Int64
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			err := b.Do(context.Background(), func(context.Context) error {
				n := inside.Add(1)
				for {
					p := peak.Load()
					if n <= p || peak.CompareAndSwap(p, n) {
						break
					}
				}
				time.Sleep(5 * time.Millisecond)
				inside.Add(-1)
				return nil
			})
			if err != nil {
				t.Errorf("Do: %v", err)
			}
		})
	}
	wg.Wait()
	if got := peak.Load(); got > 3 {
		t.Fatalf("peak concurrency = %d, want at most 3", got)
	}
	if b.InUse() != 0 {
		t.Fatalf("InUse = %d after all calls returned, want 0", b.InUse())
	}
}

func TestFullFailsFast(t *testing.T) {
	tests := []struct {
		name    string
		maxWait time.Duration
		ctx     func() (context.Context, context.CancelFunc)
		want    error
		atMost  time.Duration
	}{
		{"no wait", 0, func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) },
			bulkhead.ErrFull, 20 * time.Millisecond},
		{"short wait", 30 * time.Millisecond, func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) },
			bulkhead.ErrFull, 500 * time.Millisecond},
		{"context ends first", time.Hour, func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(context.Background(), 30*time.Millisecond)
		}, context.DeadlineExceeded, 500 * time.Millisecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bulkhead.New(1, tt.maxWait)
			if err := b.Acquire(context.Background()); err != nil {
				t.Fatal(err)
			}
			defer b.Release()
			ctx, cancel := tt.ctx()
			defer cancel()
			start := time.Now()
			err := b.Acquire(ctx)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Acquire on a full bulkhead = %v, want %v", err, tt.want)
			}
			if el := time.Since(start); el > tt.atMost {
				t.Fatalf("Acquire took %v, want at most %v", el, tt.atMost)
			}
		})
	}
}

func TestWaiterGetsReleasedSlot(t *testing.T) {
	b := bulkhead.New(1, time.Second)
	if err := b.Acquire(context.Background()); err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		b.Release()
	}()
	if err := b.Acquire(context.Background()); err != nil {
		t.Fatalf("waiter: %v, want the slot that was released", err)
	}
	b.Release()
}
