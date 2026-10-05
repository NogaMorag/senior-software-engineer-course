# Stage 3 review — self-review against the handoff brief (2026-10-05)

**Method.** The brief was turned into a checklist (`08-quality-checklist.md`). Every chapter, exercise README,
appendix, Obsidian note and both schedules were read line by line by six delegated reviewers working one territory
each, with fixed instructions (`REVIEW-BRIEF.md`): language and format, honesty, hidden prerequisites, code fidelity
(scripted: every fence in a chapter searched in its exercise or project files), the 13-item exercise template,
schedule realism and section ranges. They fixed language and format directly and reported code problems, which the
coordinator handled. A seventh agent re-checked about 35 doubted facts on primary sources. All suites and checks were
then re-run by the coordinator. Provenance: the Stage 1 files did not transfer from the authoring Linux machine and
were reconstructed (`03-decisions-2026-10-04.md`); this book was built and executed on the reader's Mac.

## 1. Compliance

| Requirement (brief) | Status | Evidence |
| --- | --- | --- |
| 40 chapters, Part 0 + ten parts, master TOC with status | ✅ | `book/00`–`39`; `book/README.md` all ✅ |
| Appendices A (generated), B, C, D, E, F (days 0–100, six fields) | ✅ | 473-term glossary; F has 101 blocks, checker verifies fields and section references |
| One exercise package per chapter with code | ✅ | 36 folders (ch01–ch36); ch31 and ch34 are non-coding by design |
| Capstone: Hebrew brief, reference, generated starter | ✅ | `exercises/project-hookline/` M1–M10; starter from `tools/make_project_starter.py` (+ `tools/gostub`) |
| CONVENTIONS adapted (bash/WSL2, Mac equivalents), TERMINOLOGY | ✅ | `book/CONVENTIONS.md`, `book/TERMINOLOGY.md` |
| Writing contract (Hebrew first, ≤2 Latin tokens, LTR fences, apparatus order, feminine, ≤5-line paragraphs) | ✅ with recorded exceptions | Reviewed by six reviewers; exceptions in §2 |
| Obsidian note = Hebrew bullets + exactly one English `text` block | ✅ | Checker enforces one `text` block per note |
| Interview questions with reasoning and tradeoffs, ≤2 **המשך** | ✅ | Apparatus sections; §39.7's project drill carries 4 follow-ups by design |
| Question bank JSONL rendered by the tool | ✅ | 135 questions, `quiz check` valid, §36.8 rendered |
| 13-item exercise README, 3–6 specific mistakes | ✅ | Checker enforces sections and bullet count |
| Honesty (no unmeasured numbers, placeholders, dated facts in config/appendix C) | ✅ | Author-run numbers labelled "בהרצה שבה נכתב הספר"; reader tables are placeholders |
| No employer material, no secrets | ✅ | Checker scans book, exercises, notes, stage-1 files |
| Code in chapters copied from exercise/capstone code | ✅ | Scripted fidelity checks per part; misses fixed by copying real code |
| Go: gofmt, vet, staticcheck, `go test` green; exact pins | ✅ | §5 |
| Python: ruff, pytest, uv only | ✅ | §5 (mypy strict too) |
| Compose, kind/kubectl/helm, Terraform plan/apply (kubernetes provider), k6 executed | ✅ | Executed by the writers under the machine-wide lock; see §5 |
| Workflows validated with actionlint, run later in her repo | ✅ | actionlint clean; job steps reproduced locally (ch23); stated in ch23 and §38.7 |
| External facts verified on primary sources, dated | ✅ | `versions-verified-2026-10-05.md`, `07-tech-stack.md`, fact-check pass |
| Schedule: 75–120 min, install days separate, lighter day weekly, six checkpoints, milestones 2–3 days, last week | ✅ | `05-daily-schedule.md` + appendix F; checkpoints 14, 45, 56, 71, 87, 100 |
| One commit per part; zip; never push | ✅ | Local commits per part; zip at `~/senior-software-engineer-course.zip`; nothing pushed |

## 2. Gaps (what is not done, or done differently, and why)

1. **WSL2 was not executed.** The author machine is macOS. Chapter 2's Windows/WSL2/Docker Desktop steps come from
   Microsoft's and Docker's documentation; the Linux commands were replayed in an `ubuntu:26.04` container. The WSL disk
   compaction procedure (§2.12) is sourced but unexecuted.
2. **GitHub Actions never ran on GitHub.** Validated with actionlint; chapter 23's steps were reproduced locally
   (Postgres container, kind smoke). First real runs happen in her repository.
3. **AWS (chapter 22 §22.7)** was validated (`fmt`, `validate -backend=false`), never applied.
4. **Hookline on kind:** the smoke test, chart and Terraform release ran on kind; the chaos experiment ran on the
   compose stack only (kind has no metrics-server in this setup; `kubectl top` not run). Stated in §39.4.
5. **Sink image supply-chain gate fails today** on one fixable High (`libpcre2-8-0`) in `python:3.14.8-slim`, fixed
   upstream in a Debian point release not yet in the tag. The gate was not lowered; §38.9.6 and §39.6 document the
   rebuild-with-`--pull` policy and the digest pin.
6. **Stage 1 reconstruction.** The approved proposal and decisions did not transfer; they were rebuilt from the
   handoff and a fresh market-research pass (69 sources). The chapter list is a reconstruction and the reader may
   amend it.
