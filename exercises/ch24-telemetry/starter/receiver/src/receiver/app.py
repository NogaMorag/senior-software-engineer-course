"""The FastAPI application: routes, dependencies and the lifespan that owns the state."""

import asyncio
import contextlib
import logging
import random
import time
from collections.abc import AsyncIterator, Awaitable, Callable
from dataclasses import asdict, dataclass, field
from typing import Annotated, Protocol

from fastapi import Depends, FastAPI, Header, HTTPException, Request
from fastapi.responses import JSONResponse
from pydantic import BaseModel

from receiver.settings import Settings
from receiver.signature import SignatureError, verify
from receiver.state import Control, ControlPatch, Store

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
        state = AppState(settings, clock, rng or random.Random(), sleep)
        app.state.receiver = state
        if not settings.secrets and not settings.default_secret:
            log.warning("no secrets configured: every hook will answer 404")
        reporter = None
        if settings.stats_interval_s:
            reporter = asyncio.create_task(report_stats(state, settings.stats_interval_s))
        try:
            yield
        finally:
            if reporter is not None:
                reporter.cancel()
                with contextlib.suppress(asyncio.CancelledError):
                    await reporter
            log.info("shutting down, final stats %s", state.store.total.as_dict())

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
        secret = state.settings.secret_for(name)
        if secret is None:
            raise HTTPException(404, detail=f"unknown hook {name!r}")
        body = await request.body()  # the raw bytes: the signature covers exactly these
        try:
            verify(signature, body, secret, now=state.clock(), tolerance_s=settings.tolerance_s)
        except SignatureError as error:
            state.store.reject(name)
            raise HTTPException(401, detail=error.reason) from error
        if not event_id:
            state.store.reject(name)
            raise HTTPException(400, detail="missing Hookline-Event-Id")

        control = state.control
        if control.timeout_mode:
            await state.sleep(settings.timeout_hold_s)
            state.store.fail(name)
            return JSONResponse({"detail": "timeout mode"}, status_code=504)
        if control.latency_ms:
            await state.sleep(control.latency_ms / 1000)
        if state.rng.random() < control.failure_rate:
            state.store.fail(name)
            return JSONResponse({"detail": "injected failure"}, status_code=control.failure_status)

        duplicate = state.store.record(name, event_id, delivery_id, attempt, state.clock())
        log.info("delivery accepted hook=%s event_id=%s duplicate=%s", name, event_id, duplicate)
        return Accepted(duplicate=duplicate)

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
        changes = patch.model_dump(exclude_none=True)
        state.control = state.control.model_copy(update=changes)
        log.info("control changed %s", changes)
        return state.control

    @app.post("/reset")
    async def reset(state: StateDep) -> dict[str, str]:
        state.store = Store()
        state.control = Control()
        return {"status": "reset"}

    return app
