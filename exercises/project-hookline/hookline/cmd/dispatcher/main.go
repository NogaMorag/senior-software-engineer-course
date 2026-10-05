// Command dispatcher consumes hookline.events and creates one delivery per matching endpoint
// (milestone M4). Run as many as the topic has partitions; the consumer group shares them out.
package main

import (
	"fmt"

	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"

	"example.com/hookline/internal/bus"
	"example.com/hookline/internal/config"
	"example.com/hookline/internal/dispatch"
	"example.com/hookline/internal/metrics"
	"example.com/hookline/internal/outbox"
	"example.com/hookline/internal/store/postgres"
	"example.com/hookline/internal/wire"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "dispatcher:", err)
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
	log, flush, err := wire.Telemetry(ctx, "hookline-dispatcher", cfg.LogLevel)
	if err != nil {
		return err
	}
	defer flush()

	pool, err := wire.Pool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	rdb := wire.Redis(cfg.RedisAddr)
	if rdb != nil {
		defer rdb.Close()
	}
	subs := wire.Subscriptions(rdb, postgres.New(pool))
	// kotel: a receive span per record whose parent comes from the headers; Handle adds the process span.
	hooks, tracer := bus.Tracing()
	consumer, err := bus.NewGroupConsumer(bus.ParseBrokers(cfg.KafkaBrokers), cfg.ConsumerGroup, outbox.Topic, hooks)
	if err != nil {
		return err
	}
	defer consumer.Close()

	d := &dispatch.Dispatcher{DB: pool, Matcher: subs, Log: log, Tracer: tracer, Metrics: metrics.Must()}
	log.Info("dispatcher running", slog.String("group", cfg.ConsumerGroup), slog.Bool("cache", cfg.RedisAddr != ""))
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return wire.Health(gctx, cfg.HealthAddr, log) })
	g.Go(func() error {
		err := dispatch.Run(gctx, consumer, d.Handle, log)
		stop()
		return err
	})
	err = g.Wait()
	log.Info("dispatcher stopped", slog.Any("cache", subs.Stats()))
	return err
}
