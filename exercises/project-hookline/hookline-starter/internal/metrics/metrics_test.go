package metrics_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"example.com/hookline/internal/metrics"
)

func collect(t *testing.T, r *sdkmetric.ManualReader) map[string]metricdata.Aggregation {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := r.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	out := map[string]metricdata.Aggregation{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			out[m.Name] = m.Data
		}
	}
	return out
}

func sum(t *testing.T, data metricdata.Aggregation) map[string]int64 {
	t.Helper()
	s, ok := data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("not an int64 sum: %T", data)
	}
	out := map[string]int64{}
	for _, dp := range s.DataPoints {
		key := ""
		for _, kv := range dp.Attributes.ToSlice() {
			key += string(kv.Key) + "=" + kv.Value.AsString() + " "
		}
		out[key] += dp.Value
	}
	return out
}

func TestInstruments(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)).Meter(metrics.Scope)
	m, err := metrics.NewWith(meter)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now()
	m.Delivered(ctx, now.Add(-3*time.Second), now)
	m.Delivered(ctx, now.Add(-45*time.Second), now)
	m.Attempt(ctx, "retry")
	m.Attempt(ctx, "dead")
	m.Postponed(ctx, "breaker")
	m.BreakerTransition(ctx, "closed", "open")
	m.Ingested(ctx, false)
	m.Ingested(ctx, true)
	m.LimiterError(ctx)
	m.Fanout(ctx, 0) // nothing new: no point recorded
	m.Fanout(ctx, 2)
	if err := metrics.ObserveOutbox(meter, func(context.Context) (int64, time.Duration, error) { return 7, 90 * time.Second, nil }); err != nil {
		t.Fatal(err)
	}

	got := collect(t, reader)
	lag, ok := got["hookline.delivery.lag"].(metricdata.Histogram[float64])
	if !ok || len(lag.DataPoints) != 1 {
		t.Fatalf("hookline.delivery.lag = %#v", got["hookline.delivery.lag"])
	}
	dp := lag.DataPoints[0]
	if dp.Count != 2 || !slices.Contains(dp.Bounds, 30) {
		t.Errorf("lag: count %d, bounds %v; want 2 samples and a bound at 30 s (the SLO threshold)", dp.Count, dp.Bounds)
	}
	within30 := uint64(0)
	for i, b := range dp.Bounds {
		if b <= 30 {
			within30 += dp.BucketCounts[i]
		}
	}
	if within30 != 1 {
		t.Errorf("deliveries within 30 s = %d, want 1 (3 s yes, 45 s no)", within30)
	}
	for name, want := range map[string]map[string]int64{
		"hookline.delivery.attempts":   {"outcome=delivered ": 2, "outcome=retry ": 1, "outcome=dead ": 1},
		"hookline.delivery.postponed":  {"reason=breaker ": 1},
		"hookline.breaker.transitions": {"from=closed to=open ": 1},
		"hookline.ingest.events":       {"result=accepted ": 1, "result=replayed ": 1},
		"hookline.ratelimit.errors":    {"": 1},
		"hookline.dispatch.deliveries": {"": 2},
	} {
		if s := sum(t, got[name]); !mapsEqual(s, want) {
			t.Errorf("%s = %v, want %v", name, s, want)
		}
	}
	if g, ok := got["hookline.outbox.pending"].(metricdata.Gauge[int64]); !ok || g.DataPoints[0].Value != 7 {
		t.Errorf("hookline.outbox.pending = %#v", got["hookline.outbox.pending"])
	}
	if g, ok := got["hookline.outbox.oldest_age"].(metricdata.Gauge[float64]); !ok || g.DataPoints[0].Value != 90 {
		t.Errorf("hookline.outbox.oldest_age = %#v", got["hookline.outbox.oldest_age"])
	}
}

// A failing database read skips the collection: no gauge is better than a wrong zero.
func TestObserveOutboxSkipsOnError(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)).Meter(metrics.Scope)
	if err := metrics.ObserveOutbox(meter, func(context.Context) (int64, time.Duration, error) { return 0, 0, errors.New("db down") }); err != nil {
		t.Fatal(err)
	}
	var rm metricdata.ResourceMetrics
	_ = reader.Collect(context.Background(), &rm) // the callback error is reported here; the gauge is absent
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if g, ok := m.Data.(metricdata.Gauge[int64]); ok && len(g.DataPoints) > 0 {
				t.Fatalf("%s reported %v despite the error", m.Name, g.DataPoints)
			}
		}
	}
}

func mapsEqual(a, b map[string]int64) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
