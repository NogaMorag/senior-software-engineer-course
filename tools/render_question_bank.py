"""Render the chapter-36 question bank section from the JSONL data (single source of truth).

    uv run --no-project python tools/render_question_bank.py

Replaces everything between the markers <!-- bank:start --> and <!-- bank:end --> in book/36-system-design-interviews.md.
"""

from __future__ import annotations

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
BANK = ROOT / "exercises/ch36-question-bank/data/questions.jsonl"
CHAPTER = ROOT / "book/36-system-design-interviews.md"
TITLES = {
    "go": "Go (פרקים 3 עד 6)",
    "python": "Python (פרקים 7 ו־8)",
    "backend": "יסודות backend (פרקים 9 עד 14)",
    "distributed": "מערכות מבוזרות (פרקים 15 עד 18)",
    "cloud-native": "Cloud-native (פרקים 19 עד 23)",
    "operations": "תצפיתיות ותפעול (פרקים 24 ו־25)",
    "quality-security": "איכות ואבטחה (פרקים 26 עד 28)",
    "senior-craft": "אומנות הבכירות (פרקים 29 עד 33)",
    "coding": "תרגילי קוד (פרק 35)",
    "project": "„ספרי לי על הפרויקט שלך” (פרקים 37 עד 39)",
}
PART_ONE = {"go", "python", "backend", "distributed", "cloud-native"}


def render() -> str:
    rows = [json.loads(line) for line in BANK.read_text(encoding="utf-8").splitlines() if line.strip()]
    out: list[str] = []
    for part, cats in (("### חלק ראשון (יום 96)", [c for c in TITLES if c in PART_ONE]), ("### חלק שני (יום 97)", [c for c in TITLES if c not in PART_ONE])):
        out += [part, ""]
        for cat in cats:
            out += [f"#### {TITLES[cat]}", ""]
            for row in (r for r in rows if r["category"] == cat):
                out.append(f"**{row['id']}. {row['question']}** (פרק {row['chapter']})")
                out.append("")
                out.append(row["answer"])
                if row.get("tradeoff"):
                    out.append("")
                    out.append(f"*שורת הפשרה:* {row['tradeoff']}")
                if row.get("follow_up"):
                    out.append("")
                    out.append(f"**המשך:** {row['follow_up']}")
                out.append("")
    return "\n".join(out).rstrip() + "\n"


def main() -> None:
    text = CHAPTER.read_text(encoding="utf-8")
    start, end = "<!-- bank:start -->", "<!-- bank:end -->"
    a, b = text.index(start) + len(start), text.index(end)
    CHAPTER.write_text(text[:a] + "\n\n" + render() + "\n" + text[b:], encoding="utf-8")
    print("question bank rendered into", CHAPTER.name)


if __name__ == "__main__":
    main()
