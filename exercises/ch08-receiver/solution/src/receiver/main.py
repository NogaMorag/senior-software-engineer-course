"""`uv run receiver`: read the settings and serve with uvicorn."""

import logging

import uvicorn

from receiver.app import create_app
from receiver.settings import Settings


def run() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s %(message)s")
    settings = Settings.from_env()
    uvicorn.run(create_app(settings), host=settings.host, port=settings.port)
