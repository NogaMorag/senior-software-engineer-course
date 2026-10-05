"""Smoke test for the Python toolchain (chapter 2)."""

import platform
import sys


def describe() -> str:
    """Return the running interpreter in one line, for example "Python 3.14.8 on Linux"."""
    v = sys.version_info
    return f"Python {v.major}.{v.minor}.{v.micro} on {platform.system()}"
