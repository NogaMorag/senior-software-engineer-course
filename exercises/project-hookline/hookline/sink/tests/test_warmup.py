"""The day 17 warm-up, checked: the limit holds, a handled timeout spares the siblings,
an unhandled failure cancels them. Delays are tiny but at least ten times apart."""

import pytest

from receiver.warmup import Call, count_ticks, in_thread, inline_block, run_calls

pytestmark = pytest.mark.anyio


async def test_semaphore_limits_concurrency() -> None:
    calls = [Call(name, 0.02) for name in "abcdef"]
    tracker = await run_calls(calls, limit=2, timeout_s=1.0)
    assert tracker.max_in_flight == 2
    assert set(tracker.results.values()) == {"ok"}


async def test_timeout_is_handled_per_call() -> None:
    calls = [Call("a", 0.01), Call("slow", 5.0), Call("b", 0.01)]
    tracker = await run_calls(calls, limit=3, timeout_s=0.1)
    assert tracker.results == {"a": "ok", "slow": "timeout", "b": "ok"}


async def test_failure_cancels_siblings() -> None:
    calls = [Call("a", 5.0), Call("boom", 0.01, fail=True), Call("b", 5.0)]
    tracker = await run_calls(calls, limit=3, timeout_s=10.0)
    assert tracker.results == {"a": "cancelled", "boom": "failed", "b": "cancelled"}
    assert any("group raised: boom failed" in line for line in tracker.timeline)


async def test_blocking_call_starves_the_loop_but_a_thread_does_not() -> None:
    blocked = await count_ticks(inline_block)
    threaded = await count_ticks(in_thread)
    assert blocked <= 1  # the heartbeat could not run while time.sleep held the thread
    assert threaded >= 5  # about ten ticks fit into 0.5 s; demand half to tolerate a busy machine
