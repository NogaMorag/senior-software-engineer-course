// Command upstream is the flaky upstream as a standalone server.
//
//	go run ./cmd/upstream -error-rate 0.3 -latency 50ms
//	curl -s localhost:8015/counts
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/ch15-resilience-lab/flaky"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8015", "listen address")
	errorRate := flag.Float64("error-rate", 0.2, "fraction of requests answered 503, 0 to 1")
	latency := flag.Duration("latency", 20*time.Millisecond, "time spent on every request")
	seed := flag.Uint64("seed", 1, "random seed for the failure sequence")
	flag.Parse()

	up := flaky.New(*errorRate, *latency, *seed)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /counts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(up.Counts())
	})
	mux.Handle("/", up)
	srv := &http.Server{Addr: *addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	fmt.Printf("upstream on http://%s error-rate=%.2f latency=%v\n", *addr, *errorRate, *latency)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	fmt.Println(up.Counts())
}
