"""Command line: `eventcheck check FILE` and `eventcheck generate COUNT OUT`."""

import argparse
import os
import sys
from collections.abc import Sequence

from eventcheck.checker import check_lines
from eventcheck.generate import generate_lines
from eventcheck.report import REPORTERS

EXIT_OK, EXIT_ISSUES, EXIT_USAGE = 0, 1, 2


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="eventcheck", description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    check = commands.add_parser("check", help="validate a JSONL file ('-' reads stdin)")
    check.add_argument("path")
    check.add_argument("--format", choices=sorted(REPORTERS), default=None)
    gen = commands.add_parser("generate", help="write a deterministic sample file")
    gen.add_argument("count", type=int)
    gen.add_argument("out")
    gen.add_argument("--seed", type=int, default=7)
    gen.add_argument("--bad-every", type=int, default=0)
    return parser


def run_check(path: str, fmt: str | None) -> int:
    fmt = fmt or os.environ.get("EVENTCHECK_FORMAT", "text")
    reporter = REPORTERS.get(fmt)
    if reporter is None:
        print(f"eventcheck: unknown format {fmt!r}", file=sys.stderr)
        return EXIT_USAGE
    if path == "-":
        report = check_lines(sys.stdin)
    else:
        try:
            with open(path, encoding="utf-8") as lines:
                report = check_lines(lines)
        except OSError as error:
            print(f"eventcheck: {error}", file=sys.stderr)
            return EXIT_USAGE
    reporter.write(report, sys.stdout)
    return EXIT_OK if report.ok else EXIT_ISSUES


def run_generate(count: int, out: str, seed: int, bad_every: int) -> int:
    with open(out, "w", encoding="utf-8") as handle:
        for line in generate_lines(count, seed=seed, bad_every=bad_every):
            handle.write(line + "\n")
    return EXIT_OK


def main(argv: Sequence[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    if args.command == "check":
        return run_check(args.path, args.format)
    return run_generate(args.count, args.out, args.seed, args.bad_every)
