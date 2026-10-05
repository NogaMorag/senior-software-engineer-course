"""Load the question bank (one JSON object per line) and check it against the schema."""

from __future__ import annotations

import json
import re
from dataclasses import dataclass
from pathlib import Path

# Category -> the chapters it may cite (inclusive). Titles live in tools/render_question_bank.py.
CATEGORIES: dict[str, tuple[int, int]] = {
    "go": (3, 6),
    "python": (7, 8),
    "backend": (9, 14),
    "distributed": (15, 18),
    "cloud-native": (19, 23),
    "operations": (24, 25),
    "quality-security": (26, 28),
    "senior-craft": (29, 34),
    "coding": (35, 35),
    "project": (37, 39),
}
FIELDS = ("id", "category", "chapter", "question", "answer", "tradeoff", "follow_up")
HEBREW = re.compile(r"[א-ת]")


@dataclass(frozen=True, slots=True)
class Question:
    id: int
    category: str
    chapter: int
    question: str
    answer: str
    tradeoff: str
    follow_up: str


class BankError(ValueError):
    """The bank file does not match the schema; `problems` lists every reason, one per line."""

    def __init__(self, problems: list[str]) -> None:
        super().__init__("\n".join(problems))
        self.problems = problems


def problems_in(rows: list[tuple[int, object]]) -> list[str]:
    """Every schema violation in the parsed rows, given as (line number, parsed JSON) pairs."""
    problems: list[str] = []
    seen: dict[int, int] = {}
    for line, row in rows:
        where = f"line {line}"
        if not isinstance(row, dict):
            problems.append(f"{where}: not a JSON object")
            continue
        keys = set(row)
        if missing := [f for f in FIELDS if f not in keys]:
            problems.append(f"{where}: missing {', '.join(missing)}")
        if extra := sorted(keys - set(FIELDS)):
            problems.append(f"{where}: unknown field {', '.join(extra)}")
        qid = row.get("id")
        if not isinstance(qid, int) or isinstance(qid, bool) or qid < 1:
            problems.append(f"{where}: id must be a positive integer")
        elif qid in seen:
            problems.append(f"{where}: id {qid} already used on line {seen[qid]}")
        else:
            seen[qid] = line
        category = row.get("category")
        chapter = row.get("chapter")
        if category not in CATEGORIES:
            problems.append(f"{where}: unknown category {category!r}")
        elif not isinstance(chapter, int) or isinstance(chapter, bool):
            problems.append(f"{where}: chapter must be an integer")
        else:
            low, high = CATEGORIES[category]
            if not low <= chapter <= high:
                problems.append(f"{where}: chapter {chapter} outside {category} ({low}-{high})")
        for field in ("question", "answer", "tradeoff"):
            value = row.get(field)
            if not isinstance(value, str) or not HEBREW.search(value):
                problems.append(f"{where}: {field} must be non-empty Hebrew text")
        follow_up = row.get("follow_up")
        if not isinstance(follow_up, str):
            problems.append(f"{where}: follow_up must be a string")
        elif follow_up and (not HEBREW.search(follow_up) or "?" not in follow_up):
            problems.append(f"{where}: follow_up must be empty or 'question? answer' in Hebrew")
    return problems


def load(path: Path) -> list[Question]:
    """Parse and validate the bank; raise BankError with every problem found."""
    rows: list[tuple[int, object]] = []
    problems: list[str] = []
    for line, text in enumerate(path.read_text(encoding="utf-8").splitlines(), start=1):
        if not text.strip():
            problems.append(f"line {line}: empty line")
            continue
        try:
            rows.append((line, json.loads(text)))
        except json.JSONDecodeError as err:
            problems.append(f"line {line}: invalid JSON ({err.msg})")
    problems += problems_in(rows)
    if problems:
        raise BankError(problems)
    return [Question(**row) for _, row in rows if isinstance(row, dict)]
