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
	"errors"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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

	traceExp, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExp), // spans leave in batches, off the request path
	)

	metricExp, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, errors.Join(err, tp.Shutdown(ctx))
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
	)

	logExp, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, errors.Join(err, tp.Shutdown(ctx), mp.Shutdown(ctx))
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetLoggerProvider(lp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // the traceparent and tracestate headers
		propagation.Baggage{},
	))

	// Go runtime metrics: memory, goroutines, GC. The raw material for USE.
	if err := runtime.Start(); err != nil {
		return nil, errors.Join(err, tp.Shutdown(ctx), mp.Shutdown(ctx), lp.Shutdown(ctx))
	}

	return func(ctx context.Context) error {
		// Logs last, so log lines written while the others shut down still leave.
		return errors.Join(tp.Shutdown(ctx), mp.Shutdown(ctx), lp.Shutdown(ctx))
	}, nil
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
