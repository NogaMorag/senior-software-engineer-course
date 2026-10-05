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
    # TODO 4 (day 18): split on ",", strip each item, partition on "=". Exactly one t= made of
    # ASCII digits (use _UNIX_SECONDS); one or more v1= matching _HEX_DIGEST, kept as bytes.
    # Anything else that breaks these rules raises SignatureError("malformed", ...).
    raise NotImplementedError


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
    # TODO 5 (day 18): missing header is malformed; parse it; reject when |now - t| is more than
    # tolerance_s; compare digest(secret, t, body) with every candidate using
    # hmac.compare_digest; raise signature_mismatch when none matches.
    raise NotImplementedError
