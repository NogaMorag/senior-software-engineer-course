package linkcheck

import "context"

// CheckPool does the same job as Check, built by hand from channels: a
// generator stage, a fixed pool of Workers goroutines (fan-out) and one
// collector that puts results back in input order (fan-in).
func (c *Checker) CheckPool(ctx context.Context, urls []string) ([]Result, error) {
	// TODO 3 (day 9, section 5.7):
	//   - ctx, cancel := context.WithCancelCause(ctx); defer cancel(nil)
	//   - jobs carries {index, url}; out carries {index, Result};
	//   - stage 1, generator: send every job, stop on <-ctx.Done(), always close(jobs);
	//   - stage 2, fan-out: c.workers() goroutines (sync.WaitGroup.Go) ranging over jobs;
	//     with FailFast, a failed result calls cancel(err) with an error wrapping it;
	//   - one more goroutine: wg.Wait(), then close(out);
	//   - stage 3, fan-in: pre-fill results with ErrSkipped, then range over out
	//     until it is closed (never stop early: section 5.9);
	//   - if ctx.Err() != nil return context.Cause(ctx).
	panic("TODO 3: CheckPool")
}
