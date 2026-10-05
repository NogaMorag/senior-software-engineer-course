# Stage 3 reviewer brief (internal)

Written 2026-10-05 by the coordinator. Each reviewer owns a fixed set of chapters, their exercise READMEs and their
Obsidian notes, reads every line, **fixes language and format problems directly**, and **reports** (does not fix)
technical problems in code. The rules are the handoff brief's, condensed in `00-stage1/AUTHORING-BRIEF.md`,
`book/CONVENTIONS.md` and `book/TERMINOLOGY.md`; book 1's Stage 4 lessons
(`~/Downloads/ai-engineering-course/00-stage1/08-stage3-consistency-review.md`, section "Stage 4") apply too.

## Checklist (per chapter)
1. **Language.** Natural Hebrew, no calques; second person feminine; ≤2 Latin/code tokens per Hebrew sentence (reshape
   into a table, bullets, or sentence + code block); no arrows, `=`, `+`, `/` between Hebrew words; no em-dashes in
   Hebrew; maqaf after Hebrew prefixes on Latin words; paragraphs ≤5 lines; commands never inline in prose; one rendering
   per concept per `TERMINOLOGY.md` (tenant = דיירת with agreeing grammar; appendices "נספח א…ו"); the author is never
   "הכותב/הכותבת" (use "בהרצה שבה נכתב הספר"); no gendered slips toward the reader.
2. **Format.** Front matter; rtl div; exact `##` sections from `04-table-of-contents.md`; apparatus order; interview
   questions explain reasoning and tradeoffs, ≤2 **המשך:**; Obsidian section = one intro line, Hebrew bullets, exactly one
   English `text` block; every fence LTR-wrapped; `**זמן:**` line names the days.
3. **Honesty.** No number presented as measured unless the exercise measures it or it is labelled as the run on the
   book's machine; no invented statistics; no unsupported superlatives; market claims cite `01-market-research.md` and
   don't exceed it; dated facts say "נכון לאוקטובר 2026" and point to appendix C.
4. **Prerequisites and sequence.** "ידע מקדים" names every earlier chapter, module or tool the chapter or exercise
   relies on; nothing used before it is taught (check against earlier chapters' TOC sections).
5. **Code fidelity.** Every fenced code block in the chapter (go, python, sql, yaml, bash scripts quoted as files)
   exists in the exercise or capstone files. Run a script: for each fence, normalize trailing whitespace and search
   the chapter's exercise folder; report misses (fix by copying the real code, never by editing exercise code).
6. **Exercise README.** All 13 items; exact commands work in the stated order (read them against the files); "common
   mistakes" 3–6 specific bullets; prerequisites name imported modules.
7. **Schedule.** Appendix F and `05-daily-schedule.md` day blocks for the chapter's days match what the chapter and
   README actually ask for on that day (sections, exercise steps, time realism 75–120 minutes). Fix the schedule text if
   it is wrong about the chapter (both languages), not the chapter.
8. **Employer and secrets.** No employer names, code, incidents; no secrets (placeholders only).

## Rules
- Edit only your chapters, their exercise READMEs, their Obsidian notes (regenerate with
  `uv run --no-project python tools/notes_from_chapters.py N` after changing a note section), and the matching day
  blocks of the two schedules. Never edit exercise code, `tools/`, `book/README.md`, or other reviewers' chapters.
- After edits: `uv run --no-project python tools/wrap_ltr_blocks.py <files>` and
  `uv run --no-project python tools/consistency_check.py` (no findings on your files).
- No Docker, no test runs needed (the coordinator re-runs every suite); no commits.

## Report (English, concise)
Per chapter: what you changed (counts by kind: language, format, honesty, prerequisites, schedule), code-fidelity
misses and how fixed, technical concerns you found but did not fix (file:line, why), facts you doubt (with what a
primary source would need to confirm), and anything left open.
