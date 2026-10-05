"""Shared fixtures. pytest finds this file by name; a test asks for a fixture by parameter name."""

from collections.abc import Callable
from pathlib import Path
from typing import Any

import pytest

from eventcheck.generate import generate_lines


@pytest.fixture
def invoice() -> dict[str, Any]:
    """A fresh valid invoice event per test, so a test may mutate it freely."""
    return {
        "id": "6f1c2a9e-3b7d-4c1a-9e2f-1a2b3c4d5e6f",
        "tenant": "acme",
        "created_at": "2026-10-01T12:00:00+03:00",
        "type": "invoice.paid",
        "data": {"invoice_id": "inv_000001", "amount_cents": 1999, "currency": "ILS"},
    }


@pytest.fixture
def shipment() -> dict[str, Any]:
    return {
        "id": "0b3e1f52-8c4d-4e6a-a1b2-c3d4e5f60718",
        "tenant": "globex",
        "created_at": "2026-10-01T09:30:00Z",
        "type": "shipment.moved",
        "data": {"shipment_id": "shp_000001", "status": "in_transit", "location": "Haifa"},
    }


@pytest.fixture
def write_jsonl(tmp_path: Path) -> Callable[[list[str]], Path]:
    """A factory fixture: returns a function that writes lines into this test's tmp_path."""

    def write(lines: list[str]) -> Path:
        path = tmp_path / "events.jsonl"
        path.write_text("".join(line + "\n" for line in lines), encoding="utf-8")
        return path

    return write


@pytest.fixture(scope="session")
def big_file(tmp_path_factory: pytest.TempPathFactory) -> Path:
    """Built once per test session and shared: 50,000 lines, one bad line in every 1,000."""
    path = tmp_path_factory.mktemp("big") / "big.jsonl"
    with path.open("w", encoding="utf-8") as handle:
        for line in generate_lines(50_000, seed=11, bad_every=1_000):
            handle.write(line + "\n")
    return path
