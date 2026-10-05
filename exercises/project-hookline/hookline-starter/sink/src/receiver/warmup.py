"""Day 17 warm-up: slow fake calls in a TaskGroup, a per-call timeout, and a Semaphore limit.

uv run python -m receiver.warmup
"""
import asyncio
import contextlib
import time
from collections.abc import Awaitable, Callable
from dataclasses import dataclass, field

@dataclass(frozen=True, slots=True)
class Call:
    name: str
    delay_s: float
    fail: bool = False

@dataclass(slots=True)
class Tracker:
    started: float = field(default_factory=time.monotonic)
    in_flight: int = 0
    max_in_flight: int = 0
    results: dict[str, str] = field(default_factory=dict)
    timeline: list[str] = field(default_factory=list)

    def note(self, text: str) -> None:
        raise NotImplementedError

async def fake_call(call: Call, limit: asyncio.Semaphore, timeout_s: float, t: Tracker) -> None:
    raise NotImplementedError

async def run_calls(calls: list[Call], *, limit: int, timeout_s: float) -> Tracker:
    raise NotImplementedError

def blocking_job(seconds: float) -> None:
    raise NotImplementedError

async def count_ticks(job: Callable[[], Awaitable[None]], tick_s: float=0.05) -> int:
    """Run `job` next to a heartbeat that ticks every tick_s; return the ticks it managed."""
    raise NotImplementedError

async def inline_block() -> None:
    raise NotImplementedError

async def in_thread() -> None:
    raise NotImplementedError

def show(title: str, tracker: Tracker) -> None:
    raise NotImplementedError

async def main() -> None:
    raise NotImplementedError
if __name__ == '__main__':
    asyncio.run(main())
