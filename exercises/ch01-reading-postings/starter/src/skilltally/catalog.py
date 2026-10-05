"""The keyword table (skills.toml) and your self-assessment (profile.toml)."""

import tomllib
from dataclasses import dataclass
from pathlib import Path
from typing import Any


@dataclass(frozen=True)
class Skill:
    """One row of the keyword table.

    `keywords` match in any letter case; `exact` keywords match only as written
    (so "Go" counts and "go ahead" does not). `chapter` is the book chapter that
    teaches the skill, or 0 when the book does not.
    """

    name: str
    area: str
    chapter: int
    keywords: tuple[str, ...]
    exact: tuple[str, ...] = ()


@dataclass(frozen=True)
class Profile:
    """Where you stand today: every skill not in `have` or `partial` is missing."""

    have: frozenset[str] = frozenset()
    partial: frozenset[str] = frozenset()

    def status(self, skill: str) -> str:
        if skill in self.have:
            return "have"
        if skill in self.partial:
            return "partial"
        return "missing"


def _strings(entry: dict[str, Any], key: str, where: str) -> tuple[str, ...]:
    value = entry.get(key, [])
    if not isinstance(value, list) or not all(isinstance(v, str) and v.strip() for v in value):
        raise ValueError(f"{where}: '{key}' must be a list of non-empty strings")
    return tuple(value)


def parse_skills(data: dict[str, Any]) -> list[Skill]:
    """Validate the parsed TOML and build the skills, in file order."""
    entries = data.get("skill")
    if not isinstance(entries, list) or not entries:
        raise ValueError("skills: expected at least one [[skill]] table")
    skills: list[Skill] = []
    seen: set[str] = set()
    for index, entry in enumerate(entries, start=1):
        where = f"skill #{index}"
        name = entry.get("name")
        if not isinstance(name, str) or not name.strip():
            raise ValueError(f"{where}: 'name' must be a non-empty string")
        where = f"skill '{name}'"
        if name in seen:
            raise ValueError(f"{where}: defined twice")
        chapter = entry.get("chapter", 0)
        if not isinstance(chapter, int) or not 0 <= chapter <= 39:
            raise ValueError(f"{where}: 'chapter' must be a number from 0 to 39")
        keywords = _strings(entry, "keywords", where)
        exact = _strings(entry, "exact", where)
        if not keywords and not exact:
            raise ValueError(f"{where}: needs at least one keyword or exact keyword")
        skills.append(Skill(name, str(entry.get("area", "")), chapter, keywords, exact))
        seen.add(name)
    return skills


def load_skills(path: Path) -> list[Skill]:
    with path.open("rb") as f:
        return parse_skills(tomllib.load(f))


def load_profile(path: Path) -> Profile:
    with path.open("rb") as f:
        data = tomllib.load(f)
    have = _strings(data, "have", "profile")
    partial = _strings(data, "partial", "profile")
    both = set(have) & set(partial)
    if both:
        raise ValueError(f"profile: listed as both have and partial: {', '.join(sorted(both))}")
    return Profile(frozenset(have), frozenset(partial))
