"""Wrap every top-level fenced code block in a left-to-right container.

Inside a right-to-left Hebrew document, Markdown viewers that honor the outer `dir="rtl"`
render code blocks right-aligned and scramble indentation. Wrapping each block in
`<div dir="ltr" align="left">` ... `</div>` keeps code left-to-right in every renderer
(GitHub, VS Code, Obsidian reading view, the generated HTML).

Idempotent: blocks that are already wrapped are left alone. Fences inside blockquotes or
list items (indented or prefixed with `>`) are not touched.

Usage (from the course root):

    uv run --no-project python tools/wrap_ltr_blocks.py [paths...]

With no paths it processes book/*.md, book/appendices/*.md, exercises/**/README.md and
obsidian/*.md.
"""

from __future__ import annotations

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
OPEN_TAG = '<div dir="ltr" align="left">'
CLOSE_TAG = "</div>"
SKIP = {"CONVENTIONS.md"}


def default_targets() -> list[Path]:
    targets = [p for p in (ROOT / "book").glob("*.md") if p.name not in SKIP]
    targets += sorted((ROOT / "book" / "appendices").glob("*.md"))
    targets += sorted((ROOT / "exercises").glob("**/README.md"))
    targets += sorted((ROOT / "obsidian").glob("*.md"))
    return targets


def wrap(text: str) -> tuple[str, int]:
    lines = text.split("\n")
    out: list[str] = []
    in_fence = False
    wrapped_here = False
    added = 0
    for i, line in enumerate(lines):
        is_fence = line.startswith("```")
        if is_fence and not in_fence:
            in_fence = True
            # Look back past blank lines for an existing wrapper.
            j = len(out) - 1
            while j >= 0 and out[j].strip() == "":
                j -= 1
            already = j >= 0 and out[j].strip().startswith('<div dir="ltr"')
            wrapped_here = not already
            if wrapped_here:
                if out and out[-1].strip() != "":
                    out.append("")
                out.append(OPEN_TAG)
                out.append("")
                added += 1
            out.append(line)
            continue
        if is_fence and in_fence:
            in_fence = False
            out.append(line)
            if wrapped_here:
                out.append("")
                out.append(CLOSE_TAG)
                # Avoid a double blank line if the source already has one next.
                if i + 1 < len(lines) and lines[i + 1].strip() != "":
                    out.append("")
            wrapped_here = False
            continue
        out.append(line)
    return "\n".join(out), added


def main(argv: list[str]) -> None:
    paths = [Path(a) for a in argv] or default_targets()
    total = 0
    for path in paths:
        if not path.exists():
            print(f"skip (missing): {path}")
            continue
        text = path.read_text(encoding="utf-8")
        new_text, added = wrap(text)
        if added:
            path.write_text(new_text, encoding="utf-8")
        total += added
        print(f"{path.relative_to(ROOT) if path.is_relative_to(ROOT) else path}: wrapped {added}")
    print(f"total blocks wrapped: {total}")


if __name__ == "__main__":
    main(sys.argv[1:])
