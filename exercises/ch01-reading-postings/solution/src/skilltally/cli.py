"""Command line: skilltally POSTINGS_DIR [--skills FILE] [--profile FILE]."""

import argparse
import sys
import tomllib
from collections.abc import Sequence
from pathlib import Path

from skilltally.catalog import Profile, load_profile, load_skills
from skilltally.report import render
from skilltally.tally import tally

EXIT_OK, EXIT_USAGE = 0, 2


def _parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="skilltally",
        description="Count which skills a folder of job postings (*.txt) asks for.",
    )
    parser.add_argument("postings", type=Path, help="folder with one posting per .txt file")
    parser.add_argument("--skills", type=Path, default=Path("skills.toml"), help="keyword table")
    parser.add_argument("--profile", type=Path, help="your have/partial lists (profile.toml)")
    return parser


def main(argv: Sequence[str] | None = None) -> int:
    args = _parser().parse_args(argv)
    files = sorted(args.postings.glob("*.txt")) if args.postings.is_dir() else []
    if not files:
        print(f"skilltally: no .txt postings in {args.postings}", file=sys.stderr)
        return EXIT_USAGE
    try:
        skills = load_skills(args.skills)
        profile: Profile | None = load_profile(args.profile) if args.profile else None
    except (OSError, tomllib.TOMLDecodeError, ValueError) as err:
        print(f"skilltally: {err}", file=sys.stderr)
        return EXIT_USAGE
    if profile:
        known = {skill.name for skill in skills}
        for name in sorted((profile.have | profile.partial) - known):
            print(f"skilltally: profile names an unknown skill: {name!r}", file=sys.stderr)
    postings = {path.name: path.read_text(encoding="utf-8") for path in files}
    print(render(tally(postings, skills), len(postings), profile, str(args.postings)), end="")
    return EXIT_OK
