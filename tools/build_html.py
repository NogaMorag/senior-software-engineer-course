"""Render the Hebrew book (book/*.md) to right-to-left HTML pages under site/.

Usage (from the course root, no project needed):

    uv run --with markdown tools/build_html.py

Every chapter becomes site/<name>.html with proper RTL layout, left-to-right code blocks,
previous/next navigation and a table of contents. book/README.md becomes site/index.html.
"""

from __future__ import annotations

import html
import re
import shutil
import sys
from pathlib import Path

try:
    import markdown
except ImportError:  # pragma: no cover - friendly message instead of a traceback
    sys.exit("The 'markdown' package is missing. Run: uv run --with markdown tools/build_html.py")

ROOT = Path(__file__).resolve().parents[1]
BOOK = ROOT / "book"
SITE = ROOT / "site"
ASSETS = BOOK / "assets"

TEMPLATE = """<!DOCTYPE html>
<html lang="{lang}" dir="{direction}">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{title}</title>
<link rel="stylesheet" href="style.css">
</head>
<body>
<nav class="top">{nav}</nav>
<main class="book">
{toc}
{body}
</main>
<nav class="bottom">{nav}</nav>
<script type="module">
  // Render ```mermaid blocks in the browser (needs internet once; cached afterwards).
  const blocks = document.querySelectorAll("pre > code.language-mermaid");
  if (blocks.length) {{
    const mermaid = (await import("https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs")).default;
    for (const code of blocks) {{
      const div = document.createElement("div");
      div.className = "mermaid";
      div.textContent = code.textContent;
      code.parentElement.replaceWith(div);
    }}
    mermaid.initialize({{ startOnLoad: false, theme: "neutral" }});
    await mermaid.run();
  }}
</script>
</body>
</html>
"""

FRONT_MATTER = re.compile(r"\A---\s*\n(.*?)\n---\s*\n", re.S)
WRAPPER_OPEN = re.compile(r"^<div dir=\"rtl\"[^>]*>\s*$", re.M)
WRAPPER_CLOSE = re.compile(r"^</div>\s*$", re.M)


def parse_front_matter(text: str) -> tuple[dict[str, str], str]:
    match = FRONT_MATTER.match(text)
    if not match:
        return {}, text
    meta: dict[str, str] = {}
    for line in match.group(1).splitlines():
        if ":" in line:
            key, value = line.split(":", 1)
            meta[key.strip()] = value.strip().strip('"').strip("'")
    return meta, text[match.end():]


def strip_wrapper(body: str) -> str:
    body = WRAPPER_OPEN.sub("", body, count=1)
    # Remove only the last closing div (the wrapper), not divs inside the text.
    closes = list(WRAPPER_CLOSE.finditer(body))
    if closes:
        last = closes[-1]
        body = body[: last.start()] + body[last.end():]
    return body


def first_heading(body: str, fallback: str) -> str:
    match = re.search(r"^#\s+(.+)$", body, re.M)
    return match.group(1).strip() if match else fallback


def chapter_files() -> list[Path]:
    files = sorted(p for p in BOOK.glob("*.md") if p.name not in {"README.md", "CONVENTIONS.md"})
    files += sorted((BOOK / "appendices").glob("*.md")) if (BOOK / "appendices").exists() else []
    return files


def render(md_text: str) -> tuple[str, str]:
    converter = markdown.Markdown(
        extensions=["tables", "fenced_code", "toc", "attr_list", "sane_lists"],
        extension_configs={"toc": {"toc_depth": "2-3", "title": "תוכן הפרק"}},
    )
    body = converter.convert(md_text)
    toc = converter.toc if "<li>" in converter.toc else ""
    return body, toc


def nav_html(index: int, pages: list[tuple[str, str]]) -> str:
    parts = ['<a href="index.html">תוכן העניינים</a>']
    if index > 0:
        name, title = pages[index - 1]
        parts.append(f'<a href="{name}.html">← {html.escape(title)}</a>')
    if index < len(pages) - 1:
        name, title = pages[index + 1]
        parts.append(f'<a href="{name}.html">{html.escape(title)} →</a>')
    return " · ".join(parts)


def build() -> None:
    SITE.mkdir(exist_ok=True)
    if ASSETS.exists():
        for asset in ASSETS.iterdir():
            shutil.copy(asset, SITE / asset.name)

    files = chapter_files()
    pages: list[tuple[str, str]] = []
    bodies: list[tuple[dict[str, str], str]] = []
    for path in files:
        meta, body = parse_front_matter(path.read_text(encoding="utf-8"))
        body = strip_wrapper(body)
        title = meta.get("title") or first_heading(body, path.stem)
        pages.append((path.stem, title))
        bodies.append((meta, body))

    for index, (path, (meta, body)) in enumerate(zip(files, bodies)):
        rendered, toc = render(body)
        page = TEMPLATE.format(
            lang=meta.get("lang", "he"),
            direction=meta.get("direction", "rtl"),
            title=html.escape(pages[index][1]),
            nav=nav_html(index, pages),
            toc=toc,
            body=rendered,
        )
        (SITE / f"{path.stem}.html").write_text(page, encoding="utf-8")

    readme = BOOK / "README.md"
    if readme.exists():
        meta, body = parse_front_matter(readme.read_text(encoding="utf-8"))
        body = strip_wrapper(body)
        # Make chapter links point at the generated pages.
        body = re.sub(r"\]\(([^)]+?)\.md\)", lambda m: f"]({Path(m.group(1)).name}.html)", body)
        rendered, _ = render(body)
        page = TEMPLATE.format(
            lang="he", direction="rtl", title="תוכן העניינים", nav="", toc="", body=rendered
        )
        (SITE / "index.html").write_text(page, encoding="utf-8")

    print(f"built {len(files)} chapter page(s) + index into {SITE}")


if __name__ == "__main__":
    build()
