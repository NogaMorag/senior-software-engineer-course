"""Parse a Hookline webhook body into an Event.

The parser trusts the shape the contract promises (../contract/webhook-event.schema.json) and
nothing more: it reads the four required properties and ignores every other one, so the producer
can add optional properties without breaking this consumer.
"""

import json
import uuid
from dataclasses import dataclass
from datetime import datetime
from typing import Any


@dataclass(frozen=True, slots=True)
class Event:
    id: uuid.UUID
    type: str
    created_at: datetime
    data: dict[str, Any]


def parse_event(body: bytes | str) -> Event:
    """Return the event in body, or raise ValueError if body does not hold one."""
    doc = json.loads(body)
    if not isinstance(doc, dict):
        raise ValueError("event body must be a JSON object")
    try:
        event_id, event_type = doc["id"], doc["type"]
        created_at, data = doc["created_at"], doc["data"]
    except KeyError as missing:
        raise ValueError(f"event body has no {missing} property") from None
    if not isinstance(event_id, str) or not isinstance(event_type, str):
        raise ValueError("id and type must be strings")
    if not isinstance(created_at, str):
        raise ValueError("created_at must be a string")
    if not isinstance(data, dict):
        raise ValueError("data must be a JSON object")
    when = datetime.fromisoformat(created_at)
    if when.tzinfo is None:
        raise ValueError("created_at must carry a UTC offset")
    return Event(id=uuid.UUID(event_id), type=event_type, created_at=when, data=data)
