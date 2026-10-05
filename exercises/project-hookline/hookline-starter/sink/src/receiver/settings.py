"""Configuration from environment variables, parsed once at startup."""

import os
from collections.abc import Mapping
from dataclasses import dataclass, field
from typing import Self

from receiver.signature import DEFAULT_TOLERANCE_S


def parse_secrets(raw: str) -> dict[str, str]:
    """`orders=change-me,billing=change-me-too` becomes {"orders": "change-me", ...}."""
    secrets: dict[str, str] = {}
    for item in filter(None, (part.strip() for part in raw.split(","))):
        name, sep, secret = item.partition("=")
        if not sep or not name or not secret:
            raise ValueError(f"RECEIVER_SECRETS item {item!r} is not name=secret")
        secrets[name] = secret
    return secrets


@dataclass(frozen=True, slots=True)
class Settings:
    secrets: dict[str, str] = field(default_factory=dict)
    default_secret: str | None = None
    tolerance_s: int = DEFAULT_TOLERANCE_S
    timeout_hold_s: float = 3600.0
    stats_interval_s: float | None = None
    host: str = "127.0.0.1"
    port: int = 8081

    def secret_for(self, hook: str) -> bytes | None:
        secret = self.secrets.get(hook, self.default_secret)
        return secret.encode() if secret else None

    @classmethod
    def from_env(cls, env: Mapping[str, str] = os.environ) -> Self:
        interval = env.get("RECEIVER_STATS_INTERVAL_S")
        return cls(
            secrets=parse_secrets(env.get("RECEIVER_SECRETS", "")),
            default_secret=env.get("RECEIVER_DEFAULT_SECRET") or None,
            tolerance_s=int(env.get("RECEIVER_TOLERANCE_S", DEFAULT_TOLERANCE_S)),
            timeout_hold_s=float(env.get("RECEIVER_TIMEOUT_HOLD_S", 3600)),
            stats_interval_s=float(interval) if interval else None,
            host=env.get("RECEIVER_HOST", "127.0.0.1"),
            port=int(env.get("RECEIVER_PORT", 8081)),
        )
