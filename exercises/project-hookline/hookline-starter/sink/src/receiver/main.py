"""`uv run receiver`: read the settings and serve with uvicorn."""
import logging
import uvicorn
from receiver.app import create_app
from receiver.settings import Settings

def run() -> None:
    raise NotImplementedError
