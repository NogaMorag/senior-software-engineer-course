"""Consistency audit over the whole book. Read-only; prints findings grouped by kind.

    uv run --no-project python tools/consistency_check.py

Checks: every TOC chapter file exists and is marked done; every chapter has the fixed apparatus in order; every
fenced block is wrapped in an LTR div; every relative link resolves; every obsidian note exists and its [[links]]
resolve; every exercise dir referenced by a chapter has README and solution/tests/data, and every exercise README has
the template sections; appendix F has day blocks 0..100 with six fields each; chapter cross-references ("פרק N")
point at existing chapters; day references ("יום N") are within 0..100; exercise day blocks are in order; every
section reference in both schedules ("פרק N, סעיפים a.b עד a.c" in appendix F, "ch N, sections a.b–a.c" in the English
schedule) exists in chapter N; no employer terms anywhere in the book or exercises.
Adapted from book 1 (ai-engineering-course) on 2026-10-05.
"""

from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
BOOK, NOTES, EXERCISES, STAGE1 = ROOT / "book", ROOT / "obsidian", ROOT / "exercises", ROOT / "00-stage1"
LAST_DAY = 100
APPARATUS = ["## מה נלמד בפרק", "## ידע מקדים", "## סיכום", "## מונחים לזכור", "## טעויות נפוצות", "## שאלות ראיון", "## פתק Obsidian", "## תרגיל", "## העשרה אופציונלית"]
FORBIDDEN = re.compile(r"drivenets|\bdnos\b|nmorag|ACO-workspace|dnai-orchestration|\bAICO-[0-9]|kvm59", re.IGNORECASE)
PROJECT_CHAPTERS = (37, 38, 39)  # they share the milestone days with the parts
SCHEDULE_FIELDS = ["**מטרה:**", "**קריאה:**", "**זמן:**", "**תרגיל:**", "**קריטריון סיום:**", "**מה את יכולה להסביר:**"]


def chapters() -> list[Path]:
    return sorted(BOOK.glob("[0-9][0-9]-*.md"))


def sections_of(number: int) -> set[str]:
    """The numbered `## N.k` sections that chapter N actually has."""
    path = next((p for p in chapters() if int(p.name[:2]) == number), None)
    if path is None:
        return set()
    return set(re.findall(rf"^## ({number}\.\d+) ", path.read_text(encoding="utf-8"), flags=re.MULTILINE))


def check_toc(findings: list[str]) -> None:
    toc = (BOOK / "README.md").read_text(encoding="utf-8")
    for m in re.finditer(r"\| (פרק \d+|[א-ו]): [^|]+ \| (\[?`([^`]+)`\]?(?:\([^)]+\))?) \| (✅|⏳|🛠) \|", toc):
        label, _, file, status = m.groups()
        path = BOOK / file
        if not path.exists():
            if status == "✅":
                findings.append(f"toc: {label} marked done but {file} is missing")
        elif status != "✅":
            findings.append(f"toc: {label} exists but is marked {status}")


