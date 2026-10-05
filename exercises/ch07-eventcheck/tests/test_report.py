"""Reporters, checked without pydantic: the day 15 part of the exercise."""

import io
import json

from eventcheck.checker import Issue, Report
from eventcheck.report import REPORTERS, JsonReporter, Reporter, TextReporter, summary_dict


def make_report() -> Report:
    report = Report(lines=3, valid=1)
    report.issues.append(Issue(2, "json", "", "Invalid JSON"))
    report.issues.append(Issue(3, "schema", "data.currency", "Input should be 'ILS'"))
    report.by_type["invoice.paid"] += 1
    return report


def test_text_reporter() -> None:
    out = io.StringIO()
    TextReporter().write(make_report(), out)
    assert out.getvalue() == (
        "line 2: json: Invalid JSON\n"
        "line 3: schema at data.currency: Input should be 'ILS'\n"
        "checked 3 lines: 1 valid, 2 bad (invoice.paid=1)\n"
    )


def test_text_reporter_with_no_valid_events() -> None:
    out = io.StringIO()
    TextReporter().write(Report(lines=1, issues=[Issue(1, "json", "", "x")]), out)
    assert out.getvalue().endswith("checked 1 lines: 0 valid, 1 bad (no valid events)\n")


def test_json_reporter_writes_the_summary() -> None:
    out = io.StringIO()
    JsonReporter().write(make_report(), out)
    assert json.loads(out.getvalue()) == {
        "lines": 3,
        "valid": 1,
        "bad_lines": [2, 3],
        "by_type": {"invoice.paid": 1},
        "issues": [
            {"line": 2, "kind": "json", "location": "", "message": "Invalid JSON"},
            {
                "line": 3,
                "kind": "schema",
                "location": "data.currency",
                "message": "Input should be 'ILS'",
            },
        ],
    }


def test_summary_dict_is_plain_data() -> None:
    summary = summary_dict(make_report())
    assert type(summary["by_type"]) is dict  # a Counter would serialize, but is not the contract


def test_reporters_satisfy_the_protocol() -> None:
    reporters: list[Reporter] = [TextReporter(), JsonReporter()]  # mypy checks this line
    assert sorted(REPORTERS) == ["json", "text"]
    assert all(hasattr(r, "write") for r in reporters)
