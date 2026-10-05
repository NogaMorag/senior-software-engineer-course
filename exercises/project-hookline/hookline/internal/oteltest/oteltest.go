// Package oteltest installs recording OpenTelemetry providers for one test and restores the previous
// globals afterwards. Test support, kept whole in the starter.
package oteltest

import (
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// Traces makes every span of the test land in the returned recorder and installs the W3C propagator.
// Tests that call it must not run in parallel: the providers are process globals.
func Traces(t testing.TB) *tracetest.SpanRecorder {
	t.Helper()
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	prevTP, prevProp := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		_ = tp.Shutdown(t.Context())
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})
	return rec
}

// Traceparent is a sampled W3C traceparent for traceID (32 hex digits) and spanID (16 hex digits).
func Traceparent(traceID, spanID string) string { return "00-" + traceID + "-" + spanID + "-01" }
