"""The `quiz` command: `quiz check BANK` validates the file, `quiz ask BANK` quizzes you."""

from __future__ import annotations

import argparse
import random
import sys
from collections.abc import Callable, Sequence
from pathlib import Path
from typing import TextIO

from quiz.bank import CATEGORIES, BankError, Question, load


def pick(
    questions: Sequence[Question], categories: Sequence[str], n: int, rng: random.Random
) -> list[Question]:
    """Up to n distinct questions from the chosen categories (all categories when none chosen)."""
    pool = [q for q in questions if not categories or q.category in categories]
    return rng.sample(pool, min(n, len(pool)))


def ask(questions: Sequence[Question], read: Callable[[str], str], out: TextIO) -> list[Question]:
    """Show each question, reveal the answer on Enter, and return the ones she did not know.

    End of input (Ctrl-D) stops the round early; the questions not reached are not counted.
    """
    missed: list[Question] = []
    asked = 0
    for i, q in enumerate(questions, start=1):
        print(f"\n[{i}/{len(questions)}] #{q.id} {q.category}, chapter {q.chapter}", file=out)
        print(q.question, file=out)
        try:
            read("(Enter: show the answer) ")
            print(f"\n{q.answer}", file=out)
            print(f"Tradeoff: {q.tradeoff}", file=out)
            if q.follow_up:
                print(f"Follow-up: {q.follow_up}", file=out)
            knew = read("Did you know it? [y/N] ").strip().lower() in {"y", "yes"}
        except EOFError:
            print("\nstopped early", file=out)
            break
        asked += 1
        if not knew:
            missed.append(q)
    print(f"\nknown {asked - len(missed)} of {asked}", file=out)
    for q in missed:
        print(f"review #{q.id} ({q.category}, chapter {q.chapter})", file=out)
    return missed


def main(
    argv: Sequence[str] | None = None, read: Callable[[str], str] = input, out: TextIO | None = None
) -> int:
    out = out or sys.stdout
    parser = argparse.ArgumentParser(prog="quiz", description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    check = sub.add_parser("check", help="validate the bank against the schema")
    check.add_argument("bank", type=Path)
    quiz = sub.add_parser("ask", help="ask random questions, answer on Enter")
    quiz.add_argument("bank", type=Path)
    quiz.add_argument("-n", type=int, default=10, help="how many questions (default 10)")
    quiz.add_argument(
        "-c",
        "--category",
        action="append",
        default=[],
        choices=sorted(CATEGORIES),
        help="repeat to choose several; default all",
    )
    quiz.add_argument("--seed", type=int, help="fix the random choice, to repeat a round")
    args = parser.parse_args(argv)
    try:
        questions = load(args.bank)
    except (OSError, BankError) as err:
        print(f"quiz: {args.bank}:\n{err}", file=sys.stderr)
        return 1
    if args.command == "check":
        counts = {c: sum(q.category == c for q in questions) for c in CATEGORIES}
        print(f"{len(questions)} questions, schema OK", file=out)
        for category, count in counts.items():
            print(f"  {category:<17}{count:>4}", file=out)
        return 0
    chosen = pick(questions, args.category, args.n, random.Random(args.seed))
    ask(chosen, read, out)
    return 0
