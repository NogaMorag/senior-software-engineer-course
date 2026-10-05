// Command loadgen sends GET requests to one URL from several goroutines for a
// fixed time, so there is something to profile. It does not follow redirects.
//
//	go run ./cmd/loadgen -url http://localhost:8080/go -c 8 -d 20s
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	target := flag.String("url", "", "URL to request (required)")
	conc := flag.Int("c", 8, "concurrent workers")
	dur := flag.Duration("d", 20*time.Second, "how long to run")
	flag.Parse()
	if *target == "" {
		flag.Usage()
		os.Exit(2)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		// Count the 302 itself; never follow it to the real site.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport:     &http.Transport{MaxIdleConnsPerHost: *conc},
	}

	ctx, cancel := context.WithTimeout(context.Background(), *dur)
	defer cancel()
	var total, failed atomic.Int64
	g, ctx := errgroup.WithContext(ctx)
	for range *conc {
		g.Go(func() error {
			for ctx.Err() == nil {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, *target, nil)
				if err != nil {
					return err // a bad -url: stop everyone
				}
				resp, err := client.Do(req)
				if err != nil {
					if ctx.Err() == nil {
						failed.Add(1)
					}
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				total.Add(1)
				if resp.StatusCode >= 400 {
					failed.Add(1)
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "loadgen:", err)
		os.Exit(1)
	}
	fmt.Printf("requests %d, failed %d, %.0f req/s over %s with %d workers\n",
		total.Load(), failed.Load(), float64(total.Load())/dur.Seconds(), *dur, *conc)
}
