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

PLAIN_FORMAT = "%(asctime)s %(levelname)s %(name)s %(message)s"


def run() -> None:
    settings = Settings.from_env()
    app = create_app(settings)
    if not os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT"):
        logging.basicConfig(level=logging.INFO, format=PLAIN_FORMAT)
        uvicorn.run(app, host=settings.host, port=settings.port)
        return
    telemetry = setup_telemetry(os.environ.get("OTEL_SERVICE_NAME", "hookline-sink"))
    configure_logging(telemetry)
    instrument(app, telemetry)
    try:
        # log_config=None: keep the logging set up above; uvicorn's loggers propagate to it.
        uvicorn.run(app, host=settings.host, port=settings.port, log_config=None)
    finally:
        telemetry.shutdown()
