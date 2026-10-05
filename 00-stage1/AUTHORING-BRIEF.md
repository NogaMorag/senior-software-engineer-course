# Authoring brief for chapter writers (internal, Stage 2)

Written 2026-10-05 by the coordinator. Every agent that writes chapters or exercises for this book follows this file.
It condenses the handoff brief; where they differ, the handoff brief wins (the coordinator holds it).

## 0. Read before writing (in this order)
1. `book/CONVENTIONS.md` and `book/TERMINOLOGY.md` — binding format, language and term renderings.
2. `00-stage1/04-table-of-contents.md` — your chapters' exact `##` sections, exercise slug, days, capstone reuse.
   **Use exactly those section numbers and titles** (you may add `###` subsections).
3. `00-stage1/02-proposal.md`, `00-stage1/03-decisions-2026-10-04.md`, `00-stage1/06-final-project.md` (Hookline),
   `00-stage1/07-tech-stack.md` (pins: use these versions, nothing else), `00-stage1/01-market-research.md` (cite it
   when you make market claims; if a claim is not there, do not make it).
4. The model for form and tone (book 1, read-only, never copy content): 
   `~/Downloads/ai-engineering-course/book/11-robust-clients.md` (a full chapter), `.../book/02-workbench-windows.md`
   (the install-chapter pattern), `.../exercises/ch11-production-client/README.md` (exercise README template).

## 1. The reader
Noga: experienced developer, Rust daily, Python, C++ long ago; beginner in Go, Kubernetes, Kafka, Terraform,
observability, 2026 senior interviews; not confident with unfamiliar installations. Intelligent, takes no concept for
granted. Address her in second person feminine singular. Never condescend; never skip a step she needs.

## 2. Chapter file (`book/NN-slug.md`)
- Front matter exactly like book 1 (`direction: rtl`, `lang: he`, `title: "פרק N: ..."`, `chapter: N`,
  `part: "חלק X: ..."`), then `<div dir="rtl" lang="he">`, a blank line, the body, a blank line, `</div>`.
- `# פרק N: <title>`; opening paragraph (2–4 lines: what and why it matters for the job); `## מה נלמד בפרק` (4–7
  bullets); `## ידע מקדים` (name every earlier chapter, module or tool it relies on — no hidden prerequisites); the TOC's
  numbered sections; then the closing apparatus **in this order with these exact headings**: `## סיכום`,
  `## מונחים לזכור` (table `| מונח בעברית | באנגלית | במשפט אחד |`), `## טעויות נפוצות`, `## שאלות ראיון`,
  `## פתק Obsidian`, `## תרגיל`, `## העשרה אופציונלית`. (Chapter 0 has no apparatus.)
- Length: comparable to book 1 chapters (about 20–45 KB of Markdown, a third of which may be code copied from the exercise). Depth over breadth: explain *why*, show the
  real code, show what goes wrong.
- Interview questions: 4–6, numbered, bold question, an answer that explains reasoning and tradeoffs (never "the
  correct answer"); at most two carry `**המשך:** <follow-up question> <1–2 sentence answer>`.
- `## פתק Obsidian`: first paragraph exactly one line: "הפתק המוכן נמצא גם כקובץ: `obsidian/NN-slug.md`. זה תוכנו:"
  then a blank line, then Hebrew bullets (one idea per bullet, ending with `- **שאלה פתוחה:** ___`), then exactly
  one English-only ```` ```text ```` cheat sheet (commands, identifiers, blanks to fill). Nothing after it in the section.
- `## תרגיל`: `**מטרה:** … **זמן:** … **ידע מקדים:** …` — **the days must be on the same source line as `**זמן:**`**
  (e.g. "**זמן:** שני ימי לימוד, ימים 24 ו־25, כ־60 דקות בנייה בכל אחד."; forms: "יום N", "ימים N עד M", "ימים N ו־M").
  Then the link line to `exercises/chNN-slug/README.md` (relative: `../exercises/...`), then `**קריטריון סיום:**` bullets.
- Rust boxes (Part A especially): `> **איך Rust עושה את זה:**` 2–5 lines, at most three per chapter.
- Code blocks in the chapter are **copied from your exercise files** (name the file in the sentence before a
  fragment). Never type code in the chapter that does not exist in the exercise.
- Commands are bash inside WSL2, in fenced `bash` blocks, never inline in Hebrew prose; expected output in a `text`
  fence. Mac differences only in chapter 2 §2.13 (and a `> **ב־Mac:**` callout where a command truly differs).
- Version-dependent facts: "נכון לאוקטובר 2026" and a pointer to appendix C (`book/appendices/c-tool-versions.md`);
  pins live in `go.mod`, `pyproject.toml`, compose files.
- Language rules that reviewers will check: ≤2 Latin/code tokens per Hebrew sentence (otherwise a table, bullets, or a
  sentence followed by a code block); no arrows, `=`, `+`, `/` between Hebrew words; no em-dashes in Hebrew; maqaf
  after Hebrew prefixes on Latin words (ב־Go); a sentence/bullet/table cell starts with a Hebrew word when possible;
  paragraphs ≤5 lines; natural Hebrew, not calques of English. Use the TERMINOLOGY renderings, one per concept.
- The author is never "הכותב" or "הכותבת": write "בהרצה שבה נכתב הספר" / "במחשב שעליו נכתב הספר". An author-run result
  is shown only when labelled that way, and the reader is asked to record her own.
- Honesty: never present a number as measured unless the exercise measures it; write "תמדדי ותרשמי" and leave a
  placeholder. No invented statistics, no unsupported superlatives.
- No employer material, no company names from the reader's work, no secrets (tokens/keys only as obvious placeholders
  like `change-me`).

