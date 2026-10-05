"""Settings come from the environment once; bad values fail at startup, not on the first request."""

import pytest

from receiver.settings import Settings


def test_defaults() -> None:
    settings = Settings.from_env({})
    assert (settings.port, settings.tolerance_s, settings.secrets) == (8081, 300, {})
    assert settings.secret_for("orders") is None


def test_named_secrets_and_default() -> None:
    env = {"RECEIVER_SECRETS": "orders=a, billing=b", "RECEIVER_DEFAULT_SECRET": "z"}
    settings = Settings.from_env(env)
    assert settings.secret_for("billing") == b"b"
    assert settings.secret_for("anything") == b"z"


@pytest.mark.parametrize("raw", ["orders", "=a", "orders="])
def test_bad_secrets_fail_fast(raw: str) -> None:
    with pytest.raises(ValueError, match="not name=secret"):
        Settings.from_env({"RECEIVER_SECRETS": raw})
