"""Shared fixture: write an assumptions file into a temporary directory."""

from collections.abc import Callable
from pathlib import Path

import pytest

VALID = """
name = "test"

[traffic]
requests_per_day = 8_640_000   # exactly 100 per second
peak_factor = 3.0

[[records]]
name = "events"
bytes = 1_000
per_request = 1.0
retention_days = 10

[[records]]
name = "deliveries"
bytes = 200
per_request = 2.0
retention_days = 5

[[transfers]]
name = "in"
bytes = 1_000
per_request = 1.0

[[transfers]]
name = "out"
bytes = 500
per_request = 2.0

[storage]
index_overhead = 0.5
replicas = 2
"""


@pytest.fixture
def write_toml(tmp_path: Path) -> Callable[[str], Path]:
    def write(text: str) -> Path:
        path = tmp_path / "assumptions.toml"
        path.write_text(text, encoding="utf-8")
        return path

    return write
