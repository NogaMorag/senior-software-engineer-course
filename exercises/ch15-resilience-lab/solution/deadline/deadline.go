// Package deadline carries a time budget across HTTP hops.
//
// A context deadline does not cross the network by itself. The caller sends the time it
// has left in a header; the server turns the header back into a context deadline; before
// calling further downstream, the server keeps a reserve for its own work and passes on
// only what is left.
package deadline

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"
)

// Header carries the remaining budget in whole milliseconds.
const Header = "X-Timeout-Ms"

// ErrNoBudget means the time left is not enough to make a downstream call worth starting.
var ErrNoBudget = errors.New("deadline: no budget left for the downstream call")

// Middleware gives every request a context deadline: the budget in the header, never
// more than max, and max when the header is missing or malformed.
func Middleware(max time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		budget := max
		if ms, err := strconv.ParseInt(r.Header.Get(Header), 10, 64); err == nil && ms > 0 {
			budget = min(budget, time.Duration(ms)*time.Millisecond)
		}
		ctx, cancel := context.WithTimeout(r.Context(), budget)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Downstream derives the context for one downstream call: the caller's deadline minus
// reserve, the time this hop keeps to handle the answer (or the failure) and reply.
// Without a deadline on ctx it returns ctx unchanged.
func Downstream(ctx context.Context, reserve time.Duration) (context.Context, context.CancelFunc, error) {
	dl, ok := ctx.Deadline()
	if !ok {
		return ctx, func() {}, nil
	}
	if time.Until(dl) <= reserve {
		return ctx, func() {}, ErrNoBudget
	}
	ctx, cancel := context.WithDeadline(ctx, dl.Add(-reserve))
	return ctx, cancel, nil
}

// Inject writes the time left on req's context into the header; nothing without a deadline.
func Inject(req *http.Request) {
	dl, ok := req.Context().Deadline()
	if !ok {
		return
	}
	ms := max(time.Until(dl).Milliseconds(), 1)
	req.Header.Set(Header, strconv.FormatInt(ms, 10))
}