7. **Length.** Most chapters are 35–60 KB (chapter 2: ~86 KB; chapter 36: ~190 KB including the rendered bank),
   above book 1's typical size; Hebrew is two bytes per character and about a third is code. Not trimmed further to
   avoid cutting install steps and worked examples.
8. **Recorded exceptions to the writing contract.** Appendix F's time split "קריאה 25 / בנייה 55 / פתק 10" keeps book 1's
   label format (slashes between labels). A few multi-word product names exceed two Latin tokens in a sentence (§36.2,
   three bank answers, §38.8). §39.7 carries four follow-ups (the project drill), the apparatus sections ≤2.
9. **Small open items.** The chapter 15 lab client ignores `Retry-After` (the chapter no longer claims it honours it).
   The M5 smoke step `docker compose exec postgres psql … UPDATE` in §38.5.4 was run in its `docker exec` equivalent,
   not verbatim. A few expected outputs (§38.7 kind and Terraform outputs, `make run-dev`'s key print) come from the
   writer's runs and were not re-run in the final pass.

## 3. Plan for the gaps (what the reader does, when)

- Day 0 is the first WSL2 run: follow chapter 2 and record any deviation in the stuck log; appendix B has the
  expected failure endings. If a step differs, it is a book bug.
- Day 70 (M7) and chapter 23 (days 66–67): push to her own repository and watch the first workflow runs.
- Before publishing Hookline (day 99): `docker pull` the sink base, update the digest per §39.6, rerun
  `make supply-chain`.
- Optional: chapter 22 §22.7 with a budget alarm first; chaos on kind after installing metrics-server (§21.2).

## 4. Corrections made during the review (selected; full detail in the commit history)

- **Technical.** Two README sequencing bugs (ch01, ch06); a false common-mistake bullet (ch05 `SetLimit`); chapter 16
  claiming the chapter 15 client honours `Retry-After`; PgBouncer `max_prepared_statements` default (200 since 1.24);
  testcontainers' Docker lookup order; Trivy and pip (scanners read pip's own metadata, not vendored libraries; the
  Dockerfile comment corrected in every copy); Copilot reads `AGENTS.md` only in some features; `GOMAXPROCS` respects
  container CPU limits since Go 1.25; a Rust box on monomorphization; a Rust box claiming compile-time lock ordering; a
  wrong M3 test filter that matched nothing; Hookline M5 replay made consistent with the §29.5 design doc
  (`attempt_base`); a tenant-isolation leak found by a database run (409 instead of 404, fixed with a composite key);
  a secret leaking through a logged map (found by the log-scan test); dispatcher fan-out stalls under SIGKILL fixed with
  a 10 s consumer session (found by the chaos run).
- **Tooling.** The starter generator guessed package names from import paths (`go-redis/v9` is package `redis`); it now
  asks `go list`. The checker gained schedule section-reference checks, apparatus order, one-`text`-block notes and the
  3–6 mistakes rule; `verify_exercises.py` runs every module.
- **Program.** Day 0 Mac wording; checkpoint 5 covers Parts F and G; all 22 Hookline milestone day blocks rewritten to
  match the briefs; day splits aligned to chapters across all parts.
- **Language.** One rendering per concept (tenant, advisory lock, fan-out, interface, method, event loop, task);
  author references neutral; appendix references as Hebrew letters; calques, gender slips, token-dense sentences and
  inline commands fixed throughout.
- **Honesty.** Unsupported superlatives and invented throughputs removed; author-run numbers labelled; market claims
  aligned to `01-market-research.md` (e.g. Kubernetes in 19/30 postings, not "almost every").

## 5. Verification (final pass, 2026-10-05)

| Check | Result |
| --- | --- |
| `tools/consistency_check.py` | OK — 40 chapters, 6 appendices, 36 exercise folders, 0 findings |
| `tools/verify_exercises.py` (all 36 folders, unit level) | 36 ok; Go: 25 modules, 857 passed, 126 skipped (integration tests without services), 0 failed; Python: 11 projects, 398 passed |
| Integration tests with services up | run by each writer under the exclusive lock (e.g. ch10–14, 16–18, 26, 30, 35 against PostgreSQL/Redis/Redpanda containers, 0 skipped) |
| Hookline reference | gofmt/vet/staticcheck clean; 33 packages `go test -race` green at unit level; full suite against compose 191 top-level tests + 152 subtests, 0 skipped (M10 run); sink 88 pytest, ruff, mypy clean; actionlint clean |
| Hookline starter | `go build ./...` and `go vet ./...` pass; tests fail with TODO panics as intended |
| Hookline runs | kind + Helm smoke (event reached the in-cluster sink); Terraform apply/plan/destroy on kind; chaos 11 SIGKILLs: 0 missing deliveries; promtool rules 6 scenarios; traces across 7 spans in Tempo; k6 N1, N5, N6 thresholds |
| Chapter 33 Rust crate | `cargo fmt --check`, `clippy -D warnings`, `cargo test` 4 passed (Rust 1.99.0) |
| Kubernetes and Terraform chapters | ch20 9+9 checks; ch21 25/25 static + live walk (QoS, OOMKilled, HPA, rollout undo, ingress, Helm); ch22 10-step smoke |
| Observability chapters | ch24 stack checks all passed; ch25 promtool tests and a staged incident that fired and resolved |
| Question bank | `quiz check`: 135 questions, valid; 25 tool tests |
| HTML | `tools/build_html.py`: 47 pages + index |

The reader pulls with `git pull --ff-only` from this machine's repository or unzips
`~/senior-software-engineer-course.zip`.
