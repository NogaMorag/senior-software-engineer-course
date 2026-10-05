// Command relay publishes the transactional outbox to Redpanda (milestone M4). Run as many as you like:
// they campaign for one advisory lock (milestone M6, chapter 18), one leads and publishes, the others
// wait and take over within a second or two when it dies. The leader's term fences its marks.
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"

	"example.com/hookline/internal/bus"
	"example.com/hookline/internal/config"
	"example.com/hookline/internal/leader"
	"example.com/hookline/internal/metrics"
	"example.com/hookline/internal/outbox"
	"example.com/hookline/internal/wire"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "relay:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.LookupEnv)
	if err == nil {
		err = cfg.Require("DATABASE_URL", "KAFKA_BROKERS")
	}
	if err != nil {
		return err
	}
	ctx, stop := wire.SignalContext()
	defer stop()
	log, flush, err := wire.Telemetry(ctx, "hookline-relay", cfg.LogLevel)
	if err != nil {
		return err
	}
	defer flush()

	pool, err := wire.Pool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	// kotel: a publish span per record, child of the ingest request's span, and traceparent in the
	// record headers for the dispatcher (milestone M8).
	hooks, _ := bus.Tracing()
	producer, err := bus.NewProducer(bus.ParseBrokers(cfg.KafkaBrokers), hooks)
	if err != nil {
		return err
	}
	defer producer.Close()
	// Every relay reports the outbox, leader or not: the gauges say whether anyone is publishing.
	if err := metrics.ObserveOutbox(otel.GetMeterProvider().Meter(metrics.Scope),
		func(ctx context.Context) (int64, time.Duration, error) { return outbox.Stats(ctx, pool) }); err != nil {
		return err
	}

	elector := leader.New(pool, leader.Config{Key: outbox.RelayLockKey, Logger: log})
	log.Info("relay campaigning", slog.String("topic", outbox.Topic), slog.Duration("interval", cfg.RelayInterval))
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return wire.Health(gctx, cfg.HealthAddr, log) })
	g.Go(func() error {
		err := elector.Run(gctx, func(ctx context.Context, term int64) error {
			relay := &outbox.Relay{DB: pool, Publisher: bus.KafkaPublisher{Client: producer}, Term: term}
			log.Info("relay leading", slog.Int64("term", term))
			return relay.Run(ctx, cfg.RelayInterval, log)
		})
		stop() // the elector only returns when ctx is done; stop the health listener too
		return err
	})
	err = g.Wait()
	log.Info("relay stopped")
	if errors.Is(err, context.Canceled) {
		return nil // SIGTERM: a clean stop
	}
	return err
}
