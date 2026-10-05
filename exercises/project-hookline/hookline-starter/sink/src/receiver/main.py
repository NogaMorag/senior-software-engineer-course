"""`uv run receiver`: read the settings, set up telemetry when a collector is configured, and serve.

Hookline's sink (milestone M8): telemetry is on only when OTEL_EXPORTER_OTLP_ENDPOINT is set, so the
sink of milestones M3 to M7, which runs without a collector, does not log failed exports.
"""
import logging
import os
import uvicorn
from receiver.app import create_app
from receiver.settings import Settings
from receiver.telemetry import configure_logging, instrument, setup_telemetry
PLAIN_FORMAT = '%(asctime)s %(levelname)s %(name)s %(message)s'

def run() -> None:
    raise NotImplementedError