def check_chapters(findings: list[str]) -> None:
    slugs = {p.stem for p in chapters()}
    numbers = {int(p.name[:2]) for p in chapters()}
    for path in chapters():
        text = path.read_text(encoding="utf-8")
        n = int(path.name[:2])
        if n > 0:
            positions = [text.find(h + "\n") for h in APPARATUS]
            for h, pos in zip(APPARATUS, positions):
                if pos < 0:
                    findings.append(f"{path.name}: missing {h}")
            present = [p for p in positions if p >= 0]
            if present != sorted(present):
                findings.append(f"{path.name}: apparatus headings out of order")
        lines = text.splitlines()
        fences = sum(1 for l in lines if l.startswith("```")) // 2
        wraps = sum(1 for l in lines if l.strip() == '<div dir="ltr" align="left">')
        if fences != wraps:
            findings.append(f"{path.name}: {fences} fences but {wraps} ltr wrappers")
        for link in re.findall(r"\]\(((?:\.\./|\./)?[^)#\s]+\.md)\)", text):
            if not (path.parent / link).exists():
                findings.append(f"{path.name}: broken link {link}")
        for ref in {int(x) for x in re.findall(r"פרק (\d+)", text)}:
            if ref not in numbers and ref <= 39:
                findings.append(f"{path.name}: refers to chapter {ref} which is not written yet")
            elif ref > 39:
                findings.append(f"{path.name}: refers to chapter {ref} which does not exist")
        for day in {int(x) for x in re.findall(r"יום (\d+)", text)}:
            if not 0 <= day <= LAST_DAY:
                findings.append(f"{path.name}: day {day} outside 0..{LAST_DAY}")
        if FORBIDDEN.search(text):
            findings.append(f"{path.name}: forbidden term")
        if n > 0:
            note = NOTES / f"{path.stem}.md"
            if not note.exists():
                findings.append(f"{path.name}: no obsidian note")
            else:
                note_text = note.read_text(encoding="utf-8")
                for target in re.findall(r"\[\[([^\]]+)\]\]", note_text):
                    if target not in slugs:
                        findings.append(f"obsidian/{note.name}: link to missing note {target}")
                if note_text.count("```text") != 1:
                    findings.append(f"obsidian/{note.name}: needs exactly one ```text cheat sheet")
        for ex in set(re.findall(r"exercises/([a-z0-9-]+)/README\.md", text)):
            folder = EXERCISES / ex
            if not folder.exists():
                findings.append(f"{path.name}: exercise folder {ex} missing")
                continue
            if not (folder / "README.md").exists():
                findings.append(f"{ex}: no README")
            if ex.startswith("ch") and not any((folder / sub).exists() for sub in ("solution", "tests", "data", "templates")):
                findings.append(f"{ex}: no solution/tests/data/templates")


def check_schedule_alignment(findings: list[str]) -> None:
    """The **זמן:** line of each chapter exercise names its days; blocks must be in order (a shared boundary day is allowed)."""
    last_end = 0
    for path in chapters():
        n = int(path.name[:2])
        if n < 1 or n in PROJECT_CHAPTERS:
            continue
        text = path.read_text(encoding="utf-8")
        section = text.split("## תרגיל\n", 1)[1] if "## תרגיל\n" in text else ""
        time_line = next((l for l in section.splitlines() if "**זמן:**" in l), "")
        found = [int(d) for pair in re.findall(r"(?:יום|ימים) (\d+)(?: עד (\d+)| ו־(\d+))?", time_line) for d in pair if d]
        found = [d for d in found if d != 0 and d != 57] if n == 2 else found  # ch 2 also owns the install days
        if not found:
            continue
        start, end = min(found), max(found)
        if start < last_end:
            findings.append(f"{path.name}: exercise days {start}-{end} overlap previous chapters (last end {last_end})")
        last_end = max(last_end, end)


def check_section_refs(findings: list[str], name: str, text: str, pattern: re.Pattern[str]) -> None:
    for m in pattern.finditer(text):
        chapter = int(m.group(1))
        have = sections_of(chapter)
        if not have:
            continue  # chapter not written yet; the TOC check covers it
        for ref in (g for g in m.groups()[1:] if g):
            if not ref.startswith(f"{chapter}."):
                findings.append(f"{name}: section {ref} cited under chapter {chapter}")
            elif ref not in have:
                findings.append(f"{name}: chapter {chapter} has no section {ref}")


HE_REF = re.compile(r"פרק (\d+), (?:סעיפים|סעיף) (\d+\.\d+)(?: עד (\d+\.\d+)| ו־(\d+\.\d+))?")
EN_REF = re.compile(r"ch (\d+), sections? (\d+\.\d+)(?:[–-](\d+\.\d+))?(?:,? (?:and|&) (\d+\.\d+))?")


def check_hebrew_schedule(findings: list[str]) -> None:
    """Appendix F must carry one '### יום N' block for every day 0..100, in order, with the six fields."""
    path = BOOK / "appendices" / "f-daily-schedule.md"
    if not path.exists():
        findings.append("appendices/f-daily-schedule.md: missing")
        return
    text = path.read_text(encoding="utf-8")
    days = [int(d) for d in re.findall(r"^### יום (\d+)", text, flags=re.MULTILINE)]
    if days != list(range(0, LAST_DAY + 1)):
        missing = sorted(set(range(0, LAST_DAY + 1)) - set(days))
        extra = sorted(set(days) - set(range(0, LAST_DAY + 1)))
        findings.append(f"appendices/f-daily-schedule.md: {len(days)} day blocks (missing {missing[:5]}, extra {extra[:5]}, in order: {days == sorted(days)})")
    for block in re.split(r"^### ", text, flags=re.MULTILINE)[1:]:
        head = block.splitlines()[0]
        for field in SCHEDULE_FIELDS:
            if field not in block:
                findings.append(f"appendices/f-daily-schedule.md: '{head}' missing {field}")
    check_section_refs(findings, "appendix F", text, HE_REF)
    en = STAGE1 / "05-daily-schedule.md"
    if en.exists():
        check_section_refs(findings, "05-daily-schedule.md", en.read_text(encoding="utf-8"), EN_REF)


