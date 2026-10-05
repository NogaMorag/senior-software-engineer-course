"""Render the gap matrix as a Markdown table you can paste into Obsidian."""

from collections.abc import Sequence

from skilltally.catalog import Profile
from skilltally.tally import Row


def render(rows: Sequence[Row], total: int, profile: Profile | None, source: str) -> str:
    """The matrix: every mentioned skill, then gaps, then the skills no posting mentioned."""
    lines = [
        f"# skilltally: {total} postings in {source}",
        "",
        "| skill | postings | share | you | chapter |",
        "| --- | ---: | ---: | --- | --- |",
    ]
    seen = [row for row in rows if row.count > 0]
    for row in seen:
        you = profile.status(row.skill.name) if profile else "?"
        chapter = str(row.skill.chapter) if row.skill.chapter else "-"
        share = round(100 * row.count / total) if total else 0
        lines.append(f"| {row.skill.name} | {row.count}/{total} | {share}% | {you} | {chapter} |")
    lines.append("")
    if profile:
        gaps = [
            row.skill.name
            for row in seen
            if profile.status(row.skill.name) == "missing" and 2 * row.count >= total
        ]
        lines.append(f"Gaps (missing, in at least half the postings): {', '.join(gaps) or 'none'}")
    unseen = [row.skill.name for row in rows if row.count == 0]
    lines.append(f"Not seen in any posting: {', '.join(unseen) or 'none'}")
    return "\n".join(lines) + "\n"
