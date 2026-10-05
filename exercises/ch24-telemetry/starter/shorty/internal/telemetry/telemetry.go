// Package telemetry sets up OpenTelemetry for one process: traces, metrics and
// logs, each exported over OTLP/HTTP to the collector. Added for chapter 24.
//
// The exporters read the standard environment variables, so the code holds no
// addresses: OTEL_EXPORTER_OTLP_ENDPOINT (default http://localhost:4318),
// OTEL_METRIC_EXPORT_INTERVAL (milliseconds, default 60000) and
// OTEL_RESOURCE_ATTRIBUTES.
package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// Shutdown flushes and stops the providers. Call it once, on the way out, with
// a context that bounds how long the flush may take.
type Shutdown func(context.Context) error

// Setup builds the three providers for service and installs them as the
// global providers, together with the W3C trace-context propagator. Every
// library that uses the globals (otelhttp, kotel, otelslog) picks them up, so
// Setup must run before those libraries are constructed.
func Setup(ctx context.Context, service, version string) (Shutdown, error) {
	res, err := newResource(ctx, service, version)
	if err != nil {
		return nil, err
	}
	_ = res
	// TODO(chapter 24, section 24.2), day 72:
	//  1. otlptracehttp.New(ctx), then sdktrace.NewTracerProvider with the
	//     resource and a batcher (sdktrace.WithBatcher).
	//  2. otlpmetrichttp.New(ctx), then sdkmetric.NewMeterProvider with the
	//     resource and a sdkmetric.NewPeriodicReader.
	//  3. otlploghttp.New(ctx), then sdklog.NewLoggerProvider with the resource
	//     and a sdklog.NewBatchProcessor.
	//  4. otel.SetTracerProvider, otel.SetMeterProvider, otel.SetLoggerProvider,
	//     and otel.SetTextMapPropagator with TraceContext and Baggage.
	//  5. runtime.Start() for the Go runtime metrics.
	//  6. Return a Shutdown that shuts all three providers down (errors.Join).
	panic("TODO(chapter 24, section 24.2): build and install the three providers")
}

// newResource describes this process. service.name becomes the Prometheus
// job label and the Loki and Tempo service_name; service.instance.id becomes
// the instance label, so two copies of one service never write the same series.
func newResource(ctx context.Context, service, version string) (*resource.Resource, error) {
	host, _ := os.Hostname() // a container's hostname is its ID
	return resource.New(ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceName(service),
			semconv.ServiceVersion(version),
			semconv.ServiceInstanceID(host),
		),
		resource.WithFromEnv(), // OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME win
		resource.WithTelemetrySDK(),
	)
}
