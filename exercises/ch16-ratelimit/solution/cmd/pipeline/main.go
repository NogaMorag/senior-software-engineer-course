// Command pipeline runs a producer that is faster than its consumer, three ways: an
// unbounded queue, a bounded queue that blocks the producer, and one that rejects.
//
//	go run ./cmd/pipeline
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"sync"
	"text/tabwriter"
	"time"

	"example.com/ch16-ratelimit/bounded"
)

func main() {
	produce := flag.Duration("produce", time.Millisecond, "time between two items from the producer")
	consume := flag.Duration("consume", 2*time.Millisecond, "time the consumer spends on one item")
	run := flag.Duration("for", 2*time.Second, "how long the producer runs")
	size := flag.Int("size", 50, "capacity of the bounded queue")
	flag.Parse()

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "mode\tproduced\trejected\tconsumed\tmax queued\tmax wait\tdrain after stop")
	for _, mode := range []string{"unbounded", "block", "reject"} {
		r := pipeline(mode, *produce, *consume, *run, *size)
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%v\t%v\n", mode, r.produced, r.rejected, r.consumed, r.maxQueued,
			r.maxWait.Round(time.Millisecond), r.drain.Round(time.Millisecond))
	}
	tw.Flush()
}

type result struct {
	produced, rejected, consumed, maxQueued int
	maxWait, drain                          time.Duration
}

// pipeline moves timestamps from a producer to a consumer; the wait of an item is the time
// between being produced and being picked up.
func pipeline(mode string, produce, consume, run time.Duration, size int) result {
	var (
		r       result
		mu      sync.Mutex
		backlog []time.Time // the unbounded queue: a slice that grows
		cond    = sync.NewCond(&mu)
		q       = bounded.New[time.Time](size)
		stopped bool
	)
	var wg sync.WaitGroup
	wg.Go(func() { // consumer
		for {
			var item time.Time
			if mode == "unbounded" {
				mu.Lock()
				for len(backlog) == 0 && !stopped {
					cond.Wait()
				}
				if len(backlog) == 0 {
					mu.Unlock()
					return
				}
				item, backlog = backlog[0], backlog[1:]
				mu.Unlock()
			} else {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				v, ok := q.Get(ctx)
				cancel()
				if !ok {
					return // nothing for 100ms after the producer stopped
				}
				item = v
			}
			r.maxWait = max(r.maxWait, time.Since(item))
			time.Sleep(consume)
			r.consumed++
		}
	})

	start := time.Now()
	tick := time.NewTicker(produce)
	for time.Since(start) < run {
		<-tick.C
		now := time.Now()
		switch mode {
		case "unbounded":
			mu.Lock()
			backlog = append(backlog, now)
			r.maxQueued = max(r.maxQueued, len(backlog))
			mu.Unlock()
			cond.Signal()
		case "block":
			_ = q.Put(context.Background(), now) // the producer waits here: back-pressure
		case "reject":
			if err := q.TryPut(now); errors.Is(err, bounded.ErrFull) {
				r.rejected++
				continue
			}
		}
		r.produced++
		if mode != "unbounded" {
			r.maxQueued = max(r.maxQueued, q.Len())
		}
	}
	tick.Stop()
	stop := time.Now()
	mu.Lock()
	stopped = true
	mu.Unlock()
	cond.Signal()
	wg.Wait()
	r.drain = time.Since(stop)
	if mode != "unbounded" {
		r.drain -= 100 * time.Millisecond // the consumer's idle timeout is not drain time
	}
	return r
}
