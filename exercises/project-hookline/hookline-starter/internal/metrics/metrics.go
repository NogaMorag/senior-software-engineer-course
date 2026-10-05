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

	"time"

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
	panic("TODO: see chapter 38")

}

// NewWith creates the instruments on meter (tests pass a meter with a manual reader).
func NewWith(meter metric.Meter) (*Metrics, error) {
	panic("TODO: see chapter 38")

}

// Must is New, panicking on error: instrument creation fails only on an invalid name, a programming error.
func Must() *Metrics {
	panic("TODO: see chapter 38")

}

// Delivered records one delivery that got a 2xx, and its lag from the event's ingest.
func (m *Metrics) Delivered(ctx context.Context, eventCreatedAt, now time.Time) {
	panic("TODO: see chapter 38")

}

// Attempt counts one recorded attempt; outcome is delivered, retry or dead.
func (m *Metrics) Attempt(ctx context.Context, outcome string) {
	panic("TODO: see chapter 38")

}

// Postponed counts a delivery given back without an attempt; reason is breaker or bulkhead.
func (m *Metrics) Postponed(ctx context.Context, reason string) {
	panic("TODO: see chapter 38")

}

// BreakerTransition counts one breaker state change. The endpoint is in the log line, not in a label:
// one series per endpoint would grow without bound.
func (m *Metrics) BreakerTransition(ctx context.Context, from, to string) {
	panic("TODO: see chapter 38")

}

// Ingested counts one accepted ingest request.
func (m *Metrics) Ingested(ctx context.Context, replayed bool) {
	panic("TODO: see chapter 38")

}

// LimiterError counts one limiter failure served open.
func (m *Metrics) LimiterError(ctx context.Context) { panic("TODO: see chapter 38") }

// Fanout counts delivery rows the dispatcher created.
func (m *Metrics) Fanout(ctx context.Context, n int) {
	panic("TODO: see chapter 38")

}

// OutboxStats reads the outbox: how many rows wait, and how old the oldest is.
type OutboxStats func(ctx context.Context) (pending int64, oldest time.Duration, err error)

// ObserveOutbox registers the two outbox gauges, read by stats at every collection (the relay's
// export interval). An error skips that collection rather than reporting a wrong zero.
func ObserveOutbox(meter metric.Meter, stats OutboxStats) error {
	panic("TODO: see chapter 38")

}
