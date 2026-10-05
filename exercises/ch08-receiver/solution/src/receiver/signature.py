"""Hookline signatures: `t=<unix seconds>,v1=<hex HMAC-SHA256(secret, "<t>.<raw body>")>`."""

import hashlib
import hmac
import re
from typing import Literal

DEFAULT_TOLERANCE_S = 300
_HEX_DIGEST = re.compile(r"[0-9a-fA-F]{64}")
_UNIX_SECONDS = re.compile(r"[0-9]{1,12}")  # ASCII digits only: str.isdigit() accepts "²"

type Reason = Literal["malformed", "timestamp_out_of_tolerance", "signature_mismatch"]


class SignatureError(Exception):
    def __init__(self, reason: Reason, message: str) -> None:
        super().__init__(message)
        self.reason: Reason = reason


def digest(secret: bytes, timestamp: int, body: bytes) -> bytes:
    """The raw 32-byte HMAC over the exact bytes "<t>." followed by the body."""
    return hmac.new(secret, f"{timestamp}.".encode() + body, hashlib.sha256).digest()


def sign(secret: bytes, timestamp: int, body: bytes) -> str:
    return f"t={timestamp},v1={digest(secret, timestamp, body).hex()}"


def parse(header: str) -> tuple[int, list[bytes]]:
    """Split the header into its timestamp and candidate signatures. Unknown keys are ignored."""
    timestamps: list[str] = []
    candidates: list[bytes] = []
    for item in header.split(","):
        key, sep, value = item.strip().partition("=")
        if not sep:
            raise SignatureError("malformed", f"item {item.strip()!r} is not key=value")
        if key == "t":
            timestamps.append(value)
        elif key == "v1":
            if not _HEX_DIGEST.fullmatch(value):
                raise SignatureError("malformed", "v1 must be 64 hex characters")
            candidates.append(bytes.fromhex(value))
    if len(timestamps) != 1 or not _UNIX_SECONDS.fullmatch(timestamps[0]):
        raise SignatureError("malformed", "exactly one numeric t= is required")
    if not candidates:
        raise SignatureError("malformed", "at least one v1= is required")
    return int(timestamps[0]), candidates


def verify(
    header: str | None,
    body: bytes,
    secret: bytes,
    *,
    now: float,
    tolerance_s: int = DEFAULT_TOLERANCE_S,
) -> int:
    """Return the signed timestamp, or raise SignatureError. Checks run in a fixed order:
    format, then freshness, then the HMAC itself (compared in constant time)."""
    if not header:
        raise SignatureError("malformed", "missing signature header")
    timestamp, candidates = parse(header)
    if abs(now - timestamp) > tolerance_s:
        raise SignatureError(
            "timestamp_out_of_tolerance", f"t={timestamp} is off by more than {tolerance_s}s"
        )
    expected = digest(secret, timestamp, body)
    # Check every candidate, without stopping early, so timing says nothing about which matched.
    matched = False
    for candidate in candidates:
        matched |= hmac.compare_digest(expected, candidate)
    if not matched:
        raise SignatureError("signature_mismatch", "no v1 signature matches")
    return timestamp
