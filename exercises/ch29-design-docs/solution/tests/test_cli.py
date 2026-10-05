from collections.abc import Callable
from pathlib import Path

import pytest
from conftest import VALID

from sizing.cli import main


def test_report_lines(
    write_toml: Callable[[str], Path], capsys: pytest.CaptureFixture[str]
) -> None:
    assert main([str(write_toml(VALID))]) == 0
    out = capsys.readouterr().out.splitlines()
    assert out[0] == "== test"
    assert out[1].split() == ["requests", "average", "100.0/s", "peak", "300.0/s"]
    assert out[2].startswith("events ")
    assert "8.0 GiB/day, kept 10 days" in out[2]
    assert out[4].startswith("storage retained ")
    assert out[5].split() == ["network", "average", "1.6", "Mbit/s", "peak", "4.8", "Mbit/s"]


def test_bad_file_exits_2(
    write_toml: Callable[[str], Path], capsys: pytest.CaptureFixture[str]
) -> None:
    assert main([str(write_toml("name = 1"))]) == 2
    assert capsys.readouterr().err.startswith("sizing: ")
