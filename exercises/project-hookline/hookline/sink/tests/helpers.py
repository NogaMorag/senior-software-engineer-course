"""Test helpers shared by several test files: fixed time, fakes, the shared vectors.
(Fixtures live in conftest.py; pytest's documentation advises against importing from it.)"""

import json
from collections.abc import Callable
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

NOW = 1_791_201_600  # 2026-10-05 12:00:00 UTC
SECRET = "whsec_change-me"


@dataclass
class FakeClock:
    now: float = NOW

    def __call__(self) -> float:
        return self.now


@dataclass
class FakeRandom:
    """Returns the scripted values in order, then 0.99 (no failure at any rate below 0.99)."""

    values: list[float] = field(default_factory=list)

    def random(self) -> float:
        return self.values.pop(0) if self.values else 0.99


@dataclass
class FakeSleep:
    """Records every requested delay and returns at once: latency tests take no time."""

    calls: list[float] = field(default_factory=list)

    async def __call__(self, delay: float) -> None:
        self.calls.append(delay)


type Headers = Callable[..., dict[str, str]]


def find_vectors() -> Path:
    """data/signature-vectors.json in the nearest ancestor folder that has one."""
    for folder in Path(__file__).resolve().parents:
        candidate = folder / "data" / "signature-vectors.json"
        if candidate.exists():
            return candidate
    raise FileNotFoundError("data/signature-vectors.json not found; copy data/ next to tests/")


def load_vectors() -> dict[str, Any]:
    vectors: dict[str, Any] = json.loads(find_vectors().read_text(encoding="utf-8"))
    return vectors
