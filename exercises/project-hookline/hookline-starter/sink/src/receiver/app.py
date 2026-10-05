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
log = logging.getLogger('receiver')
type Clock = Callable[[], float]
type Sleep = Callable[[float], Awaitable[None]]

class RandomSource(Protocol):

    def random(self) -> float:
        raise NotImplementedError

@dataclass(slots=True)
class AppState:
    settings: Settings
    clock: Clock
    rng: RandomSource
    sleep: Sleep
    store: Store = field(default_factory=Store)
    control: Control = field(default_factory=Control)

def get_state(request: Request) -> AppState:
    raise NotImplementedError
StateDep = Annotated[AppState, Depends(get_state)]

class Accepted(BaseModel):
    status: str = 'ok'
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
    raise NotImplementedError

def create_app(settings: Settings | None=None, *, clock: Clock=time.time, rng: RandomSource | None=None, sleep: Sleep=asyncio.sleep) -> FastAPI:
    raise NotImplementedError
