"""Generate obsidian/<slug>.md notes from each chapter's "## פתק Obsidian" section (single source of truth).

    uv run --no-project python tools/notes_from_chapters.py 27 28 29

Prev/next links follow the chapter order in book/README.md. The chapter's note section must contain
"זה תוכנו:" followed by a blank line and then the note body (bullets + one ```text block).
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
BOOK = ROOT / "book"
NOTES = ROOT / "obsidian"


def chapter_order() -> list[str]:
    """Slugs in TOC order, from the links in book/README.md."""
    text = (BOOK / "README.md").read_text(encoding="utf-8")
    slugs = re.findall(r"\| פרק \d+:[^|]*\| \[?`([0-9]{2}-[a-z0-9-]+)\.md`", text)
    return slugs


def title_of(text: str) -> str:
    match = re.search(r'^title: "פרק \d+: (.+)"$', text, re.MULTILINE)
    return match.group(1) if match else "?"


def make_note(number: int) -> Path:
    order = chapter_order()
    slug = next(s for s in order if s.startswith(f"{number:02d}-"))
    index = order.index(slug)
    prev_slug = order[index - 1] if index > 0 else None
    next_slug = order[index + 1] if index + 1 < len(order) else None
    text = (BOOK / f"{slug}.md").read_text(encoding="utf-8")
    start = text.index("## פתק Obsidian")
    end = text.index("## תרגיל", start)
    section = text[start:end]
    # The note section opens with one intro paragraph (where the file lives, how the note is built); the note
    # proper is everything after it. Do not key on the intro's wording: it varies between chapters.
    paragraphs = section.split("\n\n")
    body = "\n\n".join(paragraphs[2:]).strip() + "\n"  # [0] = heading, [1] = intro paragraph
    links = " · ".join(f"[[{s}]]" for s in (prev_slug, next_slug) if s)
    note = f"---\ndirection: rtl\ntags: [senior-course, book]\nchapter: {number}\n---\n# {title_of(text)}\n\n{body}\nקישורים: {links}\n"
    path = NOTES / f"{slug}.md"
    path.write_text(note, encoding="utf-8")
    return path


def main(argv: list[str]) -> int:
    for arg in argv:
        print("wrote", make_note(int(arg)))
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
