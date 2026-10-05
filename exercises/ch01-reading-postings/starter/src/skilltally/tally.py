"""Find which postings mention which skills."""

import re
from collections.abc import Mapping, Sequence
from dataclasses import dataclass
from functools import cache

from skilltally.catalog import Skill

# A keyword must not touch another Latin letter, digit or underscore, so "java" does not
# match inside "JavaScript". Hebrew letters do not count as touching: postings write
# "בKubernetes" and "ב־Kubernetes", and both should count.
_WORD = "A-Za-z0-9_"


@cache
def keyword_pattern(keyword: str, ignore_case: bool) -> re.Pattern[str]:
    """Compile one keyword into a regular expression.

    - Spaces inside the keyword match any run of whitespace, line breaks included.
    - The boundary on each side applies only when the keyword starts or ends with a
      word character, so "C++" still matches in "C++17" and ".NET" in "ASP.NET".
    """
    # TODO (day 2): build the pattern in three parts and compile it.
    # 1. body: re.escape each space-separated part of the keyword, joined by r"\s+".
    # 2. before: a negative lookbehind for [A-Za-z0-9_] when the keyword starts with such a
    #    character; after: a negative lookahead when it ends with one. (_WORD holds the set.)
    # 3. re.compile(before + body + after), with re.IGNORECASE when ignore_case is true.
    # The table in tests/test_skilltally.py::test_mentions lists every case it must get right.
    raise NotImplementedError("keyword_pattern: see the TODO in src/skilltally/tally.py")


def mentions(text: str, skill: Skill) -> bool:
    """True when the text contains any of the skill's keywords."""
    return any(keyword_pattern(k, True).search(text) for k in skill.keywords) or any(
        keyword_pattern(k, False).search(text) for k in skill.exact
    )


@dataclass(frozen=True)
class Row:
    """One skill and the postings (file names) that mention it."""

    skill: Skill
    postings: tuple[str, ...]

    @property
    def count(self) -> int:
        return len(self.postings)


def tally(postings: Mapping[str, str], skills: Sequence[Skill]) -> list[Row]:
    """One row per skill, most-mentioned first; ties in table order. A posting counts once."""
    rows = [
        Row(skill, tuple(name for name, text in sorted(postings.items()) if mentions(text, skill)))
        for skill in skills
    ]
    return sorted(rows, key=lambda row: -row.count)  # sorted() is stable: ties keep table order
