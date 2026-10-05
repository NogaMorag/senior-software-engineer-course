package linkcheck

import (
	"context"
	"fmt"
	"sync"
)

// CheckPool does the same job as Check, built by hand from channels: a
// generator stage, a fixed pool of Workers goroutines (fan-out) and one
// collector that puts results back in input order (fan-in).
func (c *Checker) CheckPool(ctx context.Context, urls []string) ([]Result, error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	type job struct {
		i   int
		url string
	}
	type done struct {
		i   int
		res Result
	}
	jobs := make(chan job)
	out := make(chan done)

	// Stage 1: the generator. It stops early when the run is cancelled and
	// always closes jobs, so the workers' range loops end.
	go func() {
		defer close(jobs)
		for i, u := range urls {
			select {
			case jobs <- job{i, u}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Stage 2: fan-out to a fixed number of workers.
	var wg sync.WaitGroup
	for range c.workers() {
		wg.Go(func() {
			for j := range jobs {
				r := c.checkOne(ctx, j.url)
				if c.FailFast && r.Err != nil {
					cancel(fmt.Errorf("fail fast on %s: %w", j.url, r.Err))
				}
				out <- done{j.i, r} // the collector reads until out is closed
			}
		})
	}
	// Close out only after every worker has stopped sending.
	go func() {
		wg.Wait()
		close(out)
	}()

	// Stage 3: fan-in. Results arrive in completion order; the index puts
	// each one back in its place.
	results := make([]Result, len(urls))
	for i, u := range urls {
		results[i] = Result{URL: u, Err: ErrSkipped}
	}
	for d := range out {
		results[d.i] = d.res
	}
	if ctx.Err() != nil {
		return results, context.Cause(ctx)
	}
	return results, nil
}
