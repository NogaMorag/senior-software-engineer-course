# Quality checklist (the handoff brief as a checklist)

Written 2026-10-05. Ticked at Stage 3 (2026-10-05); evidence and the recorded exceptions are in `09-review.md` §1, §2 and §5.

## Deliverables
- [x] README.md (status, start here, layout, tools, how to verify, limitations)
- [x] book/CONVENTIONS.md adapted (bash/WSL2 primary, Mac equivalents), book/TERMINOLOGY.md
- [x] 40 chapters (00–39) in Part 0 + ten parts; book/README.md master TOC with status
- [x] Appendices A (generated), B, C, D, E, F (Hebrew, days 0–100, six fields each)
- [x] One exercise package per chapter with code (README, starter, solution, tests, data)
- [x] project-hookline: Hebrew README, hookline/ reference, hookline-starter/ (generated)
- [x] obsidian/NN-*.md generated + snippets/book-rtl.css; tools/ adapted; site/ generated
- [x] 00-stage1/04–09 written

## Writing contract
- [x] Hebrew first; English in parentheses on first use per chapter; identifiers only in backticks
- [x] ≤2 Latin/code tokens per Hebrew sentence; commands never inline; every fence LTR-wrapped
- [x] No arrows, equals, pluses or slashes between Hebrew words; feminine singular; paragraphs ≤5 lines; numbered sections
- [x] Closing apparatus in order; Obsidian note = Hebrew bullets + exactly one English `text` cheat sheet
- [x] Interview questions explain reasoning and tradeoffs; ≤2 **המשך:** per chapter; ch 36 bank from JSONL
- [x] Exercise READMEs: all 13 items; 3–6 exercise-specific common mistakes; prerequisites name imported modules
- [x] Honesty: nothing described as measured unless measured; placeholders not numbers; dated facts in config/appendix C
- [x] No employer material; no secrets

## Technical bar
- [x] Every code sample in a chapter copied from exercise/capstone code
- [x] Go: gofmt, go vet, staticcheck clean; go test ./... green per exercise and capstone; exact pins
- [x] Python: ruff clean, pytest green, uv only
- [x] Compose stacks, kind/kubectl/helm, terraform plan/apply (kubernetes provider), k6 executed here
- [x] Workflows validated with actionlint; stated to run in the reader's repo later
- [x] External facts verified on primary sources with dates in 07-tech-stack.md
- [x] Anything not run listed in 09-review.md (AWS section; WSL2 itself; GitHub-hosted runs)

## Schedule
- [x] Day 0 + 100 days; 75–120 min/day; install days separate; one lighter day a week; six checkpoints
- [x] Each exercise names its days; schedule section ranges match chapters (checker)
- [x] Milestones 2–3 days each; last week = mocks, polish, publish, checklists, retrospective

## Process
- [x] One commit per part after wrap/glossary/html/consistency/tests/linters
- [x] Stage 3 self-review in the five-part format; zip refreshed at ~/senior-software-engineer-course.zip
