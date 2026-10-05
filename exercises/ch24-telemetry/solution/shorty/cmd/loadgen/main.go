// Command loadgen sends a steady mix of requests to shorty, so the dashboard
// and the SLO rules have traffic to show: creates, resolves of known codes,
// and a few resolves of codes that do not exist (404). Rewritten for chapter
// 24; the chapter 6 version hammered one URL for profiling.
//
//	go run ./cmd/loadgen -base http://localhost:8080 -rps 20 -d 5m
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	base := flag.String("base", "http://localhost:8080", "shorty's base URL")
	rps := flag.Float64("rps", 20, "requests per second")
	dur := flag.Duration("d", time.Minute, "how long to run (0: until Ctrl-C)")
	create := flag.Float64("create", 0.2, "fraction of requests that create a link")
	miss := flag.Float64("miss", 0.05, "fraction of requests that resolve a missing code")
	flag.Parse()
	if *rps <= 0 {
		fmt.Fprintln(os.Stderr, "loadgen: -rps must be positive")
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if *dur > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *dur)
		defer cancel()
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		// Count the 302 itself; never follow it to the real site.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	var (
		mu    sync.Mutex
		codes []string
		wg    sync.WaitGroup
	)
	var byClass [6]atomic.Int64 // index 0: transport errors; 1..5: 1xx..5xx

	send := func(method, url, body string) int {
		req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
		if err != nil {
			return 0
		}
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := client.Do(req)
		if err != nil {
			if ctx.Err() == nil {
				byClass[0].Add(1)
			}
			return 0
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if c := resp.StatusCode / 100; c >= 1 && c <= 5 {
			byClass[c].Add(1)
		}
		return resp.StatusCode
	}

	one := func() {
		defer wg.Done()
		mu.Lock()
		known := len(codes)
		code := ""
		if known > 0 {
			code = codes[rand.IntN(known)]
		}
		mu.Unlock()
		switch p := rand.Float64(); {
		case known == 0 || p < *create:
			c := fmt.Sprintf("lg%08x", rand.Uint32())
			if send(http.MethodPost, *base+"/v1/links", `{"url":"https://example.com/`+c+`","code":"`+c+`"}`) == http.StatusCreated {
				mu.Lock()
				codes = append(codes, c)
				mu.Unlock()
			}
		case p < *create+*miss:
			send(http.MethodGet, *base+"/missing-"+fmt.Sprint(rand.Uint32()), "")
		default:
			send(http.MethodGet, *base+"/"+code, "")
		}
	}

	tick := time.NewTicker(time.Duration(float64(time.Second) / *rps))
	defer tick.Stop()
	start := time.Now()
	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
		case <-tick.C:
			wg.Add(1)
			go one()
		}
	}
	wg.Wait()
	fmt.Printf("%.0fs: 2xx %d, 3xx %d, 4xx %d, 5xx %d, transport errors %d\n", time.Since(start).Seconds(),
		byClass[2].Load(), byClass[3].Load(), byClass[4].Load(), byClass[5].Load(), byClass[0].Load())
}
