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
        self.timeline.append(f"{time.monotonic() - self.started:6.2f}s  {text}")


async def fake_call(call: Call, limit: asyncio.Semaphore, timeout_s: float, t: Tracker) -> None:
    async with limit:
        t.in_flight += 1
        t.max_in_flight = max(t.max_in_flight, t.in_flight)
        t.note(f"start {call.name} (in flight: {t.in_flight})")
        try:
            async with asyncio.timeout(timeout_s):
                await asyncio.sleep(call.delay_s)
            if call.fail:
                raise RuntimeError(f"{call.name} failed")
            t.results[call.name] = "ok"
            t.note(f"done  {call.name}")
        except TimeoutError:
            t.results[call.name] = "timeout"  # handled here: the siblings keep running
            t.note(f"timeout {call.name}")
        except RuntimeError:
            t.results[call.name] = "failed"
            t.note(f"failed {call.name}")
            raise  # the TaskGroup sees it and cancels the siblings
        except asyncio.CancelledError:
            t.results[call.name] = "cancelled"
            t.note(f"cancelled {call.name}")
            raise  # never swallow cancellation
        finally:
            t.in_flight -= 1


async def run_calls(calls: list[Call], *, limit: int, timeout_s: float) -> Tracker:
    tracker = Tracker()
    semaphore = asyncio.Semaphore(limit)
    try:
        async with asyncio.TaskGroup() as group:
            for call in calls:
                group.create_task(fake_call(call, semaphore, timeout_s, tracker), name=call.name)
    except* RuntimeError as failures:
        for error in failures.exceptions:
            tracker.note(f"group raised: {error}")
    return tracker


def blocking_job(seconds: float) -> None:
    time.sleep(seconds)  # stands in for a blocking library call or a CPU-heavy function


async def count_ticks(job: Callable[[], Awaitable[None]], tick_s: float = 0.05) -> int:
    """Run `job` next to a heartbeat that ticks every tick_s; return the ticks it managed."""
    ticks = 0

    async def heartbeat() -> None:
        nonlocal ticks
        while True:
            await asyncio.sleep(tick_s)
            ticks += 1

    beat = asyncio.create_task(heartbeat())
    await asyncio.sleep(0)  # let the heartbeat start
    try:
        await job()
    finally:
        beat.cancel()
        with contextlib.suppress(asyncio.CancelledError):
            await beat
    return ticks


async def inline_block() -> None:
    blocking_job(0.5)  # the mistake on purpose; ruff ASYNC251 only sees a direct time.sleep


async def in_thread() -> None:
    await asyncio.to_thread(blocking_job, 0.5)


def show(title: str, tracker: Tracker) -> None:
    print(f"== {title} (max in flight: {tracker.max_in_flight})")
    print("\n".join(tracker.timeline))
    print("results:", dict(sorted(tracker.results.items())))


async def main() -> None:
    calls = [Call("a", 0.3), Call("b", 0.3), Call("c", 0.3), Call("slow", 5.0), Call("d", 0.3)]
    show("timeout handled per call", await run_calls(calls, limit=2, timeout_s=1.0))
    calls = [Call("a", 1.0), Call("boom", 0.2, fail=True), Call("b", 1.0), Call("c", 1.0)]
    show("one failure cancels the siblings", await run_calls(calls, limit=3, timeout_s=2.0))
    print("== a 0.5 s blocking call, heartbeat every 0.05 s")
    print("ticks while blocking the loop:", await count_ticks(inline_block))
    print("ticks while in a thread:      ", await count_ticks(in_thread))


if __name__ == "__main__":
    asyncio.run(main())
