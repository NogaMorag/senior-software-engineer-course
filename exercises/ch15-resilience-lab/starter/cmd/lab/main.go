// Command lab runs the chapter's experiments in one process, each with and without the
// mechanism it demonstrates, and prints call counts. Your numbers are yours: record them.
//
//	go run ./cmd/lab -scenario storm
//	go run ./cmd/lab               # every scenario
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"example.com/ch15-resilience-lab/breaker"
	"example.com/ch15-resilience-lab/bulkhead"
	"example.com/ch15-resilience-lab/degrade"
	"example.com/ch15-resilience-lab/flaky"
	"example.com/ch15-resilience-lab/hop"
	"example.com/ch15-resilience-lab/retry"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout)) }

type config struct {
	n         int
	errorRate float64
	seed      uint64
	scale     time.Duration // multiplies every time span; tests use a small one
}

func run(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("lab", flag.ContinueOnError)
	scenario := fs.String("scenario", "all", "storm, deadline, breaker, bulkhead, degrade or all")
	var cfg config
	fs.IntVar(&cfg.n, "n", 100, "user requests in the storm scenario")
	fs.Float64Var(&cfg.errorRate, "error-rate", 1.0, "upstream error rate in the storm scenario")
	fs.Uint64Var(&cfg.seed, "seed", 1, "random seed for the upstream")
	scale := fs.Float64("scale", 1, "time scale for the timed scenarios")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	cfg.scale = time.Duration(*scale * float64(time.Millisecond))
	scenarios := map[string]func(io.Writer, config){
		"storm": storm, "deadline": deadlines, "breaker": breakers, "bulkhead": bulkheads, "degrade": degradation,
	}
	order := []string{"storm", "deadline", "breaker", "bulkhead", "degrade"}
	if *scenario != "all" {
		if _, ok := scenarios[*scenario]; !ok {
			fmt.Fprintf(out, "unknown scenario %q\n", *scenario)
			return 2
		}
		order = []string{*scenario}
	}
	for _, name := range order {
		scenarios[name](out, cfg)
		fmt.Fprintln(out)
	}
	return 0
}

// built reports whether fn runs without panicking, so the lab can skip what the starter
// has not built yet (its bodies panic with a TODO) instead of crashing.
func built(fn func()) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	fn()
	return true
}

// ms turns a number of lab milliseconds into a duration at the configured scale.
func (c config) ms(n int) time.Duration { return time.Duration(n) * c.scale }

// buildChain starts the upstream and three hops a, b and c in front of it, each hop
// retrying with its own policy, and returns the URL of a and the hops.
func buildChain(up http.Handler, policies [3]retry.Policy, propagate bool, max time.Duration) (string, []*hop.Hop, func()) {
	upSrv := httptest.NewServer(up)
	closers := []func(){upSrv.Close}
	next := upSrv.URL
	hops := make([]*hop.Hop, 3)
	for i := 2; i >= 0; i-- {
		hops[i] = &hop.Hop{
			Next:      &hop.Caller{URL: next, Client: upSrv.Client(), Retry: policies[i]},
			Max:       max,
			Reserve:   5 * time.Millisecond,
			Propagate: propagate,
		}
		s := httptest.NewServer(hops[i].Handler())
		closers = append(closers, s.Close)
		next = s.URL
	}
	return next, hops, func() {
		for i := len(closers) - 1; i >= 0; i-- {
			closers[i]()
		}
	}
}

// storm: three layers that each retry three times multiply one user request into 27
// upstream requests; a retry budget per layer keeps the multiplier near 1.
func storm(out io.Writer, cfg config) {
	fmt.Fprintf(out, "== storm: %d user requests, 3 layers, upstream error rate %.2f\n", cfg.n, cfg.errorRate)
	three := func(b *retry.Budget) retry.Policy {
		return retry.Policy{MaxAttempts: 3, Base: time.Millisecond, Cap: 4 * time.Millisecond, Budget: b}
	}
	one := retry.Policy{MaxAttempts: 1}
	configs := []struct {
		name     string
		policies func() [3]retry.Policy
	}{
		{"no retries", func() [3]retry.Policy { return [3]retry.Policy{one, one, one} }},
		{"3 attempts at every layer", func() [3]retry.Policy { return [3]retry.Policy{three(nil), three(nil), three(nil)} }},
		{"3 attempts at the top only", func() [3]retry.Policy { return [3]retry.Policy{three(nil), one, one} }},
		{"3 attempts + 10% budget per layer", func() [3]retry.Policy {
			return [3]retry.Policy{three(retry.NewBudget(0.1, 10)), three(retry.NewBudget(0.1, 10)), three(retry.NewBudget(0.1, 10))}
		}},
	}
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "config\tA\tB\tC\tupstream\tper request\tuser ok")
	for i, c := range configs {
		if i == 3 && !built(func() { retry.NewBudget(0.1, 1).Deposit() }) {
			fmt.Fprintf(tw, "%s\tskipped: retry.Budget not built yet\n", c.name)
			continue
		}
		up := flaky.New(cfg.errorRate, 0, cfg.seed)
		url, hops, stop := buildChain(up, c.policies(), true, 5*time.Second)
		client := &hop.Caller{URL: url, Client: &http.Client{}, Retry: one}
		ok := 0
		for range cfg.n {
			if client.Call(context.Background()) == nil {
				ok++
			}
		}
		stop()
		got := up.Counts().Received
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%.1f\t%d\n", c.name, hops[0].Received(), hops[1].Received(),
			hops[2].Received(), got, float64(got)/float64(cfg.n), ok)
	}
	tw.Flush()
}

