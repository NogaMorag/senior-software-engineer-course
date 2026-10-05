// Package bus builds franz-go clients with the settings chapter 13 explains, and adapts a producer to
// outbox.Publisher.
package bus

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"example.com/hookline/internal/outbox"
)

// ParseBrokers splits "host:port,host:port"; empty means "not configured".
func ParseBrokers(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}

// NewProducer returns a client for producing. franz-go's defaults are already the safe ones; they are
// spelled out here so that you can see them:
//   - acks=all: the leader answers only after every in-sync replica has the record;
//   - idempotent writes (on by default): the broker drops a retried batch it already wrote, so client
//     retries add no duplicates and keep per-partition order;
//   - the default partitioner hashes the key, so one tenant always lands on the same partition while
//     the partition count stays the same.
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

// NewGroupConsumer returns a consumer-group member for topic. Offsets are committed by the caller after
// processing (dispatch.Run), never automatically:
//   - DisableAutoCommit: no background commit of records you have not finished;
//   - BlockRebalanceOnPoll: a rebalance waits until the caller calls AllowRebalance, so a commit never
//     lands on a partition that has moved to another member;
//   - CooperativeStickyBalancer: in a rebalance only the partitions that move are revoked;
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

// KafkaPublisher publishes outbox messages to Topic with a producer client.
type KafkaPublisher struct {
	Client *kgo.Client
	Topic  string // empty means outbox.Topic
}

// Publish produces every message, keyed by tenant, and waits for all acknowledgements. Records go out
// in slice order; with idempotent writes, records with the same key keep that order.
func (p KafkaPublisher) Publish(ctx context.Context, msgs []outbox.Message) error {
	topic := p.Topic
	if topic == "" {
		topic = outbox.Topic
	}
	recs := make([]*kgo.Record, len(msgs))
	for i, m := range msgs {
		recs[i] = &kgo.Record{
			Topic:   topic,
			Key:     []byte(m.TenantID),
			Value:   m.Value(),
			Headers: []kgo.RecordHeader{{Key: outbox.HeaderEventID, Value: []byte(m.EventID)}},
			// The ingest request's trace, so kotel's publish span (when the producer has its hooks)
			// joins it and writes traceparent into the headers for the dispatcher (milestone M8).
			Context: withTraceparent(ctx, m.TraceContext),
		}
	}
	return p.Client.ProduceSync(ctx, recs...).FirstErr()
}

// withTraceparent returns ctx with the remote span context of traceparent, or ctx unchanged.
func withTraceparent(ctx context.Context, traceparent string) context.Context {
	if traceparent == "" {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier{"traceparent": traceparent})
}

// CreateTopic creates topic with the given partition count (replication factor 1, the only value a
// single broker allows). A topic that exists already is not an error: the call is safe to repeat.
func CreateTopic(ctx context.Context, brokers []string, topic string, partitions int32) error {
	cl, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		return err
	}
	defer cl.Close()
	req := kmsg.NewPtrCreateTopicsRequest()
	rt := kmsg.NewCreateTopicsRequestTopic()
	rt.Topic, rt.NumPartitions, rt.ReplicationFactor = topic, partitions, 1
	req.Topics = append(req.Topics, rt)
	resp, err := req.RequestWith(ctx, cl)
	if err != nil {
		return fmt.Errorf("create topic %s: %w", topic, err)
	}
	if err := kerr.ErrorForCode(resp.Topics[0].ErrorCode); err != nil && !errors.Is(err, kerr.TopicAlreadyExists) {
		return fmt.Errorf("create topic %s: %w", topic, err)
	}
	return nil
}

// Tracing returns the kotel hooks for a client and the tracer behind them (chapter 24): a publish span
// per produced record, whose context is injected into the record headers, and a receive span per
// fetched record, whose parent is extracted from them. It uses the global providers, so call it after
// telemetry.Start; without a collector they are no-ops.
func Tracing() (kgo.Opt, *kotel.Tracer) {
	tracer := kotel.NewTracer()
	k := kotel.NewKotel(kotel.WithTracer(tracer))
	return kgo.WithHooks(k.Hooks()...), tracer
}
