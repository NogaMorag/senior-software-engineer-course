"""OpenTelemetry for the receiver: traces, metrics and logs over OTLP/HTTP. Added for chapter 24.

The exporters read the standard variables (OTEL_EXPORTER_OTLP_ENDPOINT, default
http://localhost:4318; OTEL_METRIC_EXPORT_INTERVAL in milliseconds; OTEL_RESOURCE_ATTRIBUTES).
Providers are passed explicitly instead of installed globally: the Python API lets a process set
the global tracer provider only once, and explicit providers let tests use in-memory exporters.
"""

import logging
import os
import socket
from dataclasses import dataclass
from typing import override

from fastapi import FastAPI
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http._log_exporter import OTLPLogExporter
from opentelemetry.exporter.otlp.proto.http.metric_exporter import OTLPMetricExporter
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.logging.handler import LoggingHandler
from opentelemetry.sdk._logs import LoggerProvider
from opentelemetry.sdk._logs.export import BatchLogRecordProcessor, LogRecordExporter
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import MetricReader, PeriodicExportingMetricReader
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import SpanProcessor, TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

LOG_FORMAT = (
    "%(asctime)s %(levelname)s %(name)s trace_id=%(trace_id)s span_id=%(span_id)s %(message)s"
)


@dataclass(slots=True)
class Telemetry:
    tracer_provider: TracerProvider
    meter_provider: MeterProvider
    logger_provider: LoggerProvider

    def shutdown(self) -> None:
        """Flush and stop. Logs last, so lines written during shutdown still leave."""
        self.tracer_provider.shutdown()
        self.meter_provider.shutdown()
        self.logger_provider.shutdown()


def setup_telemetry(
    service_name: str = "receiver",
    *,
    span_processor: SpanProcessor | None = None,
    metric_reader: MetricReader | None = None,
    log_exporter: LogRecordExporter | None = None,
) -> Telemetry:
    """Build the three providers. The keyword arguments replace the OTLP exporters (tests)."""
    resource = Resource.create(
        {"service.name": service_name, "service.instance.id": socket.gethostname()}
    )  # Resource.create also reads OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME

    tracer_provider = TracerProvider(resource=resource)
    tracer_provider.add_span_processor(span_processor or BatchSpanProcessor(OTLPSpanExporter()))

    reader = metric_reader or PeriodicExportingMetricReader(OTLPMetricExporter())
    meter_provider = MeterProvider(resource=resource, metric_readers=[reader])

    logger_provider = LoggerProvider(resource=resource)
    logger_provider.add_log_record_processor(
        BatchLogRecordProcessor(log_exporter or OTLPLogExporter())
    )
    return Telemetry(tracer_provider, meter_provider, logger_provider)


class TraceContextFilter(logging.Filter):
    """Adds trace_id and span_id of the current span to every record (empty strings outside one)."""

    @override
    def filter(self, record: logging.LogRecord) -> bool:
        context = trace.get_current_span().get_span_context()
        record.trace_id = format(context.trace_id, "032x") if context.is_valid else ""
        record.span_id = format(context.span_id, "016x") if context.is_valid else ""
        return True


def instrument(app: FastAPI, telemetry: Telemetry) -> None:
    """Server spans and http.server.request.duration for every route except /stats and /control."""
    # The Python instrumentations still emit the old HTTP conventions (http.server.duration in
    # milliseconds) unless asked for the stable ones, which the Go services and the dashboard use.
    os.environ.setdefault("OTEL_SEMCONV_STABILITY_OPT_IN", "http")
    FastAPIInstrumentor.instrument_app(
        app,
        tracer_provider=telemetry.tracer_provider,
        meter_provider=telemetry.meter_provider,
        excluded_urls="stats,control,receipts,reset",
        exclude_spans=["receive", "send"],  # no extra spans per ASGI message
    )


def otlp_handler(telemetry: Telemetry, level: int = logging.INFO) -> logging.Handler:
    """A logging handler that turns records into OTLP log records, with the current trace."""
    return LoggingHandler(level=level, logger_provider=telemetry.logger_provider)


def configure_logging(telemetry: Telemetry, level: int = logging.INFO) -> None:
    """Root logging: text lines with trace and span IDs to stderr, and every record to OTLP."""
    stream = logging.StreamHandler()
    stream.setFormatter(logging.Formatter(LOG_FORMAT))
    stream.addFilter(TraceContextFilter())
    # The OTLP handler takes the trace context from the current span by itself.
    logging.basicConfig(level=level, handlers=[stream, otlp_handler(telemetry, level)], force=True)
