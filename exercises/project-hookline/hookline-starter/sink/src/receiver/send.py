"""`uv run receiver-send`: send one signed delivery, as Hookline's worker will (stdlib only)."""
import argparse
import time
import urllib.error
import urllib.request
import uuid
from collections.abc import Sequence
from receiver.signature import sign

def main(argv: Sequence[str] | None=None) -> int:
    raise NotImplementedError
