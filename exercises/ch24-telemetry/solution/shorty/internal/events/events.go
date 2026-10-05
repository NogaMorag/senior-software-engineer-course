// Package events is the Kafka hop of chapter 24: shorty publishes one record
// per created link, and linkworker consumes it. The trace context travels in
// the record headers (traceparent, W3C Trace Context), written and read by
// franz-go's OpenTelemetry plugin, kotel.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
)

// LinkCreated is the value of every record on the topic, as JSON.
type LinkCreated struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// Hooks returns the kotel hooks: a "publish" span per produced record, whose
// context is injected into the record headers, and a "receive" span per
// fetched record, whose parent is extracted from them. It uses the global
// tracer provider and propagator, so call it after telemetry.Setup.
func Hooks() (kgo.Opt, *kotel.Tracer) {
	tracer := kotel.NewTracer()
	k := kotel.NewKotel(kotel.WithTracer(tracer))
	return kgo.WithHooks(k.Hooks()...), tracer
}

// Producer publishes LinkCreated records to one topic.
type Producer struct {
	client  *kgo.Client
	topic   string
	timeout time.Duration
}

// NewProducer connects to brokers (a list of host:port). The topic must
// already exist: the client never asks the broker to create it, so a typo in
// the topic name fails every publish instead of creating a stray topic.
func NewProducer(brokers []string, topic string, timeout time.Duration) (*Producer, error) {
	hooks, _ := Hooks()
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.DefaultProduceTopic(topic),
		hooks,
	)
	if err != nil {
		return nil, err
	}
	return &Producer{client: client, topic: topic, timeout: timeout}, nil
}

// LinkCreated publishes ev and waits for the broker to acknowledge it. ctx
// carries the caller's span; kotel makes the publish span its child.
func (p *Producer) LinkCreated(ctx context.Context, ev LinkCreated) error {
	value, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	rec := &kgo.Record{Key: []byte(ev.Code), Value: value, Context: ctx}
	if err := p.client.ProduceSync(ctx, rec).FirstErr(); err != nil {
		return fmt.Errorf("publish to %s: %w", p.topic, err)
	}
	return nil
}

// Close flushes nothing (every publish is synchronous) and closes the client.
func (p *Producer) Close() { p.client.Close() }
