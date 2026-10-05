# Formatting and writing conventions for the Hebrew book (book 2)

Copied from book 1 (`ai-engineering-course/book/CONVENTIONS.md`) on 2026-10-05 and adapted: the platform is bash
inside WSL2, the code is mostly Go, and the terminology policy lives in `TERMINOLOGY.md`.

Every chapter file follows these rules so that the book reads correctly right-to-left in three
places: the generated HTML (`site/`), Obsidian (with the RTL Support plugin), and GitHub's Markdown
preview.

## File skeleton

```markdown
---
direction: rtl
lang: he
title: "פרק 2: סביבת העבודה: WSL2, Docker, Go ו־uv"
chapter: 2
part: "חלק 0: לפני שמתחילים"
---

<div dir="rtl" lang="he">

# פרק 2: סביבת העבודה: WSL2, Docker, Go ו־uv

...chapter body in Markdown...

</div>
```

- `direction: rtl` in the front matter is read by Obsidian's "RTL Support" plugin.
- The `<div dir="rtl" lang="he">` wrapper, with a blank line after it and before the closing tag,
  makes GitHub and VS Code previews right-to-left. `tools/build_html.py` strips it and sets the
  direction on the `<html>` element instead.
- **Every fenced code block is wrapped** in `<div dir="ltr" align="left">` with a blank line after
  the opening tag and a blank line before the closing `</div>`. Markdown viewers that honor the
  chapter's `dir="rtl"` would otherwise right-align the code and wreck Go and Python indentation (reader
  feedback, 2026-10-03). The site CSS and the Obsidian snippet add the same guarantee, but the
  wrapper is what works everywhere. `tools/wrap_ltr_blocks.py` adds missing wrappers idempotently;
  run it after writing a chapter.

  ```markdown
  <div dir="ltr" align="left">

  ```go
  fmt.Println("left-to-right, indentation intact")
  ```

  </div>
  ```

## Structure of a chapter

1. Title: `# פרק N: <שם>`.
2. Opening paragraph (2–4 lines): what the chapter does and why it matters for the job.
3. `## מה נלמד בפרק` — 4–7 bullets. `## ידע מקדים` — which chapters it relies on.
4. Numbered sections `## N.1 <שם>`, subsections `### N.1.1 <שם>`. Every section is numbered so the
   schedule and other chapters can reference it.
5. Closing apparatus, always in this order and with these exact headings:
   `## סיכום`, `## מונחים לזכור` (table: מונח בעברית | באנגלית | במשפט אחד),
   `## טעויות נפוצות`, `## שאלות ראיון`, `## פתק Obsidian`, `## תרגיל`, `## העשרה אופציונלית`.
6. The Obsidian note is given as a blockquote in the chapter *and* as a file in `obsidian/`.
7. The exercise section states goal, time, prerequisites and completion criteria, and links to
   `exercises/chNN-<slug>/README.md`, which holds the full steps, starter, hints, solution and tests.
   Every coding-exercise README has, in this order: goal + time + prerequisites, `## מה בתיקייה הזאת`,
   the steps (`## צעדים` or one `## יום N: …` per study day), `## פלט צפוי`, `## רשימת בדיקה`,
   `## טעויות נפוצות ותיקונים` (3–6 bullets specific to the exercise: **הטעות.** what you see, what to do),
   `## רמזים`, `## הפתרון` (a pointer plus 2–4 sentences on why it works), `## מה התרגיל מדגים`.
   `tools/consistency_check.py` enforces the section set.
8. Interview questions: numbered, bold question, a model answer that explains the reasoning and the tradeoff
   (never a single "correct" answer); at most two questions per chapter carry a follow-up line
   `**המשך:** <question> <1–2 sentence answer>`. The chapter 36 bank (§36.8) is rendered from
   `exercises/ch36-question-bank/data/questions.jsonl` (fields: id, category, chapter, question, answer,
   tradeoff, follow_up) by `tools/render_question_bank.py`; edit the data, not the rendered region.

## Writing rules

- Short paragraphs (at most five lines). Parallel items go in bullets; comparisons go in tables.
- Second person feminine singular throughout (את, תריצי, שימי לב), matching the reader.
- Every technical term: Hebrew first, English in parentheses on first use in the chapter, for
  example "אידמפוטנטיות (idempotency)". Afterwards use the Hebrew term, or the English identifier in
  backticks when it names something in code. Product nouns that stay English (Kubernetes, pod, topic,
  goroutine, …) get a Hebrew gloss on first use instead; the full policy and the fixed renderings are in
  `TERMINOLOGY.md`, and every new term goes into the chapter's terms table.
