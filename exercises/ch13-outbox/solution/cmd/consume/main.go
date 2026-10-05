// Command consume joins a consumer group, prints every record and every rebalance, and
// commits offsets only after it has "processed" a batch. Run two in two terminals to
// watch partitions move; kill one with Ctrl-C (or kill -9) to watch redelivery.
//
//	go run ./cmd/consume -topic demo -group demo-group -name A -sleep 500ms
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"example.com/ch13-outbox/bus"
	"example.com/ch13-outbox/consumer"
)

func main() {
	topic := flag.String("topic", "demo", "topic to consume")
	group := flag.String("group", "demo-group", "consumer group")
	name := flag.String("name", "A", "label printed on every line")
	sleep := flag.Duration("sleep", 0, "pretend each record takes this long to process")
	flag.Parse()

	brokers := bus.Brokers()
	if brokers == nil {
		fmt.Fprintln(os.Stderr, "consume: set KAFKA_BROKERS, for example localhost:19092")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cl, err := bus.NewGroupConsumer(brokers, *group, *topic,
		kgo.OnPartitionsAssigned(func(_ context.Context, _ *kgo.Client, p map[string][]int32) {
			fmt.Printf("[%s] assigned %v\n", *name, p)
		}),
		kgo.OnPartitionsRevoked(func(_ context.Context, _ *kgo.Client, p map[string][]int32) {
			fmt.Printf("[%s] revoked  %v\n", *name, p)
		}),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "consume:", err)
		os.Exit(1)
	}
	defer cl.Close() // leaves the group, so the others rebalance at once

	err = consumer.Run(ctx, cl, func(ctx context.Context, r *kgo.Record) error {
		fmt.Printf("[%s] partition=%d offset=%d key=%s value=%s\n", *name, r.Partition, r.Offset, r.Key, r.Value)
		select {
		case <-time.After(*sleep):
			return nil
		case <-ctx.Done():
			return ctx.Err() // not processed, so not committed
		}
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "consume:", err)
		os.Exit(1)
	}
}
