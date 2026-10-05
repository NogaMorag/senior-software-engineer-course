// Command elect is one candidate in a leader election. Run it in two terminals with
// different names; the one that wins works, the other waits. Kill the leader (Ctrl+C,
// kill -9, or pg_terminate_backend on its session) and watch the other take over.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch18-leader/leader"
)

func main() {
	name := flag.String("name", "candidate", "name shown in the log lines")
	key := flag.Int64("key", 1818, "advisory lock key shared by all candidates")
	heartbeat := flag.Duration("heartbeat", time.Second, "leader heartbeat interval")
	retry := flag.Duration("retry", time.Second, "follower retry interval")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("name", *name)
	if err := run(*key, *heartbeat, *retry, log); err != nil {
		log.Error("elect failed", "err", err)
		os.Exit(1)
	}
}

func run(key int64, heartbeat, retry time.Duration, log *slog.Logger) error {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return fmt.Errorf("set DATABASE_URL; see the README")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := leader.EnsureSchema(ctx, pool); err != nil {
		return err
	}
	e := leader.New(pool, leader.Config{Key: key, Heartbeat: heartbeat, Retry: retry, Logger: log})
	log.Info("campaigning", "key", key)
	err = e.Run(ctx, func(ctx context.Context, term int64) error {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for n := 1; ; n++ {
			select {
			case <-ctx.Done():
				log.Info("stopping work", "term", term)
				return ctx.Err()
			case <-tick.C:
				log.Info("working as leader", "term", term, "tick", n)
			}
		}
	})
	if ctx.Err() != nil {
		log.Info("shut down cleanly")
		return nil
	}
	return err
}