- A sentence, bullet or table cell never starts with a Latin word or a digit when it can be avoided;
  begin with a Hebrew word ("המונח `uv` ...", "הפקודה `uv run` ...").
- Hebrew prefix letters attach to a Latin word with a maqaf: ה־API, ב־bash, ל־Kubernetes.
- No em-dashes in Hebrew text; use a colon, a comma or parentheses. Quotation marks: „...” or "...".
- Numbers in Western digits. Units and currency after the number: 3 דקות, 0.30 דולר.
- Code, commands, file names, flags, JSON and model names appear only in backticks or fenced blocks.
  Comments inside code are short and in English. Explanations come after the block, in Hebrew.
- Commands are **bash inside WSL2 (Ubuntu)**, written without a prompt prefix so they can be copied as they are.
  Expected output is shown in its own fenced block labelled `text`. Commands that must run in Windows itself
  (installing WSL, Docker Desktop settings) are PowerShell, appear only in chapter 2, and say so in the sentence
  before the block.
- Callouts are blockquotes with a bold Hebrew label on the first line:
  `> **שימי לב:**`, `> **טעות נפוצה:**`, `> **מונח:**`, `> **ב־Mac:**`, `> **איך Rust עושה את זה:**`,
  `> **העשרה אופציונלית:**`. WSL2 is the primary platform; Mac differences are confined to chapter 2 §2.13
  (equivalents table) and the Mac rows of appendix B. Do not add per-command Mac variants elsewhere; where a
  command really differs on a Mac (for example `sed -i`), the chapter says it once in a `> **ב־Mac:**` callout.
- Rust comparison boxes: `> **איך Rust עושה את זה:**` in Part A chapters (Go taught against Rust), and
  `> **איך Go הייתה עושה את זה:**` where a later chapter shows the Go idiom she would reach for in Rust terms.
  At most three such boxes per chapter, each 2–5 lines, never teaching Rust from scratch.
- Code in chapters is copied from the exercise or capstone files, never typed from memory; a block that shows a
  fragment names the file it comes from in the sentence before it.
- Cross-references: "ראי פרק 5, סעיף 5.3". Links to files in the repository use relative paths.
- Versions, image tags, prices and limits never appear hard-coded in prose as facts of the world; they
  are quoted "נכון לאוקטובר 2026" and live in `go.mod`, `pyproject.toml`, compose files, workflows or appendix C.
- Honesty rule inherited into every exercise: a result the reader did not measure is never
  described as measured.

## Mixed-direction readability (added 2026-10-03 after reader feedback)

The reader found a note unreadable: several labelled items collapsed into one paragraph, dense with
English identifiers inside Hebrew text. Rules that prevent it:

- **One item per line means one list item.** Consecutive lines in a paragraph or blockquote merge
  when rendered. Anything meant to be read line by line is a bulleted list.
- **At most two English or code tokens per Hebrew sentence.** If a sentence needs more, change the
  shape: a table (Hebrew column + English column), a bulleted list with one token per bullet, or a
  Hebrew sentence followed by a left-to-right code block.
- **Commands never live inside Hebrew prose.** A single short token in backticks is fine; anything
  with spaces, flags or paths goes in a fenced block under the sentence that explains it.
- **Cheat sheets are English-only, in a `text` fence.** The Obsidian note of every chapter has two
  parts: Hebrew bullets for concepts, then an English-only `text` block for commands, identifiers and
  blanks to fill in. Never interleave the two.
- **Explain command sequences with a two-column table** (command | what it did), not a paragraph.
- **No arrows, pluses or slashes between words in Hebrew text** (←, →, +, /). Write the words.
- **Completion criteria, summaries and checklists are bullets**, Hebrew first, with any identifier at
  the end of the sentence.
- **Avoid long inline paths in sentences.** Say what the file is in Hebrew, then give the path at
  the end of the sentence or in a code block.

## Mermaid diagrams

Use fenced blocks labelled `mermaid`. The HTML build renders them in the browser; Obsidian renders
them natively. Keep node labels short and in English or transliterated Hebrew (Mermaid handles
Hebrew labels poorly).
