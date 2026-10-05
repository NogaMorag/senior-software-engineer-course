// Command relay publishes the transactional outbox to Redpanda (milestone M4). Run as many as you like:
// they campaign for one advisory lock (milestone M6, chapter 18), one leads and publishes, the others
// wait and take over within a second or two when it dies. The leader's term fences its marks.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"example.com/hookline/internal/bus"
	"example.com/hookline/internal/config"
	"example.com/hookline/internal/leader"
	"example.com/hookline/internal/outbox"
	"example.com/hookline/internal/platform"
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
	log := platform.NewLogger(os.Stdout, cfg.LogLevel)
	ctx, stop := wire.SignalContext()
	defer stop()

	pool, err := wire.Pool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	producer, err := bus.NewProducer(bus.ParseBrokers(cfg.KafkaBrokers))
	if err != nil {
		return err
	}
	defer producer.Close()

	elector := leader.New(pool, leader.Config{Key: outbox.RelayLockKey, Logger: log})
	log.Info("relay campaigning", slog.String("topic", outbox.Topic), slog.Duration("interval", cfg.RelayInterval))
	err = elector.Run(ctx, func(ctx context.Context, term int64) error {
		relay := &outbox.Relay{DB: pool, Publisher: bus.KafkaPublisher{Client: producer}, Term: term}
		log.Info("relay leading", slog.Int64("term", term))
		return relay.Run(ctx, cfg.RelayInterval, log)
	})
	log.Info("relay stopped")
	if errors.Is(err, context.Canceled) {
		return nil // SIGTERM: a clean stop
	}
	return err
}
