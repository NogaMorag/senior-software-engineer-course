"""Hookline signatures: `t=<unix seconds>,v1=<hex HMAC-SHA256(secret, "<t>.<raw body>")>`."""
import hashlib
import hmac
import re
from typing import Literal
DEFAULT_TOLERANCE_S = 300
_HEX_DIGEST = re.compile('[0-9a-fA-F]{64}')
_UNIX_SECONDS = re.compile('[0-9]{1,12}')
type Reason = Literal['malformed', 'timestamp_out_of_tolerance', 'signature_mismatch']

class SignatureError(Exception):

    def __init__(self, reason: Reason, message: str) -> None:
        raise NotImplementedError

def digest(secret: bytes, timestamp: int, body: bytes) -> bytes:
    """The raw 32-byte HMAC over the exact bytes "<t>." followed by the body."""
    raise NotImplementedError

def sign(secret: bytes, timestamp: int, body: bytes) -> str:
    raise NotImplementedError

def parse(header: str) -> tuple[int, list[bytes]]:
    """Split the header into its timestamp and candidate signatures. Unknown keys are ignored."""
    raise NotImplementedError

def verify(header: str | None, body: bytes, secret: bytes, *, now: float, tolerance_s: int=DEFAULT_TOLERANCE_S) -> int:
    """Return the signed timestamp, or raise SignatureError. Checks run in a fixed order:
    format, then freshness, then the HMAC itself (compared in constant time)."""
    raise NotImplementedError