EXERCISE_SECTIONS = ["## מה בתיקייה הזאת", "## פלט צפוי", "## רשימת בדיקה", "## טעויות נפוצות ותיקונים", "## רמזים", "## הפתרון", "## מה התרגיל מדגים"]
NON_CODING = {"ch31-senior-craft", "ch34-interview-loop"}


def check_exercise_readmes(findings: list[str]) -> None:
    """Every exercise README carries the template sections (13 items for coding exercises)."""
    for folder in sorted(EXERCISES.glob("ch*/")):
        readme = folder / "README.md"
        if not readme.exists():
            findings.append(f"{folder.name}: no README")
            continue
        text = readme.read_text(encoding="utf-8")
        required = ["**מטרה:**", "ידע מקדים", "## מה התרגיל מדגים", "## רשימת בדיקה"]
        if folder.name not in NON_CODING:
            required += EXERCISE_SECTIONS
        if not re.search(r"\*\*זמן(?: משוער)?:\*\*", text):
            findings.append(f"{folder.name}/README.md: no time estimate")
        for section in required:
            if section not in text:
                findings.append(f"{folder.name}/README.md: missing {section}")
        lines = text.splitlines()
        if sum(1 for l in lines if l.startswith("```")) // 2 != sum(1 for l in lines if l.strip() == '<div dir="ltr" align="left">'):
            findings.append(f"{folder.name}/README.md: fence/wrapper mismatch")
        mistakes = text.split("## טעויות נפוצות ותיקונים", 1)[1].split("\n## ", 1)[0] if "## טעויות נפוצות ותיקונים" in text else ""
        if folder.name not in NON_CODING and mistakes and not 3 <= sum(1 for l in mistakes.splitlines() if l.startswith("- ")) <= 6:
            findings.append(f"{folder.name}/README.md: common mistakes must be 3-6 bullets")


def check_forbidden_everywhere(findings: list[str]) -> None:
    for path in list(EXERCISES.rglob("*")) + list(STAGE1.glob("*.md")) + list(NOTES.glob("*.md")):
        if path.is_file() and path.suffix in {".md", ".go", ".py", ".yaml", ".yml", ".json", ".jsonl", ".tf", ".js", ".sql", ".toml", ".sh", ".proto"}:
            if any(part in {".venv", "node_modules", ".terraform"} for part in path.parts):
                continue
            if FORBIDDEN.search(path.read_text(encoding="utf-8", errors="ignore")):
                findings.append(f"{path.relative_to(ROOT)}: forbidden term")


def main() -> int:
    findings: list[str] = []
    check_toc(findings)
    check_chapters(findings)
    check_schedule_alignment(findings)
    check_hebrew_schedule(findings)
    check_exercise_readmes(findings)
    check_forbidden_everywhere(findings)
    for appendix in sorted((BOOK / "appendices").glob("*.md")):
        text = appendix.read_text(encoding="utf-8")
        lines = text.splitlines()
        if sum(1 for l in lines if l.startswith("```")) // 2 != sum(1 for l in lines if l.strip() == '<div dir="ltr" align="left">'):
            findings.append(f"appendices/{appendix.name}: fence/wrapper mismatch")
        if FORBIDDEN.search(text):
            findings.append(f"appendices/{appendix.name}: forbidden term")
    print(f"{len(chapters())} chapters, {len(list((BOOK / 'appendices').glob('*.md')))} appendices, {len(list(EXERCISES.glob('ch*/')))} exercise folders checked")
    for f in findings:
        print("-", f)
    print("OK" if not findings else f"{len(findings)} finding(s)")
    return 0 if not findings else 1


if __name__ == "__main__":
    raise SystemExit(main())
