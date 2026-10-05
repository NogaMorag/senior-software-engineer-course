// Command server is a service with a fixed capacity: -workers requests at a time, each
// taking -work. Requests beyond that wait, unless shedding or rate limiting is switched on.
//
//	go run ./cmd/server                       # no protection: waiting requests pile up
//	go run ./cmd/server -shed 8               # at most 8 in flight, the rest get 503
//	go run ./cmd/server -rate 50 -burst 20    # per-tenant token bucket in memory
//	go run ./cmd/server -rate 50 -burst 20 -redis 127.0.0.1:6316
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"example.com/ch16-ratelimit/ratelimit"
	"example.com/ch16-ratelimit/shed"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8016", "listen address")
	workers := flag.Int("workers", 4, "requests the service can work on at once")
	work := flag.Duration("work", 20*time.Millisecond, "time one request needs")
	shedLimit := flag.Int("shed", 0, "requests in flight before shedding; 0 means no shedding")
	reserved := flag.Int("reserved-high", 0, "part of -shed kept for requests with X-Priority: high")
	rate := flag.Float64("rate", 0, "tokens per second per tenant; 0 means no rate limit")
	burst := flag.Int("burst", 10, "bucket size per tenant")
	redisAddr := flag.String("redis", "", "Redis address for a shared limiter; empty means in memory")
	flag.Parse()
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// The capacity: a semaphore. Waiting for it is the queue nobody sees.
	capacity := make(chan struct{}, *workers)
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case capacity <- struct{}{}:
		case <-r.Context().Done():
			return // the client gave up while waiting
		}
		defer func() { <-capacity }()
		time.Sleep(*work)
		fmt.Fprintln(w, "done")
	})

	if *rate > 0 {
		var l ratelimit.Limiter = &ratelimit.Local{Rate: *rate, Burst: *burst}
		if *redisAddr != "" {
			l = &ratelimit.Redis{Client: redis.NewClient(&redis.Options{Addr: *redisAddr}), Prefix: "rl:ch16:", Rate: *rate, Burst: *burst}
		}
		h = ratelimit.Middleware(ratelimit.Options{
			Limiter: l, Rate: *rate, Burst: *burst, FailOpen: true,
			Key: func(r *http.Request) string {
				if t := r.Header.Get("X-Tenant"); t != "" {
					return t
				}
				return "anonymous"
			},
			OnError: func(err error) { log.Warn("rate limiter failed, serving anyway", "err", err) },
		}, h)
	}
	var shedder *shed.Shedder
	if *shedLimit > 0 {
		// Outermost: a shed request costs nothing, not even a round trip to Redis.
		shedder = shed.New(shed.Options{
			Limit: *shedLimit, ReservedForHigh: *reserved, RetryAfter: time.Second,
			Priority: shed.ByHeader("X-Priority", "high"),
		})
		h = shedder.Wrap(h)
	}

	mux := http.NewServeMux()
	mux.Handle("/work", h)
	mux.HandleFunc("GET /stats", func(w http.ResponseWriter, r *http.Request) {
		stats := map[string]int64{"waiting_or_working": int64(len(capacity))}
		if shedder != nil {
			stats["served"], stats["shed_low"], stats["shed_high"] = shedder.Served(), shedder.Shed(shed.Low), shedder.Shed(shed.High)
		}
		_ = json.NewEncoder(w).Encode(stats)
	})
	srv := &http.Server{Addr: *addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		grace, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(grace)
	}()
	log.Info("listening", "addr", *addr, "workers", *workers, "work", *work, "shed", *shedLimit, "rate", *rate)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("serve", "err", err)
		os.Exit(1)
	}
}
