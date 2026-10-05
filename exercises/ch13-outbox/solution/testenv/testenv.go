// Package testenv gives tests a private PostgreSQL schema and private Kafka topics, and
// skips the test when the stack is not configured. It is test support, not exercise code.
package testenv

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"

	"example.com/ch13-outbox/bus"
	"example.com/ch13-outbox/db"
)

var counter atomic.Int64

// unique returns a name nobody else uses, even across parallel test binaries.
func unique(prefix string) string {
	return fmt.Sprintf("%s_%d_%d_%d", prefix, os.Getpid(), time.Now().UnixNano(), counter.Add(1))
}

// Postgres returns a pool whose search_path is a fresh schema with db.Schema applied.
// The schema is dropped when the test ends.
func Postgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("integration test: set DATABASE_URL to the chapter 13 database; see README")
	}
	ctx := context.Background()
	schema := unique("ch13_test")
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop schema: %v", err)
		}
	})
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := db.Connect(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(ctx, db.Schema); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	return pool
}

// Brokers returns KAFKA_BROKERS or skips the test.
func Brokers(t *testing.T) []string {
	t.Helper()
	brokers := bus.Brokers()
	if brokers == nil {
		t.Skip("integration test: set KAFKA_BROKERS to the chapter 13 Redpanda; see README")
	}
	return brokers
}

// Topic creates a fresh topic with the given partition count (replication factor 1, the
// only possible value on a single broker) and deletes it when the test ends.
func Topic(t *testing.T, brokers []string, partitions int32) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cl, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()

	name := unique("ch13-test")
	req := kmsg.NewPtrCreateTopicsRequest()
	rt := kmsg.NewCreateTopicsRequestTopic()
	rt.Topic, rt.NumPartitions, rt.ReplicationFactor = name, partitions, 1
	req.Topics = append(req.Topics, rt)
	resp, err := req.RequestWith(ctx, cl)
	if err != nil {
		t.Fatalf("create topic: %v", err)
	}
	if err := kerr.ErrorForCode(resp.Topics[0].ErrorCode); err != nil {
		t.Fatalf("create topic %s: %v", name, err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cl, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
		if err != nil {
			t.Errorf("delete topic: %v", err)
			return
		}
		defer cl.Close()
		del := kmsg.NewPtrDeleteTopicsRequest()
		dt := kmsg.NewDeleteTopicsRequestTopic()
		dt.Topic = kmsg.StringPtr(name)
		del.Topics = append(del.Topics, dt)
		del.TopicNames = []string{name}
		if _, err := del.RequestWith(ctx, cl); err != nil {
			t.Errorf("delete topic %s: %v", name, err)
		}
	})
	return name
}

// Group returns a fresh consumer-group name.
func Group(prefix string) string {
	return unique(prefix)
}
