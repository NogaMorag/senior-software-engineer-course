package telemetry_test

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"

	"example.com/hookline/internal/telemetry"
)

func TestStartWithoutACollectorIsANoOp(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	before := otel.GetTracerProvider()
	shutdown, err := telemetry.Start(context.Background(), "hookline-test")
	if err != nil || shutdown(context.Background()) != nil {
		t.Fatalf("Start = %v", err)
	}
	if telemetry.Enabled() {
		t.Error("Enabled with an empty endpoint")
	}
	if otel.GetTracerProvider() != before {
		t.Error("Start installed a tracer provider without a collector")
	}
}
