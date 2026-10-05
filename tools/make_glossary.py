"""Build appendix A (the glossary) from every chapter's "## מונחים לזכור" table. Single source of truth: the chapters.

    uv run --no-project python tools/make_glossary.py
"""

from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
BOOK = ROOT / "book"
OUT = BOOK / "appendices" / "a-glossary.md"
ROW = re.compile(r"^\| ([^|]+) \| ([^|]+) \| ([^|]+) \|$")


def collect() -> dict[str, dict]:
    entries: dict[str, dict] = {}
    for path in sorted(BOOK.glob("[0-9][0-9]-*.md")):
        number = int(path.name[:2])
        text = path.read_text(encoding="utf-8")
        if "## מונחים לזכור" not in text:
            continue
        section = text.split("## מונחים לזכור", 1)[1].split("\n## ", 1)[0]
        for line in section.splitlines():
            m = ROW.match(line.strip())
            if not m or m.group(1).strip() in ("מונח בעברית", "---"):
                continue
            he, en, definition = (g.strip() for g in m.groups())
            key = en.lower()
            if key in entries:
                entries[key]["chapters"].append(number)
            else:
                entries[key] = {"he": he, "en": en, "definition": definition, "chapters": [number]}
    return entries


def render(entries: dict[str, dict]) -> str:
    rows = sorted(entries.values(), key=lambda e: e["en"].lower())
    lines = [
        "---", "direction: rtl", "lang: he", 'title: "נספח א: מילון מונחים"', "---", "", '<div dir="rtl" lang="he">', "",
        "# נספח א: מילון מונחים",
        "",
        f"{len(rows)} מונחים, נאספים אוטומטית מטבלאות „מונחים לזכור” של כל הפרקים (הכלי: `tools/make_glossary.py`).",
        "הסדר לפי המונח באנגלית. העמודה האחרונה מפנה לפרק שבו המונח הוגדר לראשונה, ובסוגריים לפרקים שחזרו עליו.",
        "", "| באנגלית | בעברית | במשפט אחד | פרק |", "| --- | --- | --- | --- |",
    ]
    for e in rows:
        first, *rest = e["chapters"]
        ref = f"{first}" + (f" ({', '.join(str(c) for c in rest)})" if rest else "")
        lines.append(f"| {e['en']} | {e['he']} | {e['definition']} | {ref} |")
    lines += ["", "</div>", ""]
    return "\n".join(lines)


def main() -> None:
    entries = collect()
    OUT.parent.mkdir(exist_ok=True)
    OUT.write_text(render(entries), encoding="utf-8")
    print(f"glossary: {len(entries)} terms -> {OUT.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
