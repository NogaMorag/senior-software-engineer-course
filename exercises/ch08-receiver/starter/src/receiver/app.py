"""The FastAPI application: routes, dependencies and the lifespan that owns the state."""

import asyncio
import contextlib
import logging
import time
from collections.abc import AsyncIterator, Awaitable, Callable
from dataclasses import asdict, dataclass, field
from typing import Annotated, Protocol

from fastapi import Depends, FastAPI, Header, Request
from fastapi.responses import JSONResponse
from pydantic import BaseModel

from receiver.settings import Settings
from receiver.state import Control, ControlPatch, Store

# TODO 7 and 8 also need: import random; HTTPException from fastapi;
# SignatureError and verify from receiver.signature.

log = logging.getLogger("receiver")

type Clock = Callable[[], float]
type Sleep = Callable[[float], Awaitable[None]]


class RandomSource(Protocol):
    def random(self) -> float: ...


@dataclass(slots=True)
class AppState:
    settings: Settings
    clock: Clock
    rng: RandomSource
    sleep: Sleep
    store: Store = field(default_factory=Store)
    control: Control = field(default_factory=Control)


def get_state(request: Request) -> AppState:
    state: AppState = request.app.state.receiver
    return state


StateDep = Annotated[AppState, Depends(get_state)]


class Accepted(BaseModel):
    status: str = "ok"
    duplicate: bool


class StatsOut(BaseModel):
    received: int
    unique: int
    duplicates: int
    rejected: int
    failed: int
    hooks: dict[str, dict[str, int]]


class ReceiptOut(BaseModel):
    hook: str
    event_id: str
    delivery_id: str | None
    attempt: int | None
    first_at: float
    last_at: float
    count: int


async def report_stats(state: AppState, interval_s: float) -> None:
    """Background task owned by the lifespan: one log line per interval until cancelled."""
    while True:
        await asyncio.sleep(interval_s)
        log.info("stats %s", state.store.total.as_dict())


def create_app(
    settings: Settings | None = None,
    *,
    clock: Clock = time.time,
    rng: RandomSource | None = None,
    sleep: Sleep = asyncio.sleep,
) -> FastAPI:
    settings = settings or Settings.from_env()

    @contextlib.asynccontextmanager
    async def lifespan(app: FastAPI) -> AsyncIterator[None]:
        # TODO 7 (day 18): build the AppState (rng defaults to random.Random()), store it on
        # app.state.receiver, warn when no secret is configured, start report_stats as a task
        # when settings.stats_interval_s is set; yield; then cancel and await that task and log
        # "shutting down, final stats ...".
        raise NotImplementedError
        yield

    app = FastAPI(title="receiver", lifespan=lifespan)

    @app.post("/hooks/{name}", response_model=Accepted)
    async def receive(
        name: str,
        request: Request,
        state: StateDep,
        signature: Annotated[str | None, Header(alias="Hookline-Signature")] = None,
        event_id: Annotated[str | None, Header(alias="Hookline-Event-Id", max_length=200)] = None,
        delivery_id: Annotated[str | None, Header(alias="Hookline-Delivery-Id")] = None,
        attempt: Annotated[int | None, Header(alias="Hookline-Attempt", ge=1)] = None,
    ) -> Accepted | JSONResponse:
        # TODO 8 (day 18): 404 for an unknown hook; read the raw body; verify (401 with the reason,
        # counted as rejected); 400 without an event id (rejected); then timeout mode (sleep
        # timeout_hold_s, count failed, 504), latency, the failure-rate draw (count failed,
        # failure_status) and finally store.record. Use state.sleep, state.rng and state.clock.
        raise NotImplementedError

    @app.get("/stats")
    async def stats(state: StateDep) -> StatsOut:
        hooks = {hook: c.as_dict() for hook, c in sorted(state.store.by_hook.items())}
        return StatsOut(**state.store.total.as_dict(), hooks=hooks)

    @app.get("/receipts")
    async def receipts(state: StateDep, hook: str | None = None) -> list[ReceiptOut]:
        return [
            ReceiptOut(**asdict(r))
            for r in state.store.receipts.values()
            if hook is None or r.hook == hook
        ]

    @app.get("/control")
    async def get_control(state: StateDep) -> Control:
        return state.control

    @app.post("/control")
    async def set_control(patch: ControlPatch, state: StateDep) -> Control:
        # TODO 9 (day 18): merge only the fields that were sent into state.control.
        raise NotImplementedError

    @app.post("/reset")
    async def reset(state: StateDep) -> dict[str, str]:
        state.store = Store()
        state.control = Control()
        return {"status": "reset"}

    return app
