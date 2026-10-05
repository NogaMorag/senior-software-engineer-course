"""Line-by-line validation. Pure functions over strings: no files, no printing."""

from collections import Counter
from collections.abc import Iterable
from dataclasses import dataclass, field
from typing import Literal, Self, override

from pydantic import ValidationError

from eventcheck.models import Event

# TODO 7-9 need EVENT_ADAPTER and EVENT_TYPES from eventcheck.models, and UUID from uuid.

type IssueKind = Literal["json", "schema", "duplicate"]


@dataclass(frozen=True, slots=True)
class Issue:
    line: int
    kind: IssueKind
    location: str
    message: str

    @classmethod
    def from_validation_error(cls, line: int, error: ValidationError) -> list[Self]:
        # TODO 7: one Issue per entry of error.errors(include_url=False).
        # kind is "json" when the entry's type is "json_invalid", else "schema";
        # location joins the entry's "loc" with "." after dropping a leading tag from EVENT_TYPES.
        raise NotImplementedError

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
    # TODO 8: EVENT_ADAPTER.validate_json(raw) wrapped in Ok, or Err with the issues.
    raise NotImplementedError


def check_lines(lines: Iterable[str]) -> Report:
    # TODO 9: number lines from 1, skip blank ones (still counted in report.lines),
    # strip the newline before parse_line, record issues, and report a repeated id
    # as a "duplicate" issue at location "id" naming the line where it was first seen.
    raise NotImplementedError
