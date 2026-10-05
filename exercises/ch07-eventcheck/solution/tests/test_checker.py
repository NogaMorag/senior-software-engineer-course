"""check_lines over planted bad lines: every bad line is reported, with its number."""

import json
from pathlib import Path
from typing import Any

import pytest

from eventcheck.checker import Err, Ok, check_lines, parse_line
from eventcheck.generate import generate_lines


def test_parse_line_returns_ok_or_err(invoice: dict[str, Any]) -> None:
    assert isinstance(parse_line(1, json.dumps(invoice)), Ok)
    match parse_line(7, "{not json"):
        case Err(issues):
            assert [(i.line, i.kind) for i in issues] == [(7, "json")]
        case Ok():
            pytest.fail("broken JSON must not parse")


def test_planted_bad_lines_are_reported_by_number(
    invoice: dict[str, Any], shipment: dict[str, Any]
) -> None:
    negative = json.loads(json.dumps(invoice))
    negative["data"]["amount_cents"] = -1
    lines = [
        json.dumps(invoice),  # 1 ok
        json.dumps(shipment),  # 2 ok
        json.dumps(negative),  # 3 schema
        "",  # 4 blank: skipped, still counted
        '{"id": ',  # 5 json
        json.dumps(invoice),  # 6 duplicate of line 1
    ]
    report = check_lines(lines)
    assert report.lines == 6
    assert report.valid == 2
    assert report.bad_lines == [3, 5, 6]
    assert [issue.kind for issue in report.issues] == ["schema", "json", "duplicate"]
    assert report.issues[0].location == "data.amount_cents"
    assert "first seen on line 1" in report.issues[2].message
    assert dict(report.by_type) == {"invoice.paid": 1, "shipment.moved": 1}


def test_one_line_can_carry_several_issues(invoice: dict[str, Any]) -> None:
    invoice["tenant"] = "X"
    invoice["data"]["currency"] = "GBP"
    report = check_lines([json.dumps(invoice)])
    assert report.bad_lines == [1]
    assert sorted(issue.location for issue in report.issues) == ["data.currency", "tenant"]


def test_issue_str_names_line_kind_and_location() -> None:
    (issue,) = check_lines(['{"type": "invoice.paid"}']).issues[:1]
    assert str(issue).startswith("line 1: schema at ")


@pytest.mark.parametrize(
    ("count", "bad_every", "expected_bad"), [(10, 0, 0), (10, 5, 2), (100, 7, 14)]
)
def test_generator_plants_exactly_the_bad_lines(
    count: int, bad_every: int, expected_bad: int
) -> None:
    report = check_lines(generate_lines(count, bad_every=bad_every))
    assert len(report.bad_lines) == expected_bad
    assert report.valid == count - expected_bad


def test_generator_is_deterministic() -> None:
    assert list(generate_lines(20, seed=3)) == list(generate_lines(20, seed=3))
    assert list(generate_lines(20, seed=3)) != list(generate_lines(20, seed=4))


@pytest.mark.slow
def test_big_file(big_file: Path) -> None:
    with big_file.open(encoding="utf-8") as lines:
        report = check_lines(lines)
    assert report.lines == 50_000
    assert len(report.bad_lines) == 50
