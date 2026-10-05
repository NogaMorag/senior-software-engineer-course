"""The receiver end to end, in process: no sockets, no real time, no real randomness."""

import asyncio

import httpx
import pytest

from receiver.app import create_app
from receiver.settings import Settings

from helpers import NOW, FakeClock, FakeRandom, FakeSleep, Headers

pytestmark = pytest.mark.anyio
BODY = b'{"type":"invoice.paid","data":{"amount_cents":1999}}'


async def stats(client: httpx.AsyncClient) -> dict[str, object]:
    response = await client.get("/stats")
    data: dict[str, object] = response.json()
    return data


async def test_signed_delivery_is_accepted(client: httpx.AsyncClient, signed: Headers) -> None:
    response = await client.post("/hooks/orders", content=BODY, headers=signed(BODY))
    assert response.status_code == 200
    assert response.json() == {"status": "ok", "duplicate": False}
    assert await stats(client) == {
        "received": 1,
        "unique": 1,
        "duplicates": 0,
        "rejected": 0,
        "failed": 0,
        "hooks": {
            "orders": {"received": 1, "unique": 1, "duplicates": 0, "rejected": 0, "failed": 0}
        },
    }


async def test_duplicate_event_is_recorded_once(client: httpx.AsyncClient, signed: Headers) -> None:
    for _ in range(3):
        response = await client.post("/hooks/orders", content=BODY, headers=signed(BODY))
    assert response.json()["duplicate"] is True
    totals = await stats(client)
    assert (totals["received"], totals["unique"], totals["duplicates"]) == (3, 1, 2)
    (receipt,) = (await client.get("/receipts")).json()
    assert (receipt["event_id"], receipt["count"], receipt["first_at"]) == ("evt_1", 3, NOW)


async def test_same_event_on_two_hooks_is_two_deliveries(
    client: httpx.AsyncClient, signed: Headers
) -> None:
    await client.post("/hooks/orders", content=BODY, headers=signed(BODY))
    billing = signed(BODY, secret="whsec_billing-change-me")
    await client.post("/hooks/billing", content=BODY, headers=billing)
    assert (await stats(client))["unique"] == 2
    assert len((await client.get("/receipts", params={"hook": "billing"})).json()) == 1


@pytest.mark.parametrize(
    ("change", "reason"),
    [
        ("wrong-secret", "signature_mismatch"),
        ("tampered-body", "signature_mismatch"),
        ("no-signature", "malformed"),
        ("stale", "timestamp_out_of_tolerance"),
    ],
)
async def test_bad_signature_is_rejected(
    client: httpx.AsyncClient, signed: Headers, change: str, reason: str
) -> None:
    headers = signed(BODY, secret="whsec_wrong") if change == "wrong-secret" else signed(BODY)
    body = BODY.replace(b"1999", b"1") if change == "tampered-body" else BODY
    if change == "no-signature":
        del headers["Hookline-Signature"]
    if change == "stale":
        headers = signed(BODY, t=NOW - 301)
    response = await client.post("/hooks/orders", content=body, headers=headers)
    assert response.status_code == 401
    assert response.json() == {"detail": reason}
    totals = await stats(client)
    assert (totals["rejected"], totals["received"]) == (1, 0)


async def test_missing_event_id_is_rejected(client: httpx.AsyncClient, signed: Headers) -> None:
    headers = signed(BODY)
    del headers["Hookline-Event-Id"]
    response = await client.post("/hooks/orders", content=BODY, headers=headers)
    assert response.status_code == 400
    assert (await stats(client))["rejected"] == 1


async def test_unknown_hook_is_404(client: httpx.AsyncClient, signed: Headers) -> None:
    response = await client.post("/hooks/nobody", content=BODY, headers=signed(BODY))
    assert response.status_code == 404


async def test_failure_rate_injects_errors(
    client: httpx.AsyncClient, signed: Headers, rng: FakeRandom
) -> None:
    control = await client.post("/control", json={"failure_rate": 0.5, "failure_status": 500})
    assert control.json()["failure_rate"] == 0.5
    rng.values = [0.2, 0.7]  # first draw is below 0.5 (fail), second above (succeed)
    first = await client.post("/hooks/orders", content=BODY, headers=signed(BODY))
    second = await client.post("/hooks/orders", content=BODY, headers=signed(BODY))
    assert (first.status_code, second.status_code) == (500, 200)
    assert second.json()["duplicate"] is False  # the failed attempt was not recorded
    totals = await stats(client)
    assert (totals["failed"], totals["received"], totals["unique"]) == (1, 1, 1)


