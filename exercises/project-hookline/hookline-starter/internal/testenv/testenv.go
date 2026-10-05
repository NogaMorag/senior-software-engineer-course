// Package testenv gives integration tests the rest of the compose stack: Kafka brokers and private
// topics, the Redis address, and the sink's URL. Each helper skips the test when its part of the stack
// is not configured. It is test support, kept whole in the starter.
package testenv

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"

	"example.com/hookline/internal/bus"
)

var counter atomic.Int64

// Unique returns a name nobody else uses, even across parallel test binaries.
func Unique(prefix string) string {
	return fmt.Sprintf("%s-%d-%d-%d", prefix, os.Getpid(), time.Now().UnixNano(), counter.Add(1))
}

func need(t testing.TB, name string) string {
	t.Helper()
	v := os.Getenv(name)
	if v == "" {
		t.Skipf("integration test: set %s to the Hookline compose stack (make up); see README", name)
	}
	return v
}

// Brokers returns KAFKA_BROKERS or skips the test.
func Brokers(t testing.TB) []string {
	t.Helper()
	return bus.ParseBrokers(need(t, "KAFKA_BROKERS"))
}

// RedisAddr returns REDIS_ADDR or skips the test.
func RedisAddr(t testing.TB) string {
	t.Helper()
	return need(t, "REDIS_ADDR")
}

// SinkURL returns SINK_URL (for example http://localhost:8091) or skips the test.
func SinkURL(t testing.TB) string {
	t.Helper()
	return need(t, "SINK_URL")
}

// Topic creates a fresh topic with the given partition count (replication factor 1, the only value a
// single broker allows) and deletes it when the test ends.
func Topic(t testing.TB, brokers []string, partitions int32) string {
	t.Helper()
	name := Unique("hookline-test")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := bus.CreateTopic(ctx, brokers, name, partitions); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		admin(t, brokers, func(ctx context.Context, cl *kgo.Client) error {
			del := kmsg.NewPtrDeleteTopicsRequest()
			dt := kmsg.NewDeleteTopicsRequestTopic()
			dt.Topic = kmsg.StringPtr(name)
			del.Topics = append(del.Topics, dt)
			del.TopicNames = []string{name}
			_, err := del.RequestWith(ctx, cl)
			return err
		})
	})
	return name
}

func admin(t testing.TB, brokers []string, fn func(context.Context, *kgo.Client) error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cl, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()
	if err := fn(ctx, cl); err != nil {
		t.Fatalf("kafka admin request: %v", err)
	}
}
