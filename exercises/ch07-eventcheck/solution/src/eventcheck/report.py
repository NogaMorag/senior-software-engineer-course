"""Reporters: turn a Report into text or JSON. Any object with the right methods qualifies."""

import json
from typing import Protocol, TypedDict

from eventcheck.checker import Issue, Report


class Writer(Protocol):
    """Anything with write(str): sys.stdout, io.StringIO, an open text file."""

    def write(self, text: str, /) -> object: ...


class Reporter(Protocol):
    def write(self, report: Report, out: Writer) -> None: ...


class IssueDict(TypedDict):
    line: int
    kind: str
    location: str
    message: str


class SummaryDict(TypedDict):
    lines: int
    valid: int
    bad_lines: list[int]
    by_type: dict[str, int]
    issues: list[IssueDict]


def issue_dict(issue: Issue) -> IssueDict:
    return {
        "line": issue.line,
        "kind": issue.kind,
        "location": issue.location,
        "message": issue.message,
    }


def summary_dict(report: Report) -> SummaryDict:
    return {
        "lines": report.lines,
        "valid": report.valid,
        "bad_lines": report.bad_lines,
        "by_type": dict(sorted(report.by_type.items())),
        "issues": [issue_dict(issue) for issue in report.issues],
    }


class TextReporter:
    def write(self, report: Report, out: Writer) -> None:
        for issue in report.issues:
            out.write(f"{issue}\n")
        types = ", ".join(f"{name}={count}" for name, count in sorted(report.by_type.items()))
        out.write(
            f"checked {report.lines} lines: {report.valid} valid, "
            f"{len(report.bad_lines)} bad ({types or 'no valid events'})\n"
        )


class JsonReporter:
    def write(self, report: Report, out: Writer) -> None:
        out.write(json.dumps(summary_dict(report), indent=2) + "\n")


REPORTERS: dict[str, Reporter] = {"text": TextReporter(), "json": JsonReporter()}
