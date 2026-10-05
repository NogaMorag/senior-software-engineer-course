// Command produce sends keyed messages and prints where each one landed.
//
//	go run ./cmd/produce -topic demo -n 9 -keys alice,bob,carol
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"example.com/ch13-outbox/bus"
)

func main() {
	topic := flag.String("topic", "demo", "topic to produce to")
	n := flag.Int("n", 9, "number of messages")
	keys := flag.String("keys", "alice,bob,carol", "comma-separated keys, used round-robin")
	flag.Parse()

	brokers := bus.Brokers()
	if brokers == nil {
		fmt.Fprintln(os.Stderr, "produce: set KAFKA_BROKERS, for example localhost:19092")
		os.Exit(2)
	}
	cl, err := bus.NewProducer(brokers)
	if err != nil {
		fmt.Fprintln(os.Stderr, "produce:", err)
		os.Exit(1)
	}
	defer cl.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ks := strings.Split(*keys, ",")
	for i := range *n {
		key := ks[i%len(ks)]
		rec := &kgo.Record{Topic: *topic, Key: []byte(key), Value: fmt.Appendf(nil, "%s-%d", key, i)}
		// ProduceSync waits for the broker's acknowledgement (acks=all) before returning.
		if err := cl.ProduceSync(ctx, rec).FirstErr(); err != nil {
			fmt.Fprintln(os.Stderr, "produce:", err)
			os.Exit(1)
		}
		fmt.Printf("key=%-6s partition=%d offset=%d value=%s\n", key, rec.Partition, rec.Offset, rec.Value)
	}
}
