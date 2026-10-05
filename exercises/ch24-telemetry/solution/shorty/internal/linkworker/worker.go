// Package linkworker consumes the links.created topic and announces every new
// link to a webhook receiver (the chapter 8 receiver), signed. It is the far
// side of the chapter 24 Kafka hop: the trace that started in shorty's
// handler continues here and then into the receiver.
package linkworker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"

	"example.com/ch24-telemetry/internal/events"
)

// Worker handles one record at a time. Its three instruments are USE for the
// worker itself, the resource of section 24.5:
//
//	linkworker.busy        seconds spent handling records (utilization: its rate)
//	linkworker.wait        seconds a record waited between produce and handling (saturation)
//	linkworker.deliveries  handled records by outcome (errors)
type Worker struct {
	tracer      *kotel.Tracer
	http        *http.Client
	receiverURL string
	secret      []byte
	log         *slog.Logger
	now         func() time.Time

	busy       metric.Float64Counter
	wait       metric.Float64Histogram
	deliveries metric.Int64Counter
}

// New builds a worker that posts to receiverURL, signing with secret. tracer
// must be the one whose hooks are installed on the consuming client (see
// events.Hooks), so the process span continues the receive span.
func New(tracer *kotel.Tracer, receiverURL, secret string, log *slog.Logger) (*Worker, error) {
	meter := otel.Meter("example.com/ch24-telemetry/linkworker")
	w := &Worker{
		tracer:      tracer,
		receiverURL: receiverURL,
		secret:      []byte(secret),
		log:         log,
		now:         time.Now,
		// otelhttp's transport makes a client span per request and injects
		// traceparent into the request headers: the hop into the receiver.
		http: &http.Client{Timeout: 5 * time.Second, Transport: otelhttp.NewTransport(http.DefaultTransport)},
	}
	var err error
	if w.busy, err = meter.Float64Counter("linkworker.busy", metric.WithUnit("s"),
		metric.WithDescription("Time spent handling records.")); err != nil {
		return nil, err
	}
	if w.wait, err = meter.Float64Histogram("linkworker.wait", metric.WithUnit("s"),
		metric.WithDescription("Time between producing a record and starting to handle it."),
		metric.WithExplicitBucketBoundaries(0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30, 60)); err != nil {
		return nil, err
	}
	if w.deliveries, err = meter.Int64Counter("linkworker.deliveries",
		metric.WithDescription("Handled records by outcome.")); err != nil {
		return nil, err
	}
	return w, nil
}

// Run polls client until ctx is cancelled and handles every record. A failed
// delivery is logged and counted, not retried: retries are chapter 14's topic.
func (w *Worker) Run(ctx context.Context, client *kgo.Client) error {
	for {
		fetches := client.PollFetches(ctx)
		if ctx.Err() != nil {
			return nil
		}
		fetches.EachError(func(topic string, partition int32, err error) {
			w.log.ErrorContext(ctx, "fetch failed", slog.String("topic", topic),
				slog.Int("partition", int(partition)), slog.Any("error", err))
		})
		fetches.EachRecord(func(r *kgo.Record) { _ = w.Handle(r) })
	}
}

// Handle delivers one record. The process span is a child of the context
// that kotel extracted from the record headers, so it joins the producer's
// trace; everything below it (the log lines, the HTTP call) joins it too.
func (w *Worker) Handle(r *kgo.Record) error {
	start := w.now()
	ctx, span := w.tracer.WithProcessSpan(r)
	defer span.End()
	w.wait.Record(ctx, start.Sub(r.Timestamp).Seconds())
	defer func() { w.busy.Add(ctx, w.now().Sub(start).Seconds()) }()

	err := w.deliver(ctx, r)
	outcome := "ok"
	if err != nil {
		outcome = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, "delivery failed")
		w.log.ErrorContext(ctx, "delivery failed", slog.Int64("offset", r.Offset), slog.Any("error", err))
	} else {
		w.log.InfoContext(ctx, "link announced", slog.String("code", string(r.Key)))
	}
	w.deliveries.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", outcome)))
	return err
}

func (w *Worker) deliver(ctx context.Context, r *kgo.Record) error {
	var ev events.LinkCreated
	if err := json.Unmarshal(r.Value, &ev); err != nil {
		return fmt.Errorf("decode record: %w", err)
	}
	body, err := json.Marshal(map[string]any{"type": "link.created", "data": ev})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.receiverURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	// Topic, partition and offset name the record uniquely: a redelivered
	// record gets the same event ID, so the receiver can drop the duplicate.
	req.Header.Set("Hookline-Event-Id", fmt.Sprintf("%s-%d-%d", r.Topic, r.Partition, r.Offset))
	req.Header.Set("Hookline-Signature", Sign(w.secret, w.now().Unix(), body))

	resp, err := w.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("receiver answered %s", resp.Status)
	}
	return nil
}

// Sign is the chapter 8 signature: t=<unix seconds>,v1=<hex HMAC-SHA256 of
// "<t>." followed by the raw body>.
func Sign(secret []byte, unix int64, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	t := strconv.FormatInt(unix, 10)
	mac.Write([]byte(t + "."))
	mac.Write(body)
	return "t=" + t + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}
