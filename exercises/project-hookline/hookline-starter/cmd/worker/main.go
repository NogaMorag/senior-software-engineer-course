// Command worker sends due deliveries, signed, and retries failures with backoff (milestone M5). Run as
// many as you like: SKIP LOCKED and the lease keep them from sending the same delivery at once.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"example.com/hookline/internal/backoff"
	"example.com/hookline/internal/breaker"
	"example.com/hookline/internal/config"
	"example.com/hookline/internal/delivery"
	"example.com/hookline/internal/platform"
	"example.com/hookline/internal/wire"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "worker:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.LookupEnv)
	if err == nil {
		err = cfg.Require("DATABASE_URL")
	}
	if err != nil {
		return err
	}
	log := platform.NewLogger(os.Stdout, cfg.LogLevel)
	ctx, stop := wire.SignalContext()
	defer stop()

	pool, err := wire.Pool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	p := &delivery.Pool{
		DB:          pool,
		Sender:      delivery.Sender{Timeout: cfg.AttemptTimeout},
		Workers:     cfg.Workers,
		Lease:       cfg.Lease,
		Poll:        cfg.Poll,
		MaxAttempts: cfg.MaxAttempts,
		Backoff:     backoff.Policy{Base: cfg.BackoffBase, Cap: cfg.BackoffCap},
		Log:         log,
		Guards: &delivery.Guards{
			Breaker:     breaker.Settings{FailureThreshold: cfg.BreakerFailures, OpenTimeout: cfg.BreakerOpen},
			Concurrency: cfg.EndpointConcurrency,
			OnStateChange: func(endpointID string, from, to breaker.State) {
				log.Info("endpoint breaker", slog.String("endpoint_id", endpointID), slog.String("from", from.String()), slog.String("to", to.String()))
			},
		},
	}
	log.Info("worker running", slog.String("config", cfg.String()))
	err = p.Run(ctx) // returns after the drain: every claimed delivery has its result recorded
	log.Info("worker stopped")
	return err
}
