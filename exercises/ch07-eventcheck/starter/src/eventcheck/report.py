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
    # TODO 1: add bad_lines (sorted line numbers), by_type (counts per event type)
    # and issues (a list of IssueDict), as the JSON output in the README shows.


def issue_dict(issue: Issue) -> IssueDict:
    return {
        "line": issue.line,
        "kind": issue.kind,
        "location": issue.location,
        "message": issue.message,
    }


def summary_dict(report: Report) -> SummaryDict:
    # TODO 2: build the SummaryDict; mypy checks every key and value type.
    raise NotImplementedError


class TextReporter:
    def write(self, report: Report, out: Writer) -> None:
        # TODO 3: one line per issue (str(issue)), then the summary line from the README.
        raise NotImplementedError


class JsonReporter:
    def write(self, report: Report, out: Writer) -> None:
        out.write(json.dumps(summary_dict(report), indent=2) + "\n")


REPORTERS: dict[str, Reporter] = {"text": TextReporter(), "json": JsonReporter()}
