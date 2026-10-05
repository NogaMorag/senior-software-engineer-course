package events_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"example.com/ch24-telemetry/internal/events"
)

// TestTraceContextThroughRedpanda publishes through a real broker and checks
// that the consumer side continues the producer's trace. It needs the
// chapter 24 stack: KAFKA_BROKERS=localhost:19092 (see the README).
func TestTraceContextThroughRedpanda(t *testing.T) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		t.Skip("set KAFKA_BROKERS=localhost:19092 with the chapter 24 stack up; see README")
	}
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	otel.SetTextMapPropagator(propagation.TraceContext{})
	seeds := strings.Split(brokers, ",")
	topic := fmt.Sprintf("events-test-%d", time.Now().UnixNano())
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// The producer never creates topics; this setup client may.
	setup, err := kgo.NewClient(kgo.SeedBrokers(seeds...), kgo.AllowAutoTopicCreation())
	if err != nil {
		t.Fatal(err)
	}
	defer setup.Close()
	if err := setup.ProduceSync(ctx, &kgo.Record{Topic: topic, Value: []byte("{}")}).FirstErr(); err != nil {
		t.Fatalf("create topic %s: %v", topic, err)
	}

	producer, err := events.NewProducer(seeds, topic, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer producer.Close()
	pctx, parent := otel.Tracer("test").Start(ctx, "POST /v1/links")
	if err := producer.LinkCreated(pctx, events.LinkCreated{Code: "abc", URL: "https://go.dev/"}); err != nil {
		t.Fatalf("LinkCreated: %v", err)
	}
	parent.End()

	hooks, _ := events.Hooks()
	consumer, err := kgo.NewClient(kgo.SeedBrokers(seeds...), kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()), hooks)
	if err != nil {
		t.Fatal(err)
	}
	defer consumer.Close()
	for ctx.Err() == nil {
		fetches := consumer.PollFetches(ctx)
		for _, r := range fetches.Records() {
			if string(r.Key) != "abc" {
				continue // the setup record
			}
			got := trace.SpanContextFromContext(r.Context).TraceID()
			if got != parent.SpanContext().TraceID() {
				t.Fatalf("consumed record is in trace %s, want %s", got, parent.SpanContext().TraceID())
			}
			return
		}
	}
	t.Fatal("record not consumed before the deadline")
}
