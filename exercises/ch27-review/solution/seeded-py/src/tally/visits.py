"""Each log line is ``name,language``; blank lines and comments (``#``) are skipped."""

from collections.abc import Iterable


def parse_line(line: str) -> tuple[str, str] | None:
    """Return (name, language) for a data line, or None for a line to skip."""
    line = line.strip()
    if not line or line.startswith("#"):
        return None
    try:
        name, language = line.split(",")
    except ValueError:
        return None
    return name.strip().lower(), language.strip().lower()


def tally(lines: Iterable[str], counts: dict[str, int] | None = None) -> dict[str, int]:
    """Count visits per name, adding to ``counts`` when one is given."""
    counts = {} if counts is None else counts
    for line in lines:
        parsed = parse_line(line)
        if parsed is not None:
            name, _ = parsed
            counts[name] = counts.get(name, 0) + 1
    return counts


def top(counts: dict[str, int]) -> str | None:
    """The most frequent visitor (alphabetically first on a tie), or None when empty."""
    if not counts:
        return None
    return min(counts, key=lambda name: (-counts[name], name))
