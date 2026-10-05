"""`uv run receiver`: read the settings, set up telemetry and serve with uvicorn."""

import uvicorn

from receiver.app import create_app
from receiver.settings import Settings
from receiver.telemetry import configure_logging, instrument, setup_telemetry


def run() -> None:
    telemetry = setup_telemetry("receiver")
    configure_logging(telemetry)
    settings = Settings.from_env()
    app = create_app(settings)
    instrument(app, telemetry)
    try:
        # log_config=None: keep the logging set up above; uvicorn's loggers propagate to it.
        uvicorn.run(app, host=settings.host, port=settings.port, log_config=None)
    finally:
        telemetry.shutdown()
