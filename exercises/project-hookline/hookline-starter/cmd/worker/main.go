// Command worker sends due deliveries, signed, and retries failures with backoff (milestone M5). Run as
// many as you like: SKIP LOCKED and the lease keep them from sending the same delivery at once.
package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"

	"example.com/hookline/internal/backoff"
	"example.com/hookline/internal/breaker"
	"example.com/hookline/internal/config"
	"example.com/hookline/internal/delivery"
	"example.com/hookline/internal/metrics"
	"example.com/hookline/internal/ssrf"
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
	ctx, stop := wire.SignalContext()
	defer stop()
	log, flush, err := wire.Telemetry(ctx, "hookline-worker", cfg.LogLevel)
	if err != nil {
		return err
	}
	defer flush()
	m := metrics.Must()
	policy := ssrf.ForMode(cfg.DevMode)

	pool, err := wire.Pool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	p := &delivery.Pool{
		DB: pool,
		// Every connection is checked after DNS resolution (milestone M9, DNS rebinding).
		Sender:      delivery.Sender{Timeout: cfg.AttemptTimeout, Policy: &policy},
		Workers:     cfg.Workers,
		Lease:       cfg.Lease,
		Poll:        cfg.Poll,
		MaxAttempts: cfg.MaxAttempts,
		Backoff:     backoff.Policy{Base: cfg.BackoffBase, Cap: cfg.BackoffCap},
		Log:         log,
		Metrics:     m,
		Guards: &delivery.Guards{
			Breaker:     breaker.Settings{FailureThreshold: cfg.BreakerFailures, OpenTimeout: cfg.BreakerOpen},
			Concurrency: cfg.EndpointConcurrency,
			OnStateChange: func(endpointID string, from, to breaker.State) {
				// Called with the breaker's lock held: a counter and a log line, nothing that blocks.
				m.BreakerTransition(context.Background(), from.String(), to.String())
				log.Info("endpoint breaker", slog.String("endpoint_id", endpointID), slog.String("from", from.String()), slog.String("to", to.String()))
			},
		},
	}
	log.Info("worker running", slog.String("config", cfg.String()))
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return wire.Health(gctx, cfg.HealthAddr, log) })
	g.Go(func() error {
		err := p.Run(gctx) // returns after the drain: every claimed delivery has its result recorded
		stop()
		return err
	})
	err = g.Wait()
	log.Info("worker stopped")
	return err
}
