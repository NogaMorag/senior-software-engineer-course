"""In-memory receipts, counters and the misbehaviour controls. One event loop, no locks needed:
nothing here awaits, so no other request can run between a check and the update that follows it."""

from dataclasses import asdict, dataclass, field
from typing import Annotated

from pydantic import BaseModel, ConfigDict, Field

FailureRate = Annotated[float, Field(ge=0.0, le=1.0)]
LatencyMs = Annotated[int, Field(ge=0, le=60_000)]
FailureStatus = Annotated[int, Field(ge=400, le=599)]


class Control(BaseModel):
    """How the receiver misbehaves. Defaults: a perfectly healthy receiver."""

    model_config = ConfigDict(extra="forbid")
    failure_rate: FailureRate = 0.0
    latency_ms: LatencyMs = 0
    failure_status: FailureStatus = 503
    timeout_mode: bool = False


class ControlPatch(BaseModel):
    """The body of POST /control: only the fields that are sent change."""

    model_config = ConfigDict(extra="forbid")
    failure_rate: FailureRate | None = None
    latency_ms: LatencyMs | None = None
    failure_status: FailureStatus | None = None
    timeout_mode: bool | None = None


@dataclass(slots=True)
class Counters:
    received: int = 0  # signed requests accepted, duplicates included
    unique: int = 0  # distinct (hook, event id) pairs
    duplicates: int = 0  # accepted requests whose (hook, event id) was already seen
    rejected: int = 0  # bad or missing signature, missing event id
    failed: int = 0  # failures injected through /control (not recorded as received)

    def as_dict(self) -> dict[str, int]:
        return asdict(self)


@dataclass(slots=True)
class Receipt:
    hook: str
    event_id: str
    delivery_id: str | None
    attempt: int | None
    first_at: float
    last_at: float
    count: int = 1


@dataclass(slots=True)
class Store:
    total: Counters = field(default_factory=Counters)
    by_hook: dict[str, Counters] = field(default_factory=dict)
    receipts: dict[tuple[str, str], Receipt] = field(default_factory=dict)

    def _counters(self, hook: str) -> list[Counters]:
        return [self.total, self.by_hook.setdefault(hook, Counters())]

    def reject(self, hook: str) -> None:
        for counters in self._counters(hook):
            counters.rejected += 1

    def fail(self, hook: str) -> None:
        for counters in self._counters(hook):
            counters.failed += 1

    def record(
        self, hook: str, event_id: str, delivery_id: str | None, attempt: int | None, now: float
    ) -> bool:
        """Record an accepted delivery. Returns True when it is a duplicate."""
        # TODO 6 (day 18): update received and unique or duplicates in both counters
        # (self._counters(hook)), then add a Receipt or bump the existing one's count and last_at.
        raise NotImplementedError
