package bus_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"example.com/hookline/internal/bus"
	"example.com/hookline/internal/outbox"
	"example.com/hookline/internal/testenv"
)

func TestParseBrokers(t *testing.T) {
	if got := bus.ParseBrokers(" a:1,b:2 "); !slices.Equal(got, []string{"a:1", "b:2"}) {
		t.Errorf("ParseBrokers = %v", got)
	}
	if bus.ParseBrokers("") != nil {
		t.Error("an empty value must mean not configured")
	}
}

// TestPublishKeysByTenantInOrder: one tenant's events land on one partition, in publish order, with the
// event id in the header.
func TestPublishKeysByTenantInOrder(t *testing.T) {
	brokers := testenv.Brokers(t)
	topic := testenv.Topic(t, brokers, 3)
	if err := bus.CreateTopic(context.Background(), brokers, topic, 3); err != nil {
		t.Fatalf("CreateTopic on an existing topic must succeed: %v", err)
	}
	producer, err := bus.NewProducer(brokers)
	if err != nil {
		t.Fatal(err)
	}
	defer producer.Close()
	var msgs []outbox.Message
	for i := range 30 {
		tenant := []string{"t-a", "t-b", "t-c"}[i%3]
		msgs = append(msgs, outbox.Message{EventID: testenv.Unique("ev"), TenantID: tenant, Type: "a.b"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := (bus.KafkaPublisher{Client: producer, Topic: topic}).Publish(ctx, msgs); err != nil {
		t.Fatal(err)
	}

	cl, err := kgo.NewClient(kgo.SeedBrokers(brokers...), kgo.ConsumeTopics(topic), kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()
	partition := map[string]int32{}
	order := map[string][]string{}
	for seen := 0; seen < len(msgs); {
		fetches := cl.PollFetches(ctx)
		if ctx.Err() != nil {
			t.Fatalf("read %d of %d records", seen, len(msgs))
		}
		fetches.EachRecord(func(r *kgo.Record) {
			seen++
			tenant := string(r.Key)
			if p, ok := partition[tenant]; ok && p != r.Partition {
				t.Errorf("tenant %s on partitions %d and %d", tenant, p, r.Partition)
			}
			partition[tenant] = r.Partition
			if len(r.Headers) != 1 || r.Headers[0].Key != outbox.HeaderEventID {
				t.Errorf("headers = %v", r.Headers)
			}
			order[tenant] = append(order[tenant], string(r.Headers[0].Value))
		})
	}
	for tenant, got := range order {
		var want []string
		for _, m := range msgs {
			if m.TenantID == tenant {
				want = append(want, m.EventID)
			}
		}
		if !slices.Equal(got, want) {
			t.Errorf("tenant %s read %v, want publish order %v", tenant, got, want)
		}
	}
}
