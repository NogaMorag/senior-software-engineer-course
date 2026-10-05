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
LOG_FORMAT = '%(asctime)s %(levelname)s %(name)s trace_id=%(trace_id)s span_id=%(span_id)s %(message)s'

@dataclass(slots=True)
class Telemetry:
    tracer_provider: TracerProvider
    meter_provider: MeterProvider
    logger_provider: LoggerProvider

    def shutdown(self) -> None:
        """Flush and stop. Logs last, so lines written during shutdown still leave."""
        raise NotImplementedError

def setup_telemetry(service_name: str='receiver', *, span_processor: SpanProcessor | None=None, metric_reader: MetricReader | None=None, log_exporter: LogRecordExporter | None=None) -> Telemetry:
    """Build the three providers. The keyword arguments replace the OTLP exporters (tests)."""
    raise NotImplementedError

class TraceContextFilter(logging.Filter):
    """Adds trace_id and span_id of the current span to every record (empty strings outside one)."""

    @override
    def filter(self, record: logging.LogRecord) -> bool:
        raise NotImplementedError

def instrument(app: FastAPI, telemetry: Telemetry) -> None:
    """Server spans and http.server.request.duration for every route except /stats and /control."""
    raise NotImplementedError

def otlp_handler(telemetry: Telemetry, level: int=logging.INFO) -> logging.Handler:
    """A logging handler that turns records into OTLP log records, with the current trace."""
    raise NotImplementedError

def configure_logging(telemetry: Telemetry, level: int=logging.INFO) -> None:
    """Root logging: text lines with trace and span IDs to stderr, and every record to OTLP."""
    raise NotImplementedError
