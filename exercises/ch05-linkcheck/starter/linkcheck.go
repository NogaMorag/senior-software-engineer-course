// Package linkcheck checks many URLs concurrently: at most Workers requests in
// flight, a timeout per request, cancellation of the whole run, and results in
// the same order as the input.
package linkcheck

import (
	"context"
	"errors"
	"net/http"
	"time"
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
	// TODO 1 (day 8, section 5.4):
	//   - derive a context with c.Timeout from ctx, and defer its cancel;
	//   - build a GET request with http.NewRequestWithContext and send it with c.client();
	//   - measure Elapsed around the call;
	//   - on an error, if ctx.Err() is not nil, store ctx.Err() (timeout or cancellation),
	//     otherwise the error itself;
	//   - close the body (drain at most 64 KiB first) and store the status code;
	//   - a status of 400 or above is fmt.Errorf("%w: status %d", ErrBroken, code).
	panic("TODO 1: checkOne")
}

// Check checks every URL with at most Workers requests in flight and returns
// one result per URL, in input order. The error is non-nil when the run was cut
// short: by ctx, or by the first failure when FailFast is set.
func (c *Checker) Check(ctx context.Context, urls []string) ([]Result, error) {
	// TODO 2 (day 8, then day 9):
	//   day 8: results := make([]Result, len(urls)); one goroutine per URL with a
	//          sync.WaitGroup; each goroutine writes results[i] = c.checkOne(ctx, u).
	//   day 9: replace the WaitGroup with errgroup.WithContext and g.SetLimit(c.workers());
	//          a goroutine that sees gctx.Err() stores ErrSkipped; with FailFast a failed
	//          result returns an error; return g.Wait()'s error, else ctx.Err().
	panic("TODO 2: Check")
}
