"""Fixtures: fake time and randomness, and an httpx client wired straight into the ASGI app."""

from collections.abc import AsyncIterator

import httpx
import pytest
from fastapi import FastAPI

from receiver.app import create_app
from receiver.settings import Settings
from receiver.signature import sign

from helpers import NOW, SECRET, FakeClock, FakeRandom, FakeSleep, Headers


@pytest.fixture
def anyio_backend() -> str:
    """The anyio pytest plugin (installed with httpx) runs `@pytest.mark.anyio` tests on asyncio."""
    return "asyncio"


@pytest.fixture
def clock() -> FakeClock:
    return FakeClock()


@pytest.fixture
def rng() -> FakeRandom:
    return FakeRandom()


@pytest.fixture
def sleep() -> FakeSleep:
    return FakeSleep()


@pytest.fixture
def settings() -> Settings:
    return Settings(secrets={"orders": SECRET, "billing": "whsec_billing-change-me"})


@pytest.fixture
def app(settings: Settings, clock: FakeClock, rng: FakeRandom, sleep: FakeSleep) -> FastAPI:
    return create_app(settings, clock=clock, rng=rng, sleep=sleep)


@pytest.fixture
async def client(app: FastAPI) -> AsyncIterator[httpx.AsyncClient]:
    """ASGITransport does not run the lifespan, so the fixture enters it explicitly."""
    async with app.router.lifespan_context(app):
        transport = httpx.ASGITransport(app=app)
        async with httpx.AsyncClient(transport=transport, base_url="http://test") as client:
            yield client


@pytest.fixture
def signed() -> Headers:
    """Build the headers Hookline's worker sends, signed over `body` at time `t`."""

    def build(
        body: bytes, *, event_id: str = "evt_1", secret: str = SECRET, t: int = NOW
    ) -> dict[str, str]:
        return {
            "Hookline-Event-Id": event_id,
            "Hookline-Delivery-Id": f"dlv_{event_id}",
            "Hookline-Attempt": "1",
            "Hookline-Signature": sign(secret.encode(), t, body),
        }

    return build
