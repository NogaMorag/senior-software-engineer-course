# Senior Software Engineer in Industry — textbook + study program (Hebrew), book 2

**Status (2026-10-05): complete and self-reviewed.** 40 Hebrew chapters (0–39) in Part 0 and ten parts, six
appendices, 39 Obsidian notes, 36 chapter exercise packages, a 100-day program plus Day 0, and the Hookline capstone
(reference implementation + generated starter). The master table of contents with per-chapter status is
`book/README.md`. Review record: `00-stage1/09-review.md`.

This repository is personal study material, separate from any work project: nothing here contains employer code,
names or data, and no secrets (placeholders such as `change-me` only).

## Start here
1. Read `book/00-how-to-study.md` (chapter 0), then do **Day 0**: chapter 2, sections 2.1–2.8 (installations only; on a
   Mac, section 2.13 replaces 2.2–2.4).
2. Follow the daily plan in `book/appendices/f-daily-schedule.md` (Hebrew; the English twin is
   `00-stage1/05-daily-schedule.md`).
3. Every chapter ends with the same apparatus: summary, terms, common mistakes, interview questions, an Obsidian note,
   the exercise, optional enrichment. Exercises live in `exercises/chNN-<slug>/` (`README.md`, `starter/`, `solution/`,
   `tests/`, `data/`).

**Platform.** Commands are bash inside WSL2 (Ubuntu) on Windows 11, with Docker Desktop's WSL2 backend. Mac
differences are in chapter 2 §2.13 and the Mac rows of appendix B. **Cost.** Everything runs locally and free; the only
cloud section (chapter 22 §22.7, AWS free tier) is optional.

## Reading the book
Right-to-left Markdown with every code block forced left-to-right (`book/CONVENTIONS.md`, `book/TERMINOLOGY.md`).
The comfortable way is the generated HTML:

```bash
uv run --with markdown tools/build_html.py     # writes site/*.html + site/index.html
```

Or open `book/` in Obsidian with the "RTL Support" plugin and the CSS snippet in `obsidian/snippets/`.

## Layout
```text
README.md                      this file
book/README.md                 master table of contents + status
book/CONVENTIONS.md            formatting and writing rules (RTL, LTR code, apparatus, bash/WSL2, Mac equivalents)
book/TERMINOLOGY.md            which terms stay English, one Hebrew rendering per concept
book/NN-<slug>.md              chapters 00..39
book/appendices/               a glossary (generated), b troubleshooting, c tool versions, d note templates, e checklists, f schedule
exercises/chNN-<slug>/         README (13-item template), starter/, solution/, tests/, data/
exercises/project-hookline/
  README.md                    the capstone brief in Hebrew
  hookline/                    reference implementation (Go module + Python sink, migrations, chart, Terraform, CI, docs)
  hookline-starter/            generated: Go bodies panic("TODO"), Python bodies NotImplementedError, tests kept
obsidian/                      one note per chapter (generated from each chapter's note section) + CSS snippet
tools/                         generators and checks (below)
00-stage1/                     plan, decisions, research, TOC, schedule (EN), capstone spec, tech stack, checklists, review
site/                          generated HTML
```

## Tools (single sources of truth)
| Command (from this folder) | What it does |
| --- | --- |
| `uv run --no-project python tools/notes_from_chapters.py 1 2 … 39` | regenerates `obsidian/NN-*.md` from the chapters |
| `uv run --no-project python tools/make_glossary.py` | regenerates appendix A from every chapter's terms table |
| `uv run --no-project python tools/render_question_bank.py` | renders §36.8 from `exercises/ch36-question-bank/data/questions.jsonl` |
| `uv run --no-project python tools/make_project_starter.py` | regenerates `hookline-starter/` from `hookline/` (uses `tools/gostub`) |
| `uv run --no-project python tools/wrap_ltr_blocks.py` | adds missing LTR wrappers around code fences (idempotent) |
| `uv run --with markdown tools/build_html.py` | regenerates `site/` |
| `uv run --no-project python tools/consistency_check.py` | read-only audit: TOC, apparatus, wrappers, links, notes, exercise templates, schedule days 0..100 and section references |
| `uv run --no-project python tools/verify_exercises.py [chNN …]` | runs every exercise's checks: test-copy identity, gofmt/vet/staticcheck/`go test -race`, starter build, uv sync/ruff/mypy/pytest |
| `uv run --no-project python tools/mark_done.py N …` | flips chapter status in `book/README.md` |

## Verifying the code
- All exercises at unit level: `uv run --no-project python tools/verify_exercises.py`.
- Integration tests skip unless their services are up. Each exercise README brings its stack up with
  `docker compose -f data/compose.yaml -p chNN up -d --wait`, sets `DATABASE_URL` / `REDIS_ADDR` / `KAFKA_BROKERS`, runs
  `go test -race ./...`, and tears down with `down -v`.
- Hookline: `cd exercises/project-hookline/hookline && make up && make test && make down`; Kubernetes, Terraform,
  observability, load and chaos targets are listed in its README.

## Limitations (details in `00-stage1/09-review.md`)
- The book was executed on macOS (Apple silicon), not on WSL2: the WSL2/Windows installation steps in chapter 2 come
  from Microsoft's and Docker's documentation and were not run; Linux commands were replayed in an Ubuntu container.
- GitHub Actions workflows were validated with actionlint (and their job steps reproduced locally); their first real
  run happens in the reader's own repository.
- The optional AWS section (chapter 22 §22.7) was validated, not applied.
- Measured numbers appear only where an exercise measures them, labelled as the run on the book's machine; every
  result table for the reader is a placeholder.
