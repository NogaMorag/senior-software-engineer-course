"""Deterministic sample data: valid events with a bad line planted every N lines."""

import json
import random
from collections.abc import Iterator
from datetime import UTC, datetime, timedelta
from uuid import UUID

START = datetime(2026, 10, 1, tzinfo=UTC)
TENANTS = ("acme", "globex", "initech")


def valid_event(rng: random.Random, n: int) -> dict[str, object]:
    common = {
        "id": str(UUID(int=rng.getrandbits(128), version=4)),
        "tenant": rng.choice(TENANTS),
        "created_at": (START + timedelta(seconds=n)).isoformat(),
    }
    if rng.random() < 0.5:
        data: dict[str, object] = {
            "invoice_id": f"inv_{n:06d}",
            "amount_cents": rng.randint(100, 500_000),
            "currency": rng.choice(["ILS", "USD", "EUR"]),
        }
        return common | {"type": "invoice.paid", "data": data}
    status = rng.choice(["picked_up", "in_transit", "delivered"])
    data = {"shipment_id": f"shp_{n:06d}", "status": status, "location": "Haifa"}
    return common | {"type": "shipment.moved", "data": data}


def bad_line(rng: random.Random, n: int, previous_id: str) -> str:
    event = valid_event(rng, n)
    match n % 5:
        case 0:
            return json.dumps(event)[:-7]  # truncated JSON
        case 1:
            event["data"] = {"invoice_id": f"inv_{n:06d}", "amount_cents": -5, "currency": "ILS"}
            event["type"] = "invoice.paid"
        case 2:
            del event["tenant"]
        case 3:
            event["id"] = previous_id  # duplicate of the line before
        case _:
            event["type"] = "invoice.refunded"
    return json.dumps(event)


def generate_lines(count: int, *, seed: int = 7, bad_every: int = 0) -> Iterator[str]:
    rng = random.Random(seed)
    previous_id = ""
    for n in range(1, count + 1):
        if bad_every and n % bad_every == 0:
            yield bad_line(rng, n, previous_id)
            continue
        event = valid_event(rng, n)
        previous_id = str(event["id"])
        yield json.dumps(event)
