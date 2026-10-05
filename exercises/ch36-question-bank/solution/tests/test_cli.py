import io
import random
from collections.abc import Callable, Iterator
from pathlib import Path

import pytest
from conftest import real_bank, row

from quiz.bank import load
from quiz.cli import ask, main, pick

WriteBank = Callable[[list[dict[str, object]]], Path]


def scripted(*answers: str) -> Callable[[str], str]:
    """An input() stand-in that replays answers, then raises EOFError like Ctrl-D."""
    it: Iterator[str] = iter(answers)

    def read(_prompt: str) -> str:
        try:
            return next(it)
        except StopIteration:
            raise EOFError from None

    return read


def three(write_bank: WriteBank) -> Path:
    return write_bank(
        [
            row(id=1, category="go", chapter=3),
            row(id=2, category="backend", chapter=13),
            row(id=3, category="backend", chapter=14),
        ]
    )


def test_pick_only_from_chosen_categories() -> None:
    questions = load(real_bank())
    chosen = pick(questions, ["distributed"], 5, random.Random(1))
    assert len(chosen) == 5
    assert {q.category for q in chosen} == {"distributed"}
    assert len({q.id for q in chosen}) == 5


def test_pick_is_repeatable_with_a_seed() -> None:
    questions = load(real_bank())
    first = pick(questions, [], 10, random.Random(7))
    again = pick(questions, [], 10, random.Random(7))
    assert [q.id for q in first] == [q.id for q in again]


def test_pick_caps_at_the_pool_size(write_bank: WriteBank) -> None:
    questions = load(three(write_bank))
    assert len(pick(questions, ["backend"], 10, random.Random(0))) == 2


def test_ask_reveals_answer_and_lists_misses(write_bank: WriteBank) -> None:
    questions = load(three(write_bank))
    out = io.StringIO()
    missed = ask(questions, scripted("", "y", "", "n", "", ""), out)
    assert [q.id for q in missed] == [2, 3]
    text = out.getvalue()
    assert text.count("Tradeoff: ") == 3
    assert "known 1 of 3" in text
    assert "review #2 (backend, chapter 13)" in text


def test_end_of_input_stops_early(write_bank: WriteBank) -> None:
    questions = load(three(write_bank))
    out = io.StringIO()
    missed = ask(questions, scripted("", "y"), out)
    assert missed == []
    assert "stopped early" in out.getvalue()
    assert "known 1 of 1" in out.getvalue()


def test_main_check_prints_counts(capsys: pytest.CaptureFixture[str]) -> None:
    assert main(["check", str(real_bank())]) == 0
    out = capsys.readouterr().out
    assert "schema OK" in out
    assert "senior-craft" in out


def test_main_check_fails_on_a_bad_bank(
    write_bank: WriteBank, capsys: pytest.CaptureFixture[str]
) -> None:
    path = write_bank([row(), row()])
    assert main(["check", str(path)]) == 1
    assert "already used" in capsys.readouterr().err


def test_main_ask_uses_the_given_input(write_bank: WriteBank) -> None:
    out = io.StringIO()
    code = main(
        ["ask", str(three(write_bank)), "-n", "1", "-c", "go"], read=scripted("", "y"), out=out
    )
    assert code == 0
    assert "#1 go, chapter 3" in out.getvalue()
