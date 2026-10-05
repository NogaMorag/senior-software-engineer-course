"""Shared helpers: find the real bank, and write small banks into a temporary directory."""

import json
from collections.abc import Callable
from pathlib import Path

import pytest


def real_bank() -> Path:
    """data/questions.jsonl, found by walking up from this file (tests/ may sit one or two levels
    below the exercise root)."""
    for parent in Path(__file__).resolve().parents:
        candidate = parent / "data" / "questions.jsonl"
        if candidate.exists():
            return candidate
    raise FileNotFoundError("data/questions.jsonl not found above the tests directory")


def row(**overrides: object) -> dict[str, object]:
    """One valid question; keyword arguments replace fields."""
    base: dict[str, object] = {
        "id": 1,
        "category": "go",
        "chapter": 5,
        "question": "מתי Mutex ומתי channel?",
        "answer": "נעילה למבנה משותף, channel לזרימה. ההחלטה תלויה במי מחזיק את המצב.",
        "tradeoff": "פשטות מול זרימה מפורשת.",
        "follow_up": "ואם יש מספר אחד? atomic.",
    }
    base.update(overrides)
    return base


@pytest.fixture
def write_bank(tmp_path: Path) -> Callable[[list[dict[str, object]]], Path]:
    def write(rows: list[dict[str, object]]) -> Path:
        path = tmp_path / "questions.jsonl"
        lines = (json.dumps(r, ensure_ascii=False) for r in rows)
        path.write_text("\n".join(lines) + "\n", encoding="utf-8")
        return path

    return write
