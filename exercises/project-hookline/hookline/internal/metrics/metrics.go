// Package metrics holds Hookline's own instruments (milestone M8): what the generic HTTP and runtime
// metrics cannot say about a webhook pipeline. Each binary makes one Metrics from the global meter
// provider, after telemetry.Start; without a collector the provider is a no-op and every call is free.
//
// In Prometheus (OTLP translation) the instruments appear as:
//
//	hookline_delivery_lag_seconds          histogram: ingest to 2xx, per delivered delivery (SLO N2)
//	hookline_delivery_attempts_total       counter by outcome: delivered, retry, dead
//	hookline_delivery_postponed_total      counter by reason: breaker, bulkhead
//	hookline_breaker_transitions_total     counter by from, to
//	hookline_ingest_events_total           counter by result: accepted, replayed
//	hookline_ratelimit_errors_total        counter: limiter failures served open
//	hookline_dispatch_deliveries_total     counter: delivery rows created by the dispatcher
//	hookline_outbox_pending                gauge: unpublished outbox rows
//	hookline_outbox_oldest_age_seconds     gauge: age of the oldest unpublished row
package metrics

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Scope is the instrumentation scope of every Hookline instrument.
const Scope = "example.com/hookline"

// LagBuckets are the boundaries of the delivery lag histogram, in seconds. 30 is one of them because
// the delivery SLO (N2) counts deliveries within 30 s: an SLI threshold must be a bucket boundary.
var LagBuckets = []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 120, 300, 900, 3600}

// Metrics is safe for concurrent use.
type Metrics struct {
	lag         metric.Float64Histogram
	attempts    metric.Int64Counter
	postponed   metric.Int64Counter
	transitions metric.Int64Counter
	ingested    metric.Int64Counter
	limiterErrs metric.Int64Counter
	fanout      metric.Int64Counter
}

// New creates the instruments on the global meter provider.
func New() (*Metrics, error) {
	return NewWith(otel.GetMeterProvider().Meter(Scope))
}

// NewWith creates the instruments on meter (tests pass a meter with a manual reader).
func NewWith(meter metric.Meter) (*Metrics, error) {
	var m Metrics
	var errs []error
	add := func(err error) { errs = append(errs, err) }
	var err error
	m.lag, err = meter.Float64Histogram("hookline.delivery.lag", metric.WithUnit("s"),
		metric.WithDescription("Time from ingest to the receiver's 2xx, per delivered delivery."),
		metric.WithExplicitBucketBoundaries(LagBuckets...))
	add(err)
	m.attempts, err = meter.Int64Counter("hookline.delivery.attempts",
		metric.WithDescription("Delivery attempts by outcome: delivered, retry or dead."))
	add(err)
	m.postponed, err = meter.Int64Counter("hookline.delivery.postponed",
		metric.WithDescription("Deliveries postponed without an attempt, by reason: breaker or bulkhead."))
	add(err)
	m.transitions, err = meter.Int64Counter("hookline.breaker.transitions",
		metric.WithDescription("Per-endpoint circuit breaker state changes."))
	add(err)
	m.ingested, err = meter.Int64Counter("hookline.ingest.events",
		metric.WithDescription("Accepted ingest requests by result: accepted (new event) or replayed."))
	add(err)
	m.limiterErrs, err = meter.Int64Counter("hookline.ratelimit.errors",
		metric.WithDescription("Rate limiter failures; each request was served without a limit."))
	add(err)
	m.fanout, err = meter.Int64Counter("hookline.dispatch.deliveries",
		metric.WithDescription("Delivery rows the dispatcher created."))
	add(err)
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return &m, nil
}

// Must is New, panicking on error: instrument creation fails only on an invalid name, a programming error.
func Must() *Metrics {
	m, err := New()
	if err != nil {
		panic(err)
	}
	return m
}

// Delivered records one delivery that got a 2xx, and its lag from the event's ingest.
func (m *Metrics) Delivered(ctx context.Context, eventCreatedAt, now time.Time) {
	m.lag.Record(ctx, max(0, now.Sub(eventCreatedAt).Seconds()))
	m.Attempt(ctx, "delivered")
}

// Attempt counts one recorded attempt; outcome is delivered, retry or dead.
func (m *Metrics) Attempt(ctx context.Context, outcome string) {
	m.attempts.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", outcome)))
}

// Postponed counts a delivery given back without an attempt; reason is breaker or bulkhead.
func (m *Metrics) Postponed(ctx context.Context, reason string) {
	m.postponed.Add(ctx, 1, metric.WithAttributes(attribute.String("reason", reason)))
}

// BreakerTransition counts one breaker state change. The endpoint is in the log line, not in a label:
// one series per endpoint would grow without bound.
func (m *Metrics) BreakerTransition(ctx context.Context, from, to string) {
	m.transitions.Add(ctx, 1, metric.WithAttributes(attribute.String("from", from), attribute.String("to", to)))
}

// Ingested counts one accepted ingest request.
func (m *Metrics) Ingested(ctx context.Context, replayed bool) {
	result := "accepted"
	if replayed {
		result = "replayed"
	}
	m.ingested.Add(ctx, 1, metric.WithAttributes(attribute.String("result", result)))
}

// LimiterError counts one limiter failure served open.
func (m *Metrics) LimiterError(ctx context.Context) { m.limiterErrs.Add(ctx, 1) }

// Fanout counts delivery rows the dispatcher created.
func (m *Metrics) Fanout(ctx context.Context, n int) {
	if n > 0 {
		m.fanout.Add(ctx, int64(n))
	}
}

// OutboxStats reads the outbox: how many rows wait, and how old the oldest is.
type OutboxStats func(ctx context.Context) (pending int64, oldest time.Duration, err error)

// ObserveOutbox registers the two outbox gauges, read by stats at every collection (the relay's
// export interval). An error skips that collection rather than reporting a wrong zero.
func ObserveOutbox(meter metric.Meter, stats OutboxStats) error {
	pending, err := meter.Int64ObservableGauge("hookline.outbox.pending",
		metric.WithDescription("Outbox rows not yet published."))
	if err != nil {
		return err
	}
	age, err := meter.Float64ObservableGauge("hookline.outbox.oldest_age", metric.WithUnit("s"),
		metric.WithDescription("Age of the oldest unpublished outbox row; 0 when the outbox is empty."))
	if err != nil {
		return err
	}
	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		n, oldest, err := stats(ctx)
		if err != nil {
			return err
		}
		o.ObserveInt64(pending, n)
		o.ObserveFloat64(age, oldest.Seconds())
		return nil
	}, pending, age)
	return err
}
