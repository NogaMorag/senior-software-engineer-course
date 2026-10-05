"""The signature contract, driven by the shared vectors the Go signer reuses."""

import hmac
from typing import Any

import pytest

from receiver.signature import SignatureError, sign, verify

from helpers import load_vectors

VECTORS = load_vectors()


@pytest.mark.parametrize("case", VECTORS["sign"], ids=lambda case: case["name"])
def test_sign_vectors(case: dict[str, Any]) -> None:
    header = sign(case["secret"].encode(), case["timestamp"], case["body"].encode())
    assert header == case["header"]


@pytest.mark.parametrize("case", VECTORS["verify"], ids=lambda case: case["name"])
def test_verify_vectors(case: dict[str, Any]) -> None:
    def check() -> None:
        verify(
            case["header"],
            case["body"].encode(),
            case["secret"].encode(),
            now=case["now"],
            tolerance_s=VECTORS["tolerance_seconds"],
        )

    if case["valid"]:
        check()
    else:
        with pytest.raises(SignatureError) as caught:
            check()
        assert caught.value.reason == case["reason"]


def test_missing_header_is_malformed() -> None:
    with pytest.raises(SignatureError) as caught:
        verify(None, b"{}", b"s", now=0)
    assert caught.value.reason == "malformed"


def test_comparison_is_constant_time(monkeypatch: pytest.MonkeyPatch) -> None:
    calls: list[tuple[bytes, bytes]] = []
    real = hmac.compare_digest

    def spy(a: bytes, b: bytes) -> bool:
        calls.append((a, b))
        return real(a, b)

    monkeypatch.setattr(hmac, "compare_digest", spy)
    header = sign(b"s", 100, b"{}")
    verify(header, b"{}", b"s", now=100)
    assert len(calls) == 1
    assert len(calls[0][0]) == 32  # raw digests are compared, not hex text
