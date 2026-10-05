"""Chapter 24: the receiver's spans, metrics and logs, captured in memory."""

import logging
from collections.abc import AsyncIterator, Iterator

import httpx
import pytest
from fastapi import FastAPI
from opentelemetry.sdk._logs.export import InMemoryLogRecordExporter, SimpleLogRecordProcessor
from opentelemetry.sdk.metrics.export import InMemoryMetricReader
from opentelemetry.sdk.trace.export import SimpleSpanProcessor
from opentelemetry.sdk.trace.export.in_memory_span_exporter import InMemorySpanExporter

from receiver.telemetry import (
    Telemetry,
    TraceContextFilter,
    instrument,
    otlp_handler,
    setup_telemetry,
)

from helpers import Headers

TRACE_ID = "4bf92f3577b34da6a3ce929d0e0e4736"
TRACEPARENT = f"00-{TRACE_ID}-00f067aa0ba902b7-01"  # what linkworker's HTTP client sends


class Captured:
    def __init__(self) -> None:
        self.spans = InMemorySpanExporter()
        self.metrics = InMemoryMetricReader()
        self.logs = InMemoryLogRecordExporter()  # type: ignore[no-untyped-call]
        self.telemetry: Telemetry = setup_telemetry(
            "receiver-test",
            span_processor=SimpleSpanProcessor(self.spans),
            metric_reader=self.metrics,
            log_exporter=self.logs,
        )
        # Synchronous log export, so the test sees records at once.
        self.telemetry.logger_provider.add_log_record_processor(SimpleLogRecordProcessor(self.logs))


@pytest.fixture
def captured() -> Iterator[Captured]:
    c = Captured()
    yield c
    c.telemetry.shutdown()


@pytest.fixture
async def traced_client(app: FastAPI, captured: Captured) -> AsyncIterator[httpx.AsyncClient]:
    instrument(app, captured.telemetry)
    handler = otlp_handler(captured.telemetry)
    logger = logging.getLogger("receiver")
    logger.addHandler(handler)
    previous = logger.level
    logger.setLevel(logging.INFO)
    try:
        async with app.router.lifespan_context(app):
            transport = httpx.ASGITransport(app=app)
            async with httpx.AsyncClient(transport=transport, base_url="http://test") as client:
                yield client
    finally:
        logger.removeHandler(handler)
        logger.setLevel(previous)


@pytest.mark.anyio
async def test_server_span_continues_the_callers_trace(
    traced_client: httpx.AsyncClient, captured: Captured, signed: Headers
) -> None:
    body = b'{"type":"link.created"}'
    headers = signed(body) | {"traceparent": TRACEPARENT}
    response = await traced_client.post("/hooks/orders", content=body, headers=headers)
    assert response.status_code == 200

    servers = [s for s in captured.spans.get_finished_spans() if s.name == "POST /hooks/{name}"]
    assert len(servers) == 1, [s.name for s in captured.spans.get_finished_spans()]
    span = servers[0]
    assert span.context is not None
    assert format(span.context.trace_id, "032x") == TRACE_ID
    assert span.attributes is not None
    assert span.attributes.get("http.route") == "/hooks/{name}"


@pytest.mark.anyio
async def test_request_duration_metric_uses_the_stable_name(
    traced_client: httpx.AsyncClient, captured: Captured, signed: Headers
) -> None:
    body = b"{}"
    await traced_client.post("/hooks/orders", content=body, headers=signed(body))
    await traced_client.get("/stats")  # excluded: no span, no metric

    data = captured.metrics.get_metrics_data()
    assert data is not None
    points = [
        point
        for rm in data.resource_metrics
        for sm in rm.scope_metrics
        for metric in sm.metrics
        if metric.name == "http.server.request.duration"
        for point in metric.data.data_points
    ]
    routes = {point.attributes.get("http.route") for point in points if point.attributes}
    assert routes == {"/hooks/{name}"}


@pytest.mark.anyio
async def test_log_records_carry_the_trace(
    traced_client: httpx.AsyncClient, captured: Captured, signed: Headers
) -> None:
    body = b"{}"
    headers = signed(body) | {"traceparent": TRACEPARENT}
    await traced_client.post("/hooks/orders", content=body, headers=headers)

    accepted = [
        f.log_record
        for f in captured.logs.get_finished_logs()
        if "delivery accepted" in str(f.log_record.body)
    ]
    assert accepted, "no 'delivery accepted' log record"
    assert format(accepted[0].trace_id or 0, "032x") == TRACE_ID


def test_filter_adds_empty_ids_outside_a_span() -> None:
    record = logging.LogRecord("t", logging.INFO, __file__, 1, "msg", None, None)
    assert TraceContextFilter().filter(record)
    assert record.__dict__["trace_id"] == ""
    assert record.__dict__["span_id"] == ""
