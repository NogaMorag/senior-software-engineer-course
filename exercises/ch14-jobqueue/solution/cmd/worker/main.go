// Command worker runs a worker pool against DATABASE_URL until SIGINT or SIGTERM, then
// drains: it stops claiming and waits for the jobs it holds.
//
//	go run ./cmd/worker -init -enqueue 4 -ms 3000 -fail 1
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch14-jobqueue/backoff"
	"example.com/ch14-jobqueue/queue"
	"example.com/ch14-jobqueue/worker"
)

// demo is the payload of the "demo" kind: sleep ms, and fail while attempt <= fail.
type demo struct {
	MS   int `json:"ms"`
	Fail int `json:"fail"`
}

func main() {
	initSchema := flag.Bool("init", false, "create the tables first (fresh database only)")
	enqueue := flag.Int("enqueue", 0, "enqueue this many demo jobs before starting")
	ms := flag.Int("ms", 2000, "how long each demo job takes, in milliseconds")
	fail := flag.Int("fail", 0, "each demo job fails on its first N attempts")
	workers := flag.Int("workers", 2, "number of worker goroutines")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		fmt.Fprintln(os.Stderr, "worker: set DATABASE_URL; see README")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "worker:", err)
		os.Exit(1)
	}
	defer db.Close()

	if *initSchema {
		if _, err := db.Exec(ctx, queue.Schema); err != nil {
			fmt.Fprintln(os.Stderr, "worker: init:", err)
			os.Exit(1)
		}
	}
	for range *enqueue {
		payload, _ := json.Marshal(demo{MS: *ms, Fail: *fail})
		id, _, err := queue.Enqueue(ctx, db, "demo", payload, queue.Options{MaxAttempts: 3})
		if err != nil {
			fmt.Fprintln(os.Stderr, "worker:", err)
			os.Exit(1)
		}
		log.Info("enqueued", "job", id)
	}

	pool := &worker.Pool{
		DB:      db,
		Name:    "w",
		Workers: *workers,
		Lease:   30 * time.Second,
		Poll:    500 * time.Millisecond,
		Backoff: backoff.Policy{Base: time.Second, Cap: 30 * time.Second},
		Log:     log,
		Handlers: map[string]worker.Handler{
			"demo": func(ctx context.Context, j queue.Job) error {
				var d demo
				if err := json.Unmarshal(j.Payload, &d); err != nil {
					return queue.Permanent(err)
				}
				select {
				case <-time.After(time.Duration(d.MS) * time.Millisecond):
				case <-ctx.Done():
					return ctx.Err()
				}
				if j.Attempts <= d.Fail {
					return errors.New("simulated failure")
				}
				return nil
			},
		},
	}
	go func() {
		<-ctx.Done()
		log.Info("shutdown requested: no new claims, waiting for running jobs")
	}()
	_ = pool.Run(ctx)
	log.Info("drained, exiting")
}
