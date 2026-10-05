// Command dispatcher consumes hookline.events and creates one delivery per matching endpoint
// (milestone M4). Run as many as the topic has partitions; the consumer group shares them out.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"example.com/hookline/internal/bus"
	"example.com/hookline/internal/config"
	"example.com/hookline/internal/dispatch"
	"example.com/hookline/internal/outbox"
	"example.com/hookline/internal/platform"
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
	log := platform.NewLogger(os.Stdout, cfg.LogLevel)
	ctx, stop := wire.SignalContext()
	defer stop()

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
	consumer, err := bus.NewGroupConsumer(bus.ParseBrokers(cfg.KafkaBrokers), cfg.ConsumerGroup, outbox.Topic)
	if err != nil {
		return err
	}
	defer consumer.Close()

	d := &dispatch.Dispatcher{DB: pool, Matcher: subs, Log: log}
	log.Info("dispatcher running", slog.String("group", cfg.ConsumerGroup), slog.Bool("cache", cfg.RedisAddr != ""))
	err = dispatch.Run(ctx, consumer, d.Handle, log)
	log.Info("dispatcher stopped", slog.Any("cache", subs.Stats()))
	return err
}