// deadlines: a client with a 300ms budget calls a 1s upstream through three hops. With
// propagation the upstream is told to stop; without it, it finishes work nobody awaits.
func deadlines(out io.Writer, cfg config) {
	const requests = 10
	budget := cfg.ms(300)
	fmt.Fprintf(out, "== deadline: %d concurrent requests, client budget %v, upstream latency %v\n", requests, budget, cfg.ms(1000))
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "config\tclient errors\tclient waited\tupstream received\tupstream finished\tupstream abandoned")
	for _, propagate := range []bool{false, true} {
		up := flaky.New(0, cfg.ms(1000), cfg.seed)
		one := retry.Policy{MaxAttempts: 1}
		url, _, stop := buildChain(up, [3]retry.Policy{one, one, one}, propagate, cfg.ms(2000))
		client := &hop.Caller{URL: url, Client: &http.Client{}, Retry: one}
		var failures atomic.Int64
		var wg sync.WaitGroup
		start := time.Now()
		for range requests {
			wg.Go(func() {
				ctx, cancel := context.WithTimeout(context.Background(), budget)
				defer cancel()
				if client.Call(ctx) != nil {
					failures.Add(1) // its own deadline, or a 504 from a hop whose budget ended first
				}
			})
		}
		wg.Wait()
		waited := time.Since(start).Round(time.Millisecond)
		// Let the upstream finish or give up on everything it started.
		for deadline := time.Now().Add(cfg.ms(3000)); time.Now().Before(deadline); time.Sleep(cfg.ms(20)) {
			c := up.Counts()
			if c.Received > 0 && c.Received == c.Succeeded+c.Failed+c.Abandoned {
				break
			}
		}
		stop()
		c := up.Counts()
		name := "fixed 2s timeout per hop"
		if propagate {
			name = "budget propagated"
		}
		fmt.Fprintf(tw, "%s\t%d\t%v\t%d\t%d\t%d\n", name, failures.Load(), waited, c.Received, c.Succeeded, c.Abandoned)
	}
	tw.Flush()
}

// breakers: the upstream is down for a second, then healthy for a second; a client calls
// it every 10ms. The breaker turns most calls during the outage into instant refusals.
func breakers(out io.Writer, cfg config) {
	if !built(func() { _, _ = breaker.New(breaker.Settings{}).Allow() }) {
		fmt.Fprintln(out, "== breaker: skipped, the breaker is not built yet")
		return
	}
	fmt.Fprintf(out, "== breaker: upstream down for %v then healthy for %v, one call every %v\n", cfg.ms(1000), cfg.ms(1000), cfg.ms(10))
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "config\tcalls\tok\tupstream received\trefused locally\ttransitions")
	for _, withBreaker := range []bool{false, true} {
		up := flaky.New(1.0, cfg.ms(2), cfg.seed)
		srv := httptest.NewServer(up)
		var transitions []string
		var mu sync.Mutex
		c := &hop.Caller{URL: srv.URL, Client: srv.Client(), Retry: retry.Policy{MaxAttempts: 1}}
		if withBreaker {
			c.Breaker = breaker.New(breaker.Settings{
				FailureThreshold: 5, OpenTimeout: cfg.ms(200), HalfOpenProbes: 1,
				OnStateChange: func(from, to breaker.State) {
					mu.Lock()
					transitions = append(transitions, to.String())
					mu.Unlock()
				},
			})
		}
		calls, ok := 0, 0
		start := time.Now()
		recovered := false
		for time.Since(start) < cfg.ms(2000) {
			if !recovered && time.Since(start) >= cfg.ms(1000) {
				up.Set(0, cfg.ms(2))
				recovered = true
			}
			calls++
			if c.Call(context.Background()) == nil {
				ok++
			}
			time.Sleep(cfg.ms(10))
		}
		srv.Close()
		name := "no breaker"
		if withBreaker {
			name = "breaker (5 failures, 200ms open)"
		}
		mu.Lock()
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%d\n", name, calls, ok, up.Counts().Received, c.Refused(), len(transitions))
		mu.Unlock()
	}
	tw.Flush()
}

