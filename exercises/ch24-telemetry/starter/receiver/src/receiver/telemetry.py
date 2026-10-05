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
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.logging.handler import LoggingHandler
from opentelemetry.sdk._logs import LoggerProvider
from opentelemetry.sdk._logs.export import LogRecordExporter
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import MetricReader
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import SpanProcessor, TracerProvider

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
    # TODO(chapter 24, section 24.3), day 72, with `resource`:
    #  1. TracerProvider(resource=...), add_span_processor(span_processor or
    #     BatchSpanProcessor(OTLPSpanExporter())).
    #  2. MeterProvider(resource=..., metric_readers=[metric_reader or
    #     PeriodicExportingMetricReader(OTLPMetricExporter())]).
    #  3. LoggerProvider(resource=...), add_log_record_processor(
    #     BatchLogRecordProcessor(log_exporter or OTLPLogExporter())).
    #  4. Return Telemetry(tracer_provider, meter_provider, logger_provider).
    # The exporters live in opentelemetry.exporter.otlp.proto.http (trace_exporter,
    # metric_exporter, _log_exporter); the readers and processors in opentelemetry.sdk.
    _ = resource
    raise NotImplementedError("TODO(chapter 24, section 24.3): build the three providers")


class TraceContextFilter(logging.Filter):
    """Adds trace_id and span_id of the current span to every record (empty strings outside one)."""

    @override
    def filter(self, record: logging.LogRecord) -> bool:
        # TODO(chapter 24, section 24.3): take trace.get_current_span().get_span_context();
        # when it is valid, set record.trace_id to format(trace_id, "032x") and
        # record.span_id to format(span_id, "016x"); otherwise set both to "".
        record.trace_id = record.span_id = "TODO"
        _ = trace
        return True


def instrument(app: FastAPI, telemetry: Telemetry) -> None:
    """Server spans and http.server.request.duration for every route except /stats and /control."""
    # The Python instrumentations still emit the old HTTP conventions (http.server.duration in
    # milliseconds) unless asked for the stable ones, which the Go services and the dashboard use.
    os.environ.setdefault("OTEL_SEMCONV_STABILITY_OPT_IN", "http")
    # TODO(chapter 24, section 24.3): FastAPIInstrumentor.instrument_app(app, ...) with the
    # tracer and meter providers of `telemetry` and excluded_urls="stats,control,receipts,reset"
    # and exclude_spans=["receive", "send"].
    _ = (app, telemetry, FastAPIInstrumentor)


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
