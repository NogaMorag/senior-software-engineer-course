from collections.abc import Callable
from pathlib import Path

import pytest
from conftest import real_bank, row

from quiz.bank import CATEGORIES, BankError, load

WriteBank = Callable[[list[dict[str, object]]], Path]


def test_real_bank_is_valid() -> None:
    questions = load(real_bank())
    assert 120 <= len(questions) <= 160


def test_real_bank_covers_every_category() -> None:
    present = {q.category for q in load(real_bank())}
    assert present == set(CATEGORIES)


def test_real_bank_cites_every_chapter_it_should() -> None:
    cited = {q.chapter for q in load(real_bank())}
    assert set(range(3, 36)) | {37} <= cited


def test_valid_row_loads(write_bank: WriteBank) -> None:
    [q] = load(write_bank([row()]))
    assert (q.id, q.category, q.chapter) == (1, "go", 5)


@pytest.mark.parametrize(
    ("rows", "expected"),
    [
        ([row(), row()], "id 1 already used on line 1"),
        ([row(id="7")], "id must be a positive integer"),
        ([row(id=True)], "id must be a positive integer"),
        ([row(category="frontend")], "unknown category 'frontend'"),
        ([row(chapter=40)], "chapter 40 outside go (3-6)"),
        ([row(category="coding", chapter=34)], "chapter 34 outside coding (35-35)"),
        ([row(question="")], "question must be non-empty Hebrew text"),
        ([row(answer="English only, no Hebrew.")], "answer must be non-empty Hebrew text"),
        ([row(follow_up="ואם יש מספר אחד")], "follow_up must be empty or"),
        ([row(extra="x")], "unknown field extra"),
    ],
)
def test_schema_violations(
    write_bank: WriteBank, rows: list[dict[str, object]], expected: str
) -> None:
    with pytest.raises(BankError) as err:
        load(write_bank(rows))
    assert any(expected in p for p in err.value.problems)


def test_missing_field_is_named(write_bank: WriteBank) -> None:
    bad = row()
    del bad["tradeoff"]
    with pytest.raises(BankError, match="missing tradeoff"):
        load(write_bank([bad]))


def test_empty_follow_up_is_allowed(write_bank: WriteBank) -> None:
    assert load(write_bank([row(follow_up="")]))[0].follow_up == ""


def test_every_problem_is_reported_with_its_line(tmp_path: Path) -> None:
    path = tmp_path / "questions.jsonl"
    path.write_text('{"id": 1}\nnot json\n\n', encoding="utf-8")
    with pytest.raises(BankError) as err:
        load(path)
    problems = err.value.problems
    assert any(p.startswith("line 1: missing") for p in problems)
    assert any(p.startswith("line 2: invalid JSON") for p in problems)
    assert "line 3: empty line" in problems