// bulkheads: 20 workers alternate calls to a slow dependency and a fast one. Without a
// bulkhead the slow one holds most workers; with 4 slots for it, the fast one keeps flowing.
func bulkheads(out io.Writer, cfg config) {
	const workers = 20
	if !built(func() { _ = bulkhead.New(1, 0).Acquire(context.Background()) }) {
		fmt.Fprintln(out, "== bulkhead: skipped, the bulkhead is not built yet")
		return
	}
	fmt.Fprintf(out, "== bulkhead: %d workers for %v, alternating a slow dependency (%v) and a fast one (%v)\n",
		workers, cfg.ms(2000), cfg.ms(500), cfg.ms(5))
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "config\tfast ok\tslow ok\tslow refused")
	for _, withBulkhead := range []bool{false, true} {
		slowSrv := httptest.NewServer(flaky.New(0, cfg.ms(500), cfg.seed))
		fastSrv := httptest.NewServer(flaky.New(0, cfg.ms(5), cfg.seed))
		slow := &hop.Caller{URL: slowSrv.URL, Client: slowSrv.Client(), Retry: retry.Policy{MaxAttempts: 1}}
		fast := &hop.Caller{URL: fastSrv.URL, Client: fastSrv.Client(), Retry: retry.Policy{MaxAttempts: 1}}
		if withBulkhead {
			slow.Bulkhead = bulkhead.New(4, 0)
		}
		var fastOK, slowOK atomic.Int64
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ms(2000))
		var wg sync.WaitGroup
		for w := range workers {
			wg.Go(func() {
				for i := w; ctx.Err() == nil; i++ {
					if i%2 == 0 {
						if slow.Call(ctx) == nil {
							slowOK.Add(1)
						}
					} else if fast.Call(ctx) == nil {
						fastOK.Add(1)
					}
				}
			})
		}
		wg.Wait()
		cancel()
		slowSrv.Close()
		fastSrv.Close()
		name := "shared workers, no bulkhead"
		if withBulkhead {
			name = "bulkhead: 4 slots for slow"
		}
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\n", name, fastOK.Load(), slowOK.Load(), slow.Refused())
	}
	tw.Flush()
}

// degradation: a value is fetched while the upstream is healthy, then the upstream fails.
// The stale cache keeps answering for MaxStale; an optional feature switched off costs nothing.
func degradation(out io.Writer, cfg config) {
	fmt.Fprintf(out, "== degrade: 20 reads %v apart, upstream fails after the first; stale allowed for %v\n", cfg.ms(100), cfg.ms(1000))
	up := flaky.New(0, 0, cfg.seed)
	srv := httptest.NewServer(up)
	defer srv.Close()
	c := &hop.Caller{URL: srv.URL, Client: srv.Client(), Retry: retry.Policy{MaxAttempts: 1}}
	cache := &degrade.Stale[string]{MaxStale: cfg.ms(1000)}
	fetch := func(ctx context.Context) (string, error) {
		if err := c.Call(ctx); err != nil {
			return "", err
		}
		return "price list", nil
	}
	fresh, stale, failed := 0, 0, 0
	for i := range 20 {
		if i == 1 {
			up.Set(1.0, 0)
		}
		_, isStale, err := cache.Get(context.Background(), fetch)
		switch {
		case err != nil:
			failed++
		case isStale:
			stale++
		default:
			fresh++
		}
		time.Sleep(cfg.ms(100))
	}
	var recommendations degrade.Feature
	recommendations.Set(false)
	before := up.Counts().Received
	for range 20 {
		degrade.Optional(context.Background(), &recommendations, []string{}, func(ctx context.Context) ([]string, error) {
			return nil, c.Call(ctx)
		})
	}
	fmt.Fprintf(out, "fresh=%d stale=%d failed=%d; with the feature off, 20 reads sent %d upstream requests\n",
		fresh, stale, failed, up.Counts().Received-before)
}