## 3. Exercise package (`exercises/chNN-slug/`)
Course conventions the READMEs rely on (set up in chapter 2): the book repository is cloned at
`~/senior-software-engineer-course` and exported as `COURSE`; her own work lives in `~/senior-course/chNN/` (one
directory per chapter, a Git repository at `~/senior-course`). Commands copy from `$COURSE/exercises/...`.

Layout:
```text
README.md        Hebrew, front matter + rtl div like a chapter
starter/         what she starts from: compiles (Go) / imports (Python); bodies with TODO
solution/        complete, green
tests/           the canonical tests; byte-identical copies live in starter/ and solution/ at the same relative path
data/            fixtures, seeds, compose files (when needed)
```
- **Go**: `solution/` and `starter/` are each a Go module (`module example.com/chNN-slug`, `go 1.27`), dependencies
  pinned to `07-tech-stack.md` versions (`go get mod@version`), `go.sum` committed. Tests are table-driven, use the
  standard library (`testing`, `httptest`) unless the chapter teaches a library. `starter/` must `go build ./...` and
  `go vet ./...`; its tests fail with clear TODO panics or wrong results until done.
- **Python**: `solution/` and `starter/` are uv projects (`pyproject.toml` with exact pins, `uv.lock`, `src/<pkg>/`,
  `tests/`), Python 3.14, ruff + mypy configured in pyproject. Never pip.
- Services that need PostgreSQL, Redis or Redpanda: a compose file in `data/` (images pinned per `07-tech-stack.md`);
  tests read `DATABASE_URL` / `REDIS_ADDR` / `KAFKA_BROKERS` and `t.Skip("... set DATABASE_URL; see README")` when unset.
  You must still **run them for real** with the stack up and report the count that ran.
- README sections, in this order, all required (the checker enforces them): `**מטרה:**`, `**זמן משוער:**` (with the
  days), `**ידע מקדים:**` (name every earlier module it imports or copies), `## מה בתיקייה הזאת` (table of exact files),
  the steps (`## צעדים`, or one `## יום N: …` per study day) with exact commands, `## פלט צפוי`, `## רשימת בדיקה`,
  `## טעויות נפוצות ותיקונים` (3–6 bullets, each `- **<the mistake>.** what you see, what to do`, specific to this
  exercise), `## רמזים`, `## הפתרון` (pointer + 2–4 sentences on why it works), `## מה התרגיל מדגים`.

## 4. Verification you must run before reporting (from the repo root unless stated)
```bash
export PATH=$HOME/.local/go/bin:$HOME/.local/bin:$PATH
# Go, for each solution/ and starter/ module:
gofmt -l . ; go vet ./... ; staticcheck ./... ; go test ./...        # solution: all clean and green
go build ./... ; go vet ./...                                         # starter: compiles
# Python, for each solution/:
uv sync --frozen && uv run ruff check . && uv run ruff format --check . && uv run mypy src && uv run pytest -q
# Book tooling:
uv run --no-project python tools/wrap_ltr_blocks.py book/NN-*.md exercises/chNN-*/README.md
uv run --no-project python tools/notes_from_chapters.py N [N ...]
uv run --no-project python tools/consistency_check.py      # fix every finding about YOUR files
```
- Copy tests to scratch dirs under `/tmp/senior-course/<your-part>/` if you need to experiment; never leave build
  artefacts (`bin/`, `.venv/`) in the repo (they are git-ignored, but keep the tree clean).
- Anything that uses Docker (compose stacks, testcontainers, kind, k6 against containers) must run under the
  machine-wide exclusive lock and be torn down afterwards, because other writers share this Docker daemon:
  `"$TEST_LOCK" (the machine-wide lock script; its path is given in your prompt) <command>` and `docker compose -p <unique> ... down -v`.
  Use a unique compose project name per exercise (`-p chNN`).
- Do **not** edit `book/README.md`, the appendices, other parts' chapters, `00-stage1/*` or `tools/*` (report needed
  changes instead). Do **not** run `git commit` or `git push`; the coordinator commits per part.

## 5. Report back (in English, concise)
Chapters and exercises written (paths); for every module: gofmt/vet/staticcheck/test results with counts (passed,
skipped and why); Python ruff/mypy/pytest counts; Docker-backed runs actually executed; consistency-check findings
remaining for your files (should be none); any fact you could not verify on a primary source; any deviation from the
TOC (should be none) and why; anything the capstone milestone writer needs from your exercise code.
