// Command client sends requests to an upstream with each mechanism switched by a flag.
//
//	go run ./cmd/client -n 200 -c 10 -attempts 3 -budget 0.1 -breaker -bulkhead 5
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"example.com/ch15-resilience-lab/breaker"
	"example.com/ch15-resilience-lab/bulkhead"
	"example.com/ch15-resilience-lab/hop"
	"example.com/ch15-resilience-lab/retry"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout)) }

func run(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("client", flag.ContinueOnError)
	url := fs.String("url", "http://127.0.0.1:8015/", "upstream URL")
	n := fs.Int("n", 100, "number of requests")
	c := fs.Int("c", 10, "concurrent workers")
	timeout := fs.Duration("timeout", 500*time.Millisecond, "budget for each request, retries included")
	attempts := fs.Int("attempts", 1, "attempts per request, the first included")
	budget := fs.Float64("budget", 0, "retry budget ratio, e.g. 0.1; 0 means no budget")
	useBreaker := fs.Bool("breaker", false, "put a circuit breaker in front of the upstream")
	slots := fs.Int("bulkhead", 0, "concurrent calls allowed to the upstream; 0 means no bulkhead")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	caller := &hop.Caller{
		URL:    *url,
		Client: &http.Client{},
		Retry:  retry.Policy{MaxAttempts: *attempts, Base: 10 * time.Millisecond, Cap: 200 * time.Millisecond},
	}
	if *budget > 0 {
		caller.Retry.Budget = retry.NewBudget(*budget, 10)
	}
	if *useBreaker {
		caller.Breaker = breaker.New(breaker.Settings{FailureThreshold: 5, OpenTimeout: time.Second})
	}
	if *slots > 0 {
		caller.Bulkhead = bulkhead.New(*slots, 100*time.Millisecond) // wait briefly for a slot, then give up
	}

	outcomes := map[string]int{}
	var mu sync.Mutex
	jobs := make(chan struct{})
	var wg sync.WaitGroup
	start := time.Now()
	for range *c {
		wg.Go(func() {
			for range jobs {
				ctx, cancel := context.WithTimeout(context.Background(), *timeout)
				err := caller.Call(ctx)
				cancel()
				mu.Lock()
				outcomes[outcome(err)]++
				mu.Unlock()
			}
		})
	}
	for range *n {
		jobs <- struct{}{}
	}
	close(jobs)
	wg.Wait()

	fmt.Fprintf(out, "requests=%d elapsed=%v sent=%d refused-locally=%d\n",
		*n, time.Since(start).Round(time.Millisecond), caller.Sent(), caller.Refused())
	keys := make([]string, 0, len(outcomes))
	for k := range outcomes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(out, "  %-16s %d\n", k, outcomes[k])
	}
	return 0
}

func outcome(err error) string {
	var se *hop.StatusError
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, retry.ErrBudgetExhausted):
		return "budget-exhausted"
	case errors.Is(err, breaker.ErrOpen), errors.Is(err, breaker.ErrTooManyProbes):
		return "breaker-open"
	case errors.Is(err, bulkhead.ErrFull):
		return "bulkhead-full"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.As(err, &se):
		return fmt.Sprintf("status-%d", se.Code)
	}
	return "error"
}
