"""Line-by-line validation. Pure functions over strings: no files, no printing."""

from collections import Counter
from collections.abc import Iterable
from dataclasses import dataclass, field
from typing import Literal, Self, override
from uuid import UUID

from pydantic import ValidationError

from eventcheck.models import EVENT_ADAPTER, EVENT_TYPES, Event

type IssueKind = Literal["json", "schema", "duplicate"]


@dataclass(frozen=True, slots=True)
class Issue:
    line: int
    kind: IssueKind
    location: str
    message: str

    @classmethod
    def from_validation_error(cls, line: int, error: ValidationError) -> list[Self]:
        issues = []
        for detail in error.errors(include_url=False):
            kind: IssueKind = "json" if detail["type"] == "json_invalid" else "schema"
            parts = detail["loc"]
            if parts and parts[0] in EVENT_TYPES:
                parts = parts[1:]  # drop the union tag: report "data.amount_cents"
            location = ".".join(str(part) for part in parts)
            issues.append(cls(line, kind, location, detail["msg"]))
        return issues

    @override
    def __str__(self) -> str:
        where = f" at {self.location}" if self.location else ""
        return f"line {self.line}: {self.kind}{where}: {self.message}"


@dataclass(frozen=True, slots=True)
class Ok[T]:
    value: T


@dataclass(frozen=True, slots=True)
class Err:
    issues: list[Issue]


type Outcome[T] = Ok[T] | Err


@dataclass(slots=True)
class Report:
    lines: int = 0
    valid: int = 0
    issues: list[Issue] = field(default_factory=list)
    by_type: Counter[str] = field(default_factory=Counter)

    @property
    def ok(self) -> bool:
        return not self.issues

    @property
    def bad_lines(self) -> list[int]:
        return sorted({issue.line for issue in self.issues})


def parse_line(line_no: int, raw: str) -> Outcome[Event]:
    try:
        return Ok(EVENT_ADAPTER.validate_json(raw))
    except ValidationError as error:
        return Err(Issue.from_validation_error(line_no, error))


def check_lines(lines: Iterable[str]) -> Report:
    report = Report()
    first_seen: dict[UUID, int] = {}
    for line_no, raw in enumerate(lines, start=1):
        report.lines = line_no
        if not raw.strip():
            continue
        match parse_line(line_no, raw.rstrip("\r\n")):
            case Err(issues):
                report.issues.extend(issues)
            case Ok(event) if event.id in first_seen:
                message = f"id {event.id} first seen on line {first_seen[event.id]}"
                report.issues.append(Issue(line_no, "duplicate", "id", message))
            case Ok(event):
                first_seen[event.id] = line_no
                report.valid += 1
                report.by_type[event.type] += 1
    return report
