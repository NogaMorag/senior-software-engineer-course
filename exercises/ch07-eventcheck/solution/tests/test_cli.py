"""The CLI as a user meets it: arguments, environment, stdin, exit codes."""

import io
import json
import sys
from collections.abc import Callable
from pathlib import Path
from typing import Any

import pytest

from eventcheck.cli import EXIT_ISSUES, EXIT_OK, EXIT_USAGE, main

type WriteJsonl = Callable[[list[str]], Path]


def test_clean_file_exits_zero(
    write_jsonl: WriteJsonl, invoice: dict[str, Any], capsys: pytest.CaptureFixture[str]
) -> None:
    path = write_jsonl([json.dumps(invoice)])
    assert main(["check", str(path)]) == EXIT_OK
    assert "1 valid, 0 bad" in capsys.readouterr().out


def test_bad_lines_exit_one_and_are_listed(
    write_jsonl: WriteJsonl, invoice: dict[str, Any], capsys: pytest.CaptureFixture[str]
) -> None:
    path = write_jsonl([json.dumps(invoice), "nope", json.dumps(invoice)])
    assert main(["check", str(path)]) == EXIT_ISSUES
    out = capsys.readouterr().out
    assert "line 2: json" in out
    assert "line 3: duplicate" in out


def test_missing_file_is_a_usage_error(capsys: pytest.CaptureFixture[str]) -> None:
    assert main(["check", "/no/such/file.jsonl"]) == EXIT_USAGE
    assert "No such file" in capsys.readouterr().err


def test_format_comes_from_the_environment(
    monkeypatch: pytest.MonkeyPatch,
    write_jsonl: WriteJsonl,
    invoice: dict[str, Any],
    capsys: pytest.CaptureFixture[str],
) -> None:
    monkeypatch.setenv("EVENTCHECK_FORMAT", "json")
    main(["check", str(write_jsonl([json.dumps(invoice), "{"]))])
    summary = json.loads(capsys.readouterr().out)
    assert summary["bad_lines"] == [2]
    assert summary["issues"][0]["kind"] == "json"


def test_unknown_format_in_environment(
    monkeypatch: pytest.MonkeyPatch, write_jsonl: WriteJsonl, invoice: dict[str, Any]
) -> None:
    monkeypatch.setenv("EVENTCHECK_FORMAT", "yaml")
    assert main(["check", str(write_jsonl([json.dumps(invoice)]))]) == EXIT_USAGE


def test_dash_reads_stdin(
    monkeypatch: pytest.MonkeyPatch, invoice: dict[str, Any], capsys: pytest.CaptureFixture[str]
) -> None:
    monkeypatch.setattr(sys, "stdin", io.StringIO(json.dumps(invoice) + "\n"))
    assert main(["check", "-"]) == EXIT_OK
    assert "invoice.paid=1" in capsys.readouterr().out


def test_generate_then_check(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    out = tmp_path / "sample.jsonl"
    assert main(["generate", "30", str(out), "--bad-every", "10"]) == EXIT_OK
    assert main(["check", str(out)]) == EXIT_ISSUES
    assert "27 valid, 3 bad" in capsys.readouterr().out
