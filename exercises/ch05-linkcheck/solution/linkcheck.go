// Package linkcheck checks many URLs concurrently: at most Workers requests in
// flight, a timeout per request, cancellation of the whole run, and results in
// the same order as the input.
package linkcheck

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

// ErrBroken marks a response with status 400 or above.
var ErrBroken = errors.New("broken link")

// ErrSkipped marks a URL that was never requested because the run was cancelled first.
var ErrSkipped = errors.New("skipped")

// Result is the outcome of checking one URL.
type Result struct {
	URL     string
	Status  int           // HTTP status code; 0 when there was no response
	Err     error         // nil when the link is fine
	Elapsed time.Duration // time spent on the request
}

// Verdict is a one-word classification of a result, used in the report.
func (r Result) Verdict() string {
	switch {
	case r.Err == nil:
		return "ok"
	case errors.Is(r.Err, ErrSkipped):
		return "skipped"
	case errors.Is(r.Err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(r.Err, context.Canceled):
		return "cancelled"
	case errors.Is(r.Err, ErrBroken):
		return "broken"
	default:
		return "error"
	}
}

// Checker holds the settings of a run. The zero value is not useful: set
// Workers and Timeout.
type Checker struct {
	Client   *http.Client  // nil means http.DefaultClient
	Workers  int           // maximum number of requests in flight (at least 1)
	Timeout  time.Duration // limit for each single request
	FailFast bool          // cancel the remaining checks after the first failure
}

func (c *Checker) client() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return http.DefaultClient
}

func (c *Checker) workers() int {
	return max(c.Workers, 1)
}

// checkOne requests one URL with its own timeout, derived from ctx.
func (c *Checker) checkOne(ctx context.Context, url string) Result {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	res := Result{URL: url}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		res.Err = err
		return res
	}
	start := time.Now()
	resp, err := c.client().Do(req)
	res.Elapsed = time.Since(start)
	if err != nil {
		// The client reports context.Cause(ctx), which may be any error.
		// Classify by what happened to ctx itself: timeout or cancellation.
		if ctxErr := ctx.Err(); ctxErr != nil {
			err = ctxErr
		}
		res.Err = err
		return res
	}
	defer resp.Body.Close()
	// Drain a bounded amount so the connection can be reused.
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))

	res.Status = resp.StatusCode
	if resp.StatusCode >= 400 {
		res.Err = fmt.Errorf("%w: status %d", ErrBroken, resp.StatusCode)
	}
	return res
}

// Check checks every URL with at most Workers requests in flight and returns
// one result per URL, in input order. The error is non-nil when the run was cut
// short: by ctx, or by the first failure when FailFast is set.
func (c *Checker) Check(ctx context.Context, urls []string) ([]Result, error) {
	results := make([]Result, len(urls))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(c.workers())

	for i, u := range urls {
		g.Go(func() error {
			if gctx.Err() != nil {
				results[i] = Result{URL: u, Err: ErrSkipped}
				return nil
			}
			results[i] = c.checkOne(gctx, u) // each goroutine writes its own element
			if c.FailFast && results[i].Err != nil {
				return fmt.Errorf("fail fast on %s: %w", u, results[i].Err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return results, err
	}
	return results, ctx.Err()
}