async def test_latency_uses_the_injected_sleep(
    client: httpx.AsyncClient, signed: Headers, sleep: FakeSleep
) -> None:
    await client.post("/control", json={"latency_ms": 250})
    response = await client.post("/hooks/orders", content=BODY, headers=signed(BODY))
    assert response.status_code == 200
    assert sleep.calls == [0.25]


async def test_control_patch_keeps_other_fields(client: httpx.AsyncClient) -> None:
    await client.post("/control", json={"latency_ms": 100})
    response = await client.post("/control", json={"failure_rate": 1.0})
    assert response.json() == {
        "failure_rate": 1.0,
        "latency_ms": 100,
        "failure_status": 503,
        "timeout_mode": False,
    }


@pytest.mark.parametrize(
    "body",
    [{"failure_rate": 1.5}, {"latency_ms": -1}, {"failure_status": 200}, {"surprise": True}],
)
async def test_control_rejects_invalid_values(
    client: httpx.AsyncClient, body: dict[str, object]
) -> None:
    response = await client.post("/control", json=body)
    assert response.status_code == 422
    assert (await client.get("/control")).json()["failure_rate"] == 0.0


async def test_timeout_mode_holds_then_fails(
    client: httpx.AsyncClient, signed: Headers, sleep: FakeSleep
) -> None:
    await client.post("/control", json={"timeout_mode": True})
    response = await client.post("/hooks/orders", content=BODY, headers=signed(BODY))
    assert response.status_code == 504
    assert sleep.calls == [3600.0]
    assert (await stats(client))["failed"] == 1


async def test_timeout_mode_really_times_out_a_client(signed: Headers) -> None:
    settings = Settings(secrets={"orders": "whsec_change-me"}, timeout_hold_s=30)
    app = create_app(settings, clock=FakeClock(), rng=FakeRandom())  # the real asyncio.sleep
    async with app.router.lifespan_context(app):
        transport = httpx.ASGITransport(app=app)
        async with httpx.AsyncClient(transport=transport, base_url="http://test") as client:
            await client.post("/control", json={"timeout_mode": True})
            with pytest.raises(TimeoutError):
                async with asyncio.timeout(0.2):
                    await client.post("/hooks/orders", content=BODY, headers=signed(BODY))


async def test_reset_clears_receipts_and_control(
    client: httpx.AsyncClient, signed: Headers
) -> None:
    await client.post("/hooks/orders", content=BODY, headers=signed(BODY))
    await client.post("/control", json={"latency_ms": 5})
    assert (await client.post("/reset")).json() == {"status": "reset"}
    assert (await stats(client))["received"] == 0
    assert (await client.get("/receipts")).json() == []
    assert (await client.get("/control")).json()["latency_ms"] == 0


async def test_clock_moves_receipt_times(
    client: httpx.AsyncClient, signed: Headers, clock: FakeClock
) -> None:
    await client.post("/hooks/orders", content=BODY, headers=signed(BODY))
    clock.now += 120  # two minutes later: still inside the five-minute tolerance
    await client.post("/hooks/orders", content=BODY, headers=signed(BODY))
    (receipt,) = (await client.get("/receipts")).json()
    assert receipt["last_at"] - receipt["first_at"] == 120


async def test_lifespan_runs_and_stops_the_stats_task(caplog: pytest.LogCaptureFixture) -> None:
    settings = Settings(default_secret="whsec_change-me", stats_interval_s=0.01)
    app = create_app(settings, clock=FakeClock(), rng=FakeRandom())
    with caplog.at_level("INFO", logger="receiver"):
        async with app.router.lifespan_context(app):
            await asyncio.sleep(0.05)
    messages = [record.getMessage() for record in caplog.records]
    assert any(message.startswith("stats ") for message in messages)
    assert messages[-1].startswith("shutting down")
