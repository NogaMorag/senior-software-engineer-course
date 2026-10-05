// Package hop is one service in a call chain: an HTTP handler that calls the next
// service through a Caller, which can retry, break the circuit and limit concurrency.
package hop

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"example.com/ch15-resilience-lab/breaker"
	"example.com/ch15-resilience-lab/bulkhead"
	"example.com/ch15-resilience-lab/deadline"
	"example.com/ch15-resilience-lab/retry"
)

// StatusError is a response that counts as a failure: 5xx or 429. RetryAfter is the
// wait the server asked for in a 429 or 503 (zero when it did not ask).
type StatusError struct {
	Code       int
	RetryAfter time.Duration
}

func (e *StatusError) Error() string { return fmt.Sprintf("downstream answered %d", e.Code) }

// ServerDelay makes the retry loop wait at least RetryAfter (retry.ServerDelay).
func (e *StatusError) ServerDelay() time.Duration { return e.RetryAfter }

// ParseRetryAfter reads a Retry-After value (RFC 9110, section 10.2.3): either
// delay-seconds (one or more digits) or an HTTP-date, which becomes the time left
// until that date as seen from now. A date in the past means no wait. ok is false
// for an empty or malformed value.
func ParseRetryAfter(v string, now time.Time) (d time.Duration, ok bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	// ParseUint takes digits only (no sign); on overflow it returns the maximum value.
	if secs, err := strconv.ParseUint(v, 10, 64); err == nil || errors.Is(err, strconv.ErrRange) {
		if secs > math.MaxInt64/uint64(time.Second) {
			return math.MaxInt64, true // centuries: no deadline will allow it anyway
		}
		return time.Duration(secs) * time.Second, true
	}
	t, err := http.ParseTime(v) // IMF-fixdate, and the two obsolete forms recipients must accept
	if err != nil {
		return 0, false
	}
	return max(t.Sub(now), 0), true
}

// Retryable is the retry policy's filter: retry failures of the downstream, never our own
// refusals (an open breaker or a full bulkhead would only refuse again) or our own deadline.
func Retryable(err error) bool {
	switch {
	case errors.Is(err, breaker.ErrOpen), errors.Is(err, breaker.ErrTooManyProbes),
		errors.Is(err, bulkhead.ErrFull), errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return false
	}
	return true
}

// Caller calls one downstream URL. Every mechanism is optional: a nil Breaker or
// Bulkhead and a one-attempt Retry policy give a plain HTTP call.
type Caller struct {
	URL      string
	Client   *http.Client
	Retry    retry.Policy
	Breaker  *breaker.Breaker
	Bulkhead *bulkhead.Bulkhead

	sent     atomic.Int64 // HTTP requests that left this process
	refused  atomic.Int64 // attempts refused locally by the breaker or the bulkhead
	failures atomic.Int64 // calls that returned an error after all attempts
}

// Sent, Refused and Failures report the counters.
func (c *Caller) Sent() int64     { return c.sent.Load() }
func (c *Caller) Refused() int64  { return c.refused.Load() }
func (c *Caller) Failures() int64 { return c.failures.Load() }

// Call makes one logical call: possibly several attempts.
func (c *Caller) Call(ctx context.Context) error {
	p := c.Retry
	if p.Retryable == nil {
		p.Retryable = Retryable
	}
	_, err := retry.Do(ctx, p, c.attempt)
	if err != nil {
		c.failures.Add(1)
	}
	return err
}

// attempt is one try: a bulkhead slot, then the breaker's permission, then the request.
func (c *Caller) attempt(ctx context.Context) (err error) {
	if c.Bulkhead != nil {
		if err := c.Bulkhead.Acquire(ctx); err != nil {
			c.refused.Add(1)
			return err
		}
		defer c.Bulkhead.Release()
	}
	if c.Breaker != nil {
		done, refusal := c.Breaker.Allow()
		if refusal != nil {
			c.refused.Add(1)
			return refusal
		}
		defer func() { done(err) }() // err is the named result: the outcome of send
	}
	return c.send(ctx)
}

func (c *Caller) send(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return err
	}
	deadline.Inject(req)
	c.sent.Add(1)
	resp, err := c.Client.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr // classify by what happened to our context, not by the transport's wording
		}
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body) // drain so the connection is reused
	if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
		se := &StatusError{Code: resp.StatusCode}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			se.RetryAfter, _ = ParseRetryAfter(resp.Header.Get("Retry-After"), c.now())
		}
		return se
	}
	return nil
}

// now is the retry policy's clock, so an HTTP-date is measured against the same time
// the retry loop compares with the deadline.
func (c *Caller) now() time.Time {
	if c.Retry.Now != nil {
		return c.Retry.Now()
	}
	return time.Now()
}

func isStatus(err error, code int) bool {
	var se *StatusError
	return errors.As(err, &se) && se.Code == code
}

// Hop is a service that answers each request by calling Next once (with retries).
type Hop struct {
	Next    *Caller
	Max     time.Duration // budget for a request that carries none
	Reserve time.Duration // kept back from the downstream call for this hop's own reply
	// Propagate passes the caller's remaining budget downstream. False is the classic
	// bug: a fixed timeout of Max on a fresh context, blind to the caller having given up.
	Propagate bool

	received atomic.Int64
}

// Received reports how many requests reached this hop.
func (h *Hop) Received() int64 { return h.received.Load() }

// Handler returns the hop's HTTP handler.
func (h *Hop) Handler() http.Handler {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.received.Add(1)
		var (
			ctx    context.Context
			cancel context.CancelFunc
		)
		if h.Propagate {
			var err error
			ctx, cancel, err = deadline.Downstream(r.Context(), h.Reserve)
			if err != nil {
				http.Error(w, err.Error(), http.StatusGatewayTimeout)
				return
			}
		} else {
			ctx, cancel = context.WithTimeout(context.Background(), h.Max)
		}
		defer cancel()
		err := h.Next.Call(ctx)
		switch {
		case err == nil:
			fmt.Fprintln(w, "ok")
		case errors.Is(err, context.DeadlineExceeded), isStatus(err, http.StatusGatewayTimeout):
			http.Error(w, err.Error(), http.StatusGatewayTimeout) // a timeout further down is still a timeout
		case errors.Is(err, breaker.ErrOpen), errors.Is(err, bulkhead.ErrFull):
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		default:
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
	})
	return deadline.Middleware(h.Max, inner)
}
