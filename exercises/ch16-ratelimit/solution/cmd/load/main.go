// Command load sends requests at a fixed rate, whether or not earlier ones have returned
// (an open model, like real users), and prints status counts and latency percentiles.
//
//	go run ./cmd/load -url http://127.0.0.1:8016/work -rate 300 -for 10s
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout)) }

type sample struct {
	status  string
	latency time.Duration
}

func run(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("load", flag.ContinueOnError)
	url := fs.String("url", "http://127.0.0.1:8016/work", "target URL")
	rate := fs.Float64("rate", 100, "requests per second")
	dur := fs.Duration("for", 5*time.Second, "how long to send")
	timeout := fs.Duration("timeout", 2*time.Second, "client timeout per request")
	maxInFlight := fs.Int("max-in-flight", 1000, "requests this client keeps open at most; above it a request is dropped")
	var headers headerFlags
	fs.Var(&headers, "H", `header to send, e.g. -H "X-Tenant: a" (repeatable)`)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	client := &http.Client{Timeout: *timeout, Transport: &http.Transport{MaxIdleConnsPerHost: *maxInFlight}}
	slots := make(chan struct{}, *maxInFlight)
	var (
		mu      sync.Mutex
		samples []sample
		dropped int
		wg      sync.WaitGroup
	)
	interval := time.Duration(float64(time.Second) / *rate)
	tick := time.NewTicker(interval)
	defer tick.Stop()
	start := time.Now()
	for time.Since(start) < *dur {
		<-tick.C
		select {
		case slots <- struct{}{}:
		default:
			dropped++
			continue
		}
		wg.Go(func() {
			defer func() { <-slots }()
			s := once(client, *url, headers)
			mu.Lock()
			samples = append(samples, s)
			mu.Unlock()
		})
	}
	sending := time.Since(start)
	wg.Wait()
	report(out, samples, dropped, sending)
	return 0
}

func once(client *http.Client, url string, headers headerFlags) sample {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return sample{status: "error"}
	}
	for _, h := range headers {
		k, v, _ := strings.Cut(h, ":")
		req.Header.Set(strings.TrimSpace(k), strings.TrimSpace(v))
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return sample{status: "error", latency: time.Since(start)}
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return sample{status: fmt.Sprint(resp.StatusCode), latency: time.Since(start)}
}

func report(out io.Writer, samples []sample, dropped int, sending time.Duration) {
	fmt.Fprintf(out, "sent %d in %v (%.0f/s), dropped by the client %d\n",
		len(samples), sending.Round(time.Millisecond), float64(len(samples))/sending.Seconds(), dropped)
	by := map[string][]time.Duration{}
	for _, s := range samples {
		by[s.status] = append(by[s.status], s.latency)
	}
	keys := make([]string, 0, len(by))
	for k := range by {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Fprintf(out, "%-7s %7s %9s %9s %9s %9s\n", "status", "count", "p50", "p90", "p99", "max")
	for _, k := range keys {
		l := by[k]
		slices.Sort(l)
		fmt.Fprintf(out, "%-7s %7d %9v %9v %9v %9v\n", k, len(l), pct(l, 50), pct(l, 90), pct(l, 99), l[len(l)-1].Round(time.Millisecond))
	}
}

// pct returns the p-th percentile of sorted durations, rounded for reading.
func pct(sorted []time.Duration, p int) time.Duration {
	return sorted[(len(sorted)-1)*p/100].Round(time.Millisecond)
}

type headerFlags []string

func (h *headerFlags) String() string     { return strings.Join(*h, ", ") }
func (h *headerFlags) Set(v string) error { *h = append(*h, v); return nil }
