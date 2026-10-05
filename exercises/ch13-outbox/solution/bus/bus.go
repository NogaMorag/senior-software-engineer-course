// Package bus builds franz-go clients with the settings the chapter explains, and adapts
// a producer to outbox.Publisher.
package bus

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"example.com/ch13-outbox/outbox"
)

// Brokers reads KAFKA_BROKERS ("host:port,host:port"); empty means "not configured".
func Brokers() []string {
	v := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}

// NewProducer returns a client for producing. franz-go's defaults are already the safe
// ones; they are spelled out here so that you can see them:
//   - acks=all: the leader answers only after every in-sync replica has the record;
//   - idempotent writes (on by default; kgo.DisableIdempotentWrite would turn them off):
//     the broker drops a retried batch it already wrote, so client retries add no
//     duplicates and keep per-partition order;
//   - the default partitioner hashes the key (murmur2, as the Java client does), so one
//     key always lands on the same partition while the partition count stays the same.
func NewProducer(brokers []string, opts ...kgo.Opt) (*kgo.Client, error) {
	base := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.ProducerLinger(5 * time.Millisecond),
	}
	cl, err := kgo.NewClient(append(base, opts...)...)
	if err != nil {
		return nil, fmt.Errorf("new producer: %w", err)
	}
	return cl, nil
}

// NewGroupConsumer returns a consumer-group member for topic. Offsets are committed by
// the caller after processing (consumer.Run), never automatically:
//   - DisableAutoCommit: no background commit of records you have not finished;
//   - BlockRebalanceOnPoll: a rebalance waits until the caller calls AllowRebalance, so
//     a commit never lands on a partition that has moved to another member;
//   - CooperativeStickyBalancer (the franz-go default, spelled out): in a rebalance only
//     the partitions that move are revoked, the others keep being consumed;
//   - a new group starts at the beginning of the topic, not at its end.
func NewGroupConsumer(brokers []string, group, topic string, opts ...kgo.Opt) (*kgo.Client, error) {
	base := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll(),
		kgo.Balancers(kgo.CooperativeStickyBalancer()),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	}
	cl, err := kgo.NewClient(append(base, opts...)...)
	if err != nil {
		return nil, fmt.Errorf("new consumer: %w", err)
	}
	return cl, nil
}

// KafkaPublisher publishes outbox messages with a producer client.
type KafkaPublisher struct {
	Client *kgo.Client
}

// Publish produces every message and waits for all acknowledgements. Records go out in
// slice order; with idempotent writes, records with the same key keep that order.
func (p KafkaPublisher) Publish(ctx context.Context, msgs []outbox.Message) error {
	recs := make([]*kgo.Record, len(msgs))
	for i, m := range msgs {
		recs[i] = &kgo.Record{
			Topic:   m.Topic,
			Key:     []byte(m.Key),
			Value:   m.Payload,
			Headers: []kgo.RecordHeader{{Key: outbox.HeaderMessageID, Value: []byte(m.ID)}},
		}
	}
	return p.Client.ProduceSync(ctx, recs...).FirstErr()
}
