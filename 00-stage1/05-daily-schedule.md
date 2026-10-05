# Stage 1 — Daily study schedule (20 weeks, 100 study days + Day 0)

The English source of appendix F. Chapters, numbered sections, exercise slugs and the day of every section follow
`04-table-of-contents.md` exactly; if the two ever disagree, the table of contents wins and this file is fixed in
the same change.

## Rhythm and rules

- **Rhythm**: five study days a week (Sunday–Thursday suits the Israeli week), 75–120 minutes each, debugging
  included. Weekends are free or catch-up. Every day shows its planned split as read / build / note; the note is
  the Obsidian note for the day (templates in appendix D).
- **Day 0** is installations only, done before Week 1 with no time pressure. **Day 57** is the second install
  slot, for the Part E tools. Neither install day mixes in the first real use of a tool: a tool is installed and
  verified one day and used for real on a later day.
- **One lighter day a week**, marked "lighter day" in its title: day 1, the checkpoint days, review day 35 and a few
  concept-heavy days. A lighter day is the natural place to catch up.
- **Six checkpoints** (days 14, 45, 56, 71, 87, 100), at the end of Parts A, C, D, E, of Parts F and G together, and of the capstone. Each
  carries a review list of what must work before you move on.
- **Hookline milestones M1–M10** get two or three days each, right after the part that teaches them. A milestone
  day reads its section of chapter 38 and names the Hookline modules it touches: `api` (HTTP API and ingest),
  `relay` (outbox relay), `dispatcher` (fan-out into deliveries), `worker` (delivery workers), `sink` (the Python
  receiver simulator) and `deploy` (images, chart, Terraform, CI). Chapter 37 is the specification.
- **A day ends when its "Done when" holds**, not when the time runs out. If a build day is not done, finish it
  before the next build day; reading may slide a day.
- **If you fall behind**: skip optional items first (the optional AWS section, optional enrichment), then merge a
  lighter day into its neighbour. **Never skip a checkpoint.**
- **Platform**: every command in the book is bash inside WSL2 (Ubuntu) on Windows 11, with Docker Desktop on the
  WSL2 backend. On a Mac, follow the same days with ch 2, section 2.13. This schedule describes tasks in words; the
  commands are in the chapters.
- **Honesty**: every number you write down (latency, image size, query time, call counts) is one you measured on
  your own machine. The schedule never states a result in advance.
- **Cost**: everything runs locally and free; the only account needed is GitHub. The AWS section on day 65 is
  optional.

---

## Before Week 1 (day 0)

### Day 0 — Installations: WSL2, Docker, Git, VS Code, Go, uv
- **Objective:** a working, verified workbench before any first use, installed calmly with time to troubleshoot.
- **Read:** ch 2, sections 2.1–2.8 (on a Mac, ch 2, section 2.13 replaces 2.2–2.4; the rest is read as is).
- **Time:** 105 min (read 25 / install 70 / note 10). Do it the weekend before Week 1; if a step fails, ch 2,
  section 2.11 and appendix B are the first stops.
- **Exercise:** `ch02-workbench`, install half only: enable WSL2 and install Ubuntu; set up the terminal; install
  Docker Desktop with the WSL2 backend and turn on integration for the distro; install Git and create an SSH key;
  install VS Code with the WSL extension; install Go and the pinned Go tools plus golangci-lint (section 2.7.1,
  checksum-verified), and uv with Python, at the versions in appendix C. Restart the terminal and check each tool's
  version.
- **Done when:** inside a fresh WSL2 terminal each of Docker, Git, VS Code (opened from WSL), Go, the Go tools,
  golangci-lint, uv and Python reports the version pinned in appendix C; the public SSH key is added to your GitHub account.
- **You can explain:** why everything lives inside the WSL2 file system and not under the Windows drive; why a new
  terminal is needed after an install changes the PATH.

---

## Week 1 (days 1–5)

### Day 1 — How to study from this book (lighter day)
- **Objective:** know the method, the daily rhythm, the honesty rule and what to do when stuck; set up the notes.
- **Read:** ch 0, sections 0.1–0.10.
- **Time:** 75 min (read 40 / build 25 / note 10).
- **Exercise:** chapter 0 has no exercise slug. Create the Obsidian vault with the RTL Support plugin and the book's
  CSS snippet active (section 0.5), using the appendix D templates; a progress note listing the six checkpoints and
  ten milestones with their days; and a "stuck log" note you will append to whenever something takes more than
  twenty minutes. Choose your study days and fixed hour for the next two weeks.
- **Done when:** the vault opens with the plugin and snippet active, the progress note lists all six checkpoint days
  and M1–M10 and records your study days and hour, and the stuck log exists with the four entry headings.
- **You can explain:** what "done when" means and why the day ends there; the order in which to cut work when
  behind.

### Day 2 — What senior means in 2026
- **Objective:** read real postings like a hiring manager and see where you stand.
- **Read:** ch 1, sections 1.1–1.8.
- **Time:** 90 min (read 40 / build 40 / note 10).
- **Exercise:** `ch01-reading-postings`: if the book is not yet inside WSL, run the first two blocks of ch 2,
  section 2.9 (bring the book in, set `COURSE`); fill the one TODO in `skilltally` (`keyword_pattern`) until its
  tests pass; save five real current postings for senior backend or infrastructure roles as text files; run the tool
  over them; check two skills by hand; fill the skills-gap matrix (have / partial / missing). The commit waits for
  day 3.
- **Done when:** the tool's tests pass, five posting files exist with URL and date on the first line, `tally.md` is
  saved from your postings, and every gap-matrix row has a state, evidence and the chapter that closes it.
- **You can explain:** the difference between scope, ownership and judgment as senior signals; how your Rust and
  C++ background reads to an infrastructure hiring manager.

### Day 3 — The course repository and the smoke test
- **Objective:** a course repository that proves every Day 0 tool works end to end, and a first commit.
- **Read:** ch 2, section 2.9; ch 2, sections 2.11–2.12.
- **Time:** 90 min (read 25 / build 55 / note 10).
- **Exercise:** `ch02-workbench`, first-use half: clone the course repository inside WSL2; run `doctor.sh`; build
  and test the hello-world Go module; run the pytest smoke test through uv; run Docker's hello-world container;
  practise stopping, restarting and cleaning up Docker and WSL2 as section 2.12 describes; make the first commit (it includes ch01).
- **Done when:** `doctor.sh` reports every tool green, the Go test and the pytest smoke test pass, the hello-world
  container prints its greeting, and the first commit is in the log.
- **You can explain:** what `doctor.sh` checks and why it is reused later; how to free disk space taken by Docker
  safely.

### Day 4 — Go first steps: toolchain, modules, syntax, pointers
- **Objective:** write and run small Go programs, and see how Go's GC model differs from Rust ownership.
- **Read:** ch 3, sections 3.1–3.5.
- **Time:** 100 min (read 40 / build 50 / note 10).
- **Exercise:** `ch03-logstat`, part 1: create the module, read access-log lines from stdin or a file, parse one
  line into a struct, and print status-class counts; run the formatter and vet on it.
- **Done when:** `logstat` counts status classes on the sample log; formatting and vet report nothing.
- **You can explain:** values versus pointers in Go and when a value escapes to the heap; what Go gives up and gains
  compared with Rust ownership.

### Day 5 — Slices, maps, table tests, staticcheck
- **Objective:** finish a small, tested, lint-clean Go CLI.
- **Read:** ch 3, sections 3.6–3.10.
- **Time:** 110 min (read 35 / build 65 / note 10).
- **Exercise:** `ch03-logstat`, part 2: top paths with a map and a sorted slice; malformed lines counted and
  reported, not fatal; the provided table-driven tests pass (parser, stats, CLI, examples); escape-analysis output
  and one explaining sentence in the note; staticcheck and gofmt clean.
- **Done when:** all table tests pass including the malformed-line cases; staticcheck reports nothing; the README
  checklist of the exercise is ticked.
- **You can explain:** what a slice header is and why `append` can surprise you; why table-driven tests are the Go
  default.

---

## Week 2 (days 6–10)

### Day 6 — Types, interfaces and errors as values
- **Objective:** structs and methods, implicit interfaces, and error handling with wrapping.
- **Read:** ch 4, sections 4.1–4.5.
- **Time:** 100 min (read 40 / build 50 / note 10).
- **Exercise:** `ch04-ledger`, part 1: the account and transaction types, a `Store` interface with the in-memory
  implementation, a sentinel error for a missing account and a typed error for insufficient funds, both wrapped
  with context.
- **Done when:** tests show that `errors.Is` finds the sentinel and `errors.As` extracts the typed error through
  one level of wrapping.
- **You can explain:** implicit interfaces versus Rust traits; why Go wraps errors instead of returning `Result`.

### Day 7 — Panic, generics, embedding; the ledger CLI
- **Objective:** know when not to panic, use generics where they pay, compose by embedding.
- **Read:** ch 4, sections 4.6–4.9.
- **Time:** 100 min (read 30 / build 60 / note 10).
- **Exercise:** `ch04-ledger`, part 2: the JSON-file `Store`, a generic `Page[T]`, the CLI on top, tests that run
  the same suite against both store implementations.
- **Done when:** the shared test suite passes for both stores; the CLI can open an account, post transactions and
  page through history.
- **You can explain:** why `recover` belongs only at a goroutine or request boundary; what embedding promotes and
  what it does not.

### Day 8 — Goroutines, channels, select, context
- **Objective:** the concurrency primitives and cancellation by context.
- **Read:** ch 5, sections 5.1–5.4.
- **Time:** 100 min (read 45 / build 45 / note 10).
- **Exercise:** `ch05-linkcheck`, part 1: check a list of URLs concurrently with a per-request timeout taken from
  context; tests against an `httptest` server with a slow handler.
- **Done when:** the slow-handler test reports a timeout for that URL and success for the others, and the test
  finishes without waiting for the slow handler.
- **You can explain:** goroutines versus async/await and OS threads; how a context deadline reaches an HTTP call.

### Day 9 — sync, errgroup and concurrency patterns
- **Objective:** bounded concurrency with a worker pool and errgroup.
- **Read:** ch 5, sections 5.5–5.7.
- **Time:** 100 min (read 35 / build 55 / note 10).
- **Exercise:** `ch05-linkcheck`, part 2: limit concurrency to N workers; global cancellation on the first fatal
  error through errgroup; an ordered final report although results arrive out of order.
- **Done when:** a test proves that no more than N requests are in flight at once (counted in the test server) and
  that the report order matches the input order.
- **You can explain:** when to use a Mutex versus a channel; how errgroup combines waiting, the first error and
  cancellation.

### Day 10 — Races, leaks, finishing linkcheck (lighter day)
- **Objective:** find data races and goroutine leaks before they find you.
- **Read:** ch 5, sections 5.8–5.10.
- **Time:** 80 min (read 30 / build 40 / note 10).
- **Exercise:** `ch05-linkcheck`, part 3: run the race demo with and without the race detector and with both fixes;
  plant a shared counter in your own `Check`, see the race report, remove it; paste the planted leak from
  `data/planted-leak.txt`, read the leak report, restore the loop; run the full battery with `-count=3`.
- **Done when:** the whole test suite passes under the race detector and the leak check reports no goroutines left
  behind.
- **You can explain:** what the race detector can and cannot prove; the most common way goroutines leak.

---

## Week 3 (days 11–15)

### Day 11 — A production HTTP service: router, handlers, config
- **Objective:** start the service that becomes Hookline's skeleton.
- **Read:** ch 6, sections 6.1–6.3.
- **Time:** 100 min (read 35 / build 55 / note 10).
- **Exercise:** `ch06-shorty`, part 1: create, resolve and stats routes on the standard router, JSON in and out,
  an in-memory store behind an interface, configuration read from environment variables with defaults and
  validation.
- **Done when:** you can create a short link and resolve it with a manual request; a missing or invalid required
  variable stops startup with a clear message.
- **You can explain:** method and wildcard patterns in the Go router; why configuration is validated at startup.

### Day 12 — Structured logs, graceful shutdown, httptest
- **Objective:** a service that logs usefully, stops cleanly and is tested in-process.
- **Read:** ch 6, sections 6.4–6.6.
- **Time:** 110 min (read 35 / build 65 / note 10).
- **Exercise:** `ch06-shorty`, part 2: request-ID, logging and recovery middleware with `slog`; graceful shutdown on
  a termination signal that lets in-flight requests finish; handler and middleware tests with `httptest`.
- **Done when:** every log line of a request carries the same request ID; a test shows a request in flight during
  shutdown completes; a panicking handler returns a 500 and the server keeps serving.
- **You can explain:** what graceful shutdown waits for and what it abandons; why logs are structured key-values.

### Day 13 — Project layout, pprof, benchmarks
- **Objective:** look inside a running service and measure one hot path.
- **Read:** ch 6, sections 6.7–6.9.
- **Time:** 100 min (read 30 / build 60 / note 10).
- **Exercise:** `ch06-shorty`, part 3: implement `DebugHandler` on the separate debug port; write the Makefile of
  section 6.7 for the existing layout; capture a CPU profile under a small load; run the provided resolve benchmark
  and record your own numbers.
- **Done when:** the Makefile targets build, test and lint; a CPU profile is saved and its top function noted; the
  benchmark result is recorded in the note with the machine it ran on.
- **You can explain:** why pprof is not exposed on the public port; what a benchmark's ns/op and allocs/op mean.

### Day 14 — Checkpoint 1: a production Go service (lighter day)
- **Objective:** prove Part A: you can write, test and operate a small Go service.
- **Read:** ch 6, section 6.10.
- **Time:** 80 min (read 15 / build 55 / note 10).
- **Exercise:** work through the review list from a clean clone: (1) `logstat` table tests pass; (2) the ledger suite
  passes for both stores; (3) `linkcheck` passes under the race detector; (4) `shorty` starts from environment
  config, serves create/resolve/stats, logs with request IDs, survives a panicking handler and shuts down
  gracefully; (5) the benchmark runs; (6) answer the chapter 3–6 interview questions aloud.
- **Done when:** every item on the review list works without consulting the solutions; anything that failed is
  written in the stuck log with its fix.
- **You can explain:** the life of one HTTP request through shorty's middleware, handler and store, in two minutes.

### Day 15 — Senior Python: uv, ruff, mypy, advanced typing
- **Objective:** project hygiene a senior Python reviewer expects.
- **Read:** ch 7, sections 7.1–7.4.
- **Time:** 90 min (read 35 / build 45 / note 10).
- **Exercise:** `ch07-eventcheck`, part 1: read the uv project (lock file, `eventcheck` script entry, ruff and mypy
  strict config) and complete the reporters in `report.py`: the `SummaryDict` JSON shape (`TypedDict`) and the text
  reporter behind a `Protocol`.
- **Done when:** the project syncs from the lock file on a clean checkout (`uv sync --locked`), `tests/test_report.py`
  passes, and ruff and mypy pass.
- **You can explain:** what the lock file guarantees; when `Protocol` beats an abstract base class.

---

## Week 4 (days 16–20)

### Day 16 — pydantic, pytest and a measured profile
- **Objective:** finish a typed, packaged, tested CLI and measure before optimizing.
- **Read:** ch 7, sections 7.5–7.8.
- **Time:** 105 min (read 30 / build 65 / note 10).
- **Exercise:** `ch07-eventcheck`, part 2: pydantic v2 models with field and model validators, and the line checker
  that reports errors with line numbers; run the ready parametrized tests (fixtures, `tmp_path`, `monkeypatch`, a
  `slow` mark); one cProfile run on a generated 200,000-line file and two `timeit` measurements, with your measured
  numbers in the note.
- **Done when:** the CLI reports all six planted bad lines with line numbers and exits 1; all tests, ruff and mypy
  pass; the profile output is saved.
- **You can explain:** what pydantic validates that type hints alone do not; why you profile before changing code.

### Day 17 — asyncio: the loop, TaskGroup, cancellation, timeouts
- **Objective:** asyncio done right, including cancellation and bounded concurrency.
- **Read:** ch 8, sections 8.1–8.3.
- **Time:** 90 min (read 40 / build 40 / note 10).
- **Exercise:** `ch08-receiver`, warm-up from the chapter: a small script that fetches several slow fake calls in a
  `TaskGroup` with a per-call timeout and a `Semaphore` limit, and shows what cancellation does to the siblings.
- **Done when:** the script's output shows the limit respected and the siblings cancelled when one task fails.
- **You can explain:** what blocks the event loop; how `TaskGroup` differs from `gather` on failure.

### Day 18 — FastAPI and the webhook receiver
- **Objective:** a FastAPI service with lifespan resources and async tests; the future Hookline `sink`.
- **Read:** ch 8, sections 8.4–8.8.
- **Time:** 120 min (read 40 / build 70 / note 10).
- **Exercise:** `ch08-receiver`: routes, models and dependencies; resources opened in lifespan; HMAC-SHA256
  signature verification; dedupe by event id; recorded deliveries; configurable failure rate and latency; async
  tests with httpx.
- **Done when:** tests show a bad signature rejected, a duplicate event recorded once, and the failure-rate setting
  producing errors; the service runs locally through uv.
- **You can explain:** why the signature is checked over the raw body; why a receiver must dedupe.

### Day 19 — REST as a contract, OpenAPI, problem details (lighter day)
- **Objective:** contract-first API design with one error shape.
- **Read:** ch 9, sections 9.1–9.3.
- **Time:** 80 min (read 35 / build 35 / note 10).
- **Exercise:** `ch09-catalog-api`, part 1: read the given OpenAPI 3.1 spec, add the problem-details error schema to
  it, and implement two catalog routes plus the shared error writer.
- **Done when:** every error path of the two routes returns a problem-details body that matches the spec.
- **You can explain:** why the spec is the source of truth and not the code; the fields of a problem-details error.

### Day 20 — Pagination, idempotency keys, versioning
- **Objective:** the API mechanics Hookline relies on.
- **Read:** ch 9, sections 9.4–9.6.
- **Time:** 110 min (read 35 / build 65 / note 10).
- **Exercise:** `ch09-catalog-api`, part 2: cursor pagination on the list route; the in-memory idempotency-key
  middleware that replays the stored response for a repeated key and rejects the same key with a different body.
- **Done when:** tests show stable pages while items are inserted, a replayed response for a repeated key, a
  422 for a reused key with a different body, and a 409 for a concurrent request with the same key.
- **You can explain:** why offset pagination breaks under inserts; what an idempotency key promises the client.

---

## Week 5 (days 21–25)

### Day 21 — gRPC, when not gRPC, input limits
- **Objective:** know where gRPC fits and bound every input.
- **Read:** ch 9, sections 9.7–9.9.
- **Time:** 105 min (read 35 / build 60 / note 10).
- **Exercise:** `ch09-catalog-api`, part 3: generate the tiny `Ping`/`GetItem` gRPC service with buf and call it
  from a test; add request size limits and input validation to the HTTP routes.
- **Done when:** the gRPC test passes; an oversized body and an invalid field both get problem-details errors; all 31 tests pass with `-race`
  and regenerating with buf leaves the committed code unchanged.
- **You can explain:** two cases for gRPC and two against it; why size limits belong before parsing.

### Day 22 — Hookline specification and M1 start
- **Objective:** understand the whole capstone, then start its API from your chapter 6 service.
- **Read:** ch 37, sections 37.1–37.8; ch 38, section 38.1.
- **Time:** 120 min (read 55 / build 55 / note 10).
- **Exercise:** Hookline M1, module `api`: set up the starter as section 37.8 describes; bring in the chapter 6
  server skeleton (config, middleware, shutdown); write the OpenAPI contract for tenants, endpoints and
  subscriptions, with the problem-details error schema from chapter 9.
- **Done when:** the `api` module builds and serves a health route; the OpenAPI file lints and covers create, get
  and list for the three resources.
- **You can explain:** Hookline's delivery guarantee in one sentence; the path of one event from ingest to the
  receiver.

### Day 23 — M1: API skeleton with in-memory resources
- **Objective:** a working, tested API skeleton that matches its contract.
- **Read:** ch 38, section 38.1.
- **Time:** 110 min (read 15 / build 85 / note 10).
- **Exercise:** Hookline M1, module `api`: tenants, endpoints and subscriptions in in-memory stores behind
  interfaces; handlers for create, get and list; every error as problem details; handler tests that check
  responses against the contract.
- **Done when:** the M1 acceptance list in section 38.1 passes; an unknown tenant, an invalid endpoint URL and a
  malformed body each return the documented problem-details error.
- **You can explain:** why the stores sit behind interfaces now although they are in memory; which M1 choices M2
  will replace.

### Day 24 — PostgreSQL in a container, EXPLAIN, B-tree indexes (lighter day)
- **Objective:** read a query plan on realistic data.
- **Read:** ch 10, sections 10.1–10.4.
- **Time:** 85 min (read 35 / build 40 / note 10).
- **Exercise:** `ch10-orders-db`, part 1: start PostgreSQL with the compose file, connect with psql, run the seed
  script that builds the orders dataset; run EXPLAIN ANALYZE on the first two slow queries, add a B-tree index,
  record your before and after timings.
- **Done when:** the dataset is loaded, both plans are saved before and after the index, and your own timings are in
  the note.
- **You can explain:** sequential scan versus index scan in a plan; why EXPLAIN without ANALYZE can mislead.

### Day 25 — Composite, partial and covering indexes; pgx
- **Objective:** choose the right index and access PostgreSQL from Go.
- **Read:** ch 10, sections 10.5–10.7.
- **Time:** 115 min (read 35 / build 70 / note 10).
- **Exercise:** `ch10-orders-db`, part 2: fix the remaining three slow queries with composite, partial or covering
  indexes, recording your timings; fix the N+1 example; write the pgx repository with keyset pagination and its
  tests.
- **Done when:** all five queries have a recorded before and after plan; the repository tests pass against the
  container; the `cmd/nplus1` output for customer 1 is in the note.
- **You can explain:** column order in a composite index; how keyset pagination avoids scanning skipped rows.

---

## Week 6 (days 26–30)

### Day 26 — Transactions, isolation levels, demonstrated anomalies
- **Objective:** see lost update and write skew happen, not just read about them.
- **Read:** ch 11, sections 11.1–11.3.
- **Time:** 110 min (read 40 / build 60 / note 10).
- **Exercise:** `ch11-isolation-lab`, part 1: Go tests that open two connections and reproduce a lost update under
  read committed and write skew under repeatable read.
- **Done when:** both anomaly tests reproduce the anomaly reliably and assert on it.
- **You can explain:** what each isolation level forbids in PostgreSQL; why write skew survives repeatable read.

### Day 27 — Row locks, SKIP LOCKED, deadlocks (lighter day)
- **Objective:** fix the anomalies with locks and build a work queue.
- **Read:** ch 11, sections 11.4–11.5.
- **Time:** 85 min (read 30 / build 45 / note 10).
- **Exercise:** `ch11-isolation-lab`, part 2: fix the lost update with `FOR UPDATE`; build the `SKIP LOCKED` work
  queue whose test runs eight workers over 300 jobs; an ordered `Transfer` that never deadlocks; provoke the chapter's deadlock and read PostgreSQL's report.
- **Done when:** the fix tests pass (`FOR UPDATE` provably waits; 25 concurrent withdrawals leave
  exactly 20 successes); no job is handled twice; the deadlock
  error is captured in the note.
- **You can explain:** what `SKIP LOCKED` buys a queue; how consistent lock ordering prevents deadlocks.

### Day 28 — Migrations with goose, pooling, serialization retries
- **Objective:** schema changes as code and the retry loop SERIALIZABLE needs.
- **Read:** ch 11, sections 11.6–11.8.
- **Time:** 110 min (read 35 / build 65 / note 10).
- **Exercise:** `ch11-isolation-lab`, part 3: goose migrations for the lab schema (up and down); read the pool's
  explicit limits and work out how many replicas fit under `max_connections`; fix write skew with SERIALIZABLE plus the retry-on-serialization-failure helper.
- **Done when:** migrations apply and roll back cleanly on an empty database; the write-skew test passes as fixed
  and the helper's test shows a retry happening.
- **You can explain:** why a retry must re-run the whole transaction; how pool size interacts with PostgreSQL's
  connection limit.

### Day 29 — M2: PostgreSQL schema, migrations, repositories
- **Objective:** move Hookline's state from memory to PostgreSQL.
- **Read:** ch 38, section 38.2.
- **Time:** 115 min (read 20 / build 85 / note 10).
- **Exercise:** Hookline M2, module `api`: goose migrations for tenants, endpoints, subscriptions and events from
  the data model in ch 37; PostgreSQL in the project compose file; pgx repositories that replace the in-memory
  stores behind the same interfaces.
- **Done when:** migrations apply from empty and roll back; the M1 handler tests pass unchanged against the pgx
  repositories.
- **You can explain:** why the interfaces from M1 made this change small; which constraints live in the database
  and why.

### Day 30 — M2: idempotent event ingest and cursor pagination
- **Objective:** ingest that is safe to retry, and lists that page correctly.
- **Read:** ch 38, section 38.2.
- **Time:** 115 min (read 15 / build 90 / note 10).
- **Exercise:** Hookline M2, module `api`: the event ingest route with an `Idempotency-Key` header, a unique
  constraint on tenant and key, and the stored response written in the same transaction as the event; cursor
  pagination on the list routes.
- **Done when:** the M2 acceptance list in section 38.2 passes; a test sends the same key concurrently and finds
  exactly one event row and identical responses.
- **You can explain:** why the unique constraint, not application code, decides the duplicate; why the key and the
  event share one transaction.

---

## Week 7 (days 31–35)

### Day 31 — Redis: data structures, cache-aside, TTL and invalidation
- **Objective:** caching patterns and their failure modes.
- **Read:** ch 12, sections 12.1–12.3.
- **Time:** 100 min (read 35 / build 55 / note 10).
- **Exercise:** `ch12-cache-aside`, part 1: Redis in a container; cache-aside with go-redis over a self-contained
  products repository that mirrors chapter 10's; TTL and invalidation on write; tests against the real Redis that skip with a message when it is
  absent.
- **Done when:** a test shows a miss then a hit, and a write that invalidates the cached value.
- **You can explain:** cache-aside versus write-through; why invalidation after the write, not before.

### Day 32 — Stampedes, singleflight, atomic counters, when not to cache
- **Objective:** survive a cold cache under concurrency.
- **Read:** ch 12, sections 12.4–12.6.
- **Time:** 100 min (read 35 / build 55 / note 10).
- **Exercise:** `ch12-cache-aside`, part 2: the stampede demo with N concurrent misses counted at the repository;
  the singleflight fix; an atomic counter with INCR; a fixed-window Lua rate limiter; the full `-race` run.
- **Done when:** the demo prints the repository call count without and with singleflight, both from your run, and a
  test asserts one call with singleflight.
- **You can explain:** what singleflight collapses and what it does not (other processes); two cases where a cache
  makes things worse.

### Day 33 — M3: Redis subscription cache with singleflight
- **Objective:** cache the subscription lookup that every event needs.
- **Read:** ch 38, section 38.3.
- **Time:** 110 min (read 15 / build 85 / note 10).
- **Exercise:** Hookline M3, module `api`: Redis in the project compose; a subscription cache for "which
  subscriptions match this tenant and event type", cache-aside with TTL, invalidated when a subscription changes,
  misses collapsed with singleflight.
- **Done when:** tests show a hit after the first lookup, invalidation on subscription change, and one database call
  for concurrent misses; with Redis stopped, lookups fall back to PostgreSQL.
- **You can explain:** why a stale subscription cache is a correctness bug here, not only a performance one; what
  happens when Redis is down.

### Day 34 — M3: the receiver simulator `sink` in compose
- **Objective:** a misbehaving receiver ready for every later milestone.
- **Read:** ch 38, section 38.3.
- **Time:** 105 min (read 15 / build 80 / note 10).
- **Exercise:** Hookline M3, module `sink`: bring the chapter 8 receiver into the project as `sink`; add it to
  compose with its failure rate and latency configurable from the environment; expose its counts of received and
  deduplicated events; a smoke script that posts a signed sample event to it.
- **Done when:** the M3 acceptance list in section 38.3 passes; the whole compose stack (api, PostgreSQL, Redis,
  sink) starts with one command and the smoke script sees the event counted.
- **You can explain:** why the receiver misbehaves on purpose; what the sink's counts will later prove.

### Day 35 — Review and consolidation: M1–M3 (lighter day)
- **Objective:** close the debt from Part C so far before the event pipeline.
- **Read:** ch 38, sections 38.1–38.3.
- **Time:** 80 min (read 20 / build 50 / note 10).
- **Exercise:** no new slug: tear the compose stack down including volumes and bring it up from scratch; re-run the
  M1–M3 acceptance lists; fix anything recorded in the stuck log; answer the chapter 9–12 interview questions aloud.
- **Done when:** all three acceptance lists pass from a clean start and the stuck log has no open item.
- **You can explain:** the ingest path from request to committed row, including idempotency and the cache, without
  notes.

---

## Week 8 (days 36–40)

### Day 36 — Log versus queue; Redpanda, topics and partitions (lighter day)
- **Objective:** Kafka's model in practice.
- **Read:** ch 13, sections 13.1–13.3.
- **Time:** 85 min (read 40 / build 35 / note 10).
- **Exercise:** `ch13-outbox`, part 1: run Redpanda in a container; with rpk, create a topic with several
  partitions, produce keyed messages and consume them; note which partition each key landed on.
- **Done when:** messages with the same key are seen on the same partition, in order, in your own run.
- **You can explain:** how a log differs from a queue; what a partition orders and what it does not.

### Day 37 — Producers, consumer groups, offsets, rebalance
- **Objective:** at-least-once processing with franz-go's semantics.
- **Read:** ch 13, sections 13.4–13.6.
- **Time:** 110 min (read 40 / build 60 / note 10).
- **Exercise:** `ch13-outbox`, part 2: read and run the provided Go producer (keys, acks); complete the consumer
  loop so it commits offsets only after processing; start a second consumer and watch the rebalance.
- **Done when:** after killing a consumer mid-batch, the restarted group reprocesses the uncommitted messages and
  none are lost.
- **You can explain:** why committing after processing gives at-least-once; what a rebalance does to in-flight work.

### Day 38 — The outbox pattern and the exactly-once contract
- **Objective:** publish exactly what was committed, no more and no less.
- **Read:** ch 13, sections 13.7–13.9.
- **Time:** 115 min (read 35 / build 70 / note 10).
- **Exercise:** `ch13-outbox`, part 3: an outbox row written in the same transaction as the business row; a relay
  that publishes pending rows and marks them; a consumer that dedupes; the crash-in-the-middle test.
- **Done when:** the crash test kills the relay between publish and mark and shows a duplicate published but none
  processed twice, and no business row without its message.
- **You can explain:** the dual-write problem the outbox solves; why "exactly once" is a contract between producer
  and consumer, not a broker feature.

### Day 39 — M4: transactional outbox and relay to Redpanda
- **Objective:** every ingested event reaches the log, even across crashes.
- **Read:** ch 38, section 38.4.
- **Time:** 115 min (read 15 / build 90 / note 10).
- **Exercise:** Hookline M4, modules `api` and `relay`: a migration for the outbox table; ingest writes the outbox
  row in the same transaction as the event and its idempotency record; the `relay` publishes pending rows to the
  events topic keyed by tenant and marks them; Redpanda added to compose.
- **Done when:** an ingested event appears on the topic; killing the relay and restarting it publishes the
  remaining rows; no event exists without an outbox row.
- **You can explain:** why ingest never talks to Redpanda directly; what ordering the tenant key gives.

### Day 40 — M4: dispatcher fans events into deliveries
- **Objective:** one event becomes one delivery per subscribed endpoint.
- **Read:** ch 38, section 38.4.
- **Time:** 110 min (read 15 / build 85 / note 10).
- **Exercise:** Hookline M4, module `dispatcher`: consume the events topic in a group; look up matching
  subscriptions through the M3 cache; insert one delivery row per subscription, idempotently on event and
  subscription; commit offsets after the rows commit.
- **Done when:** the M4 acceptance list in section 38.4 passes; replaying the same events creates no extra delivery
  rows; a test with three subscriptions yields three deliveries.
- **You can explain:** why delivery rows are created idempotently; what happens if the dispatcher crashes after the
  insert but before the offset commit.

---

## Week 9 (days 41–45)

### Day 41 — Background jobs on PostgreSQL, backoff, jitter, idempotent consumers
- **Objective:** durable jobs with scheduled retries.
- **Read:** ch 14, sections 14.1–14.4.
- **Time:** 110 min (read 35 / build 65 / note 10).
- **Exercise:** `ch14-jobqueue`, part 1: the provided job table (`run_at`, attempts) and the retry branch of `Fail`; claiming with `SKIP
  LOCKED`; the backoff function with exponential growth, a cap and full jitter, with tests; an idempotent handler.
- **Done when:** backoff tests check the bounds of each attempt's delay and the cap; a failing job is rescheduled
  with a later `run_at`.
- **You can explain:** why full jitter beats fixed backoff for many clients; what makes a handler idempotent.

### Day 42 — Dead letters, replay, schedulers, draining workers
- **Objective:** finish the job queue's lifecycle.
- **Read:** ch 14, sections 14.5–14.7.
- **Time:** 105 min (read 35 / build 60 / note 10).
- **Exercise:** `ch14-jobqueue`, part 2: move a job to the dead state after the maximum attempts; a replay
  operation; read the scheduler and run its tests; graceful drain of workers on shutdown.
- **Done when:** tests show a job going dead after the last attempt, replay making it runnable again, and shutdown
  waiting for the claimed job.
- **You can explain:** why dead letters need a replay path; what a worker must do with a claimed job when told to
  stop.

### Day 43 — M5: delivery worker, signature, timeouts, attempt log
- **Objective:** deliver signed webhooks and record every attempt.
- **Read:** ch 38, section 38.5.
- **Time:** 115 min (read 15 / build 90 / note 10).
- **Exercise:** Hookline M5, module `worker`: claim due deliveries with `SKIP LOCKED`; POST the payload to the
  endpoint with an HMAC-SHA256 signature header and a timestamp; a strict request timeout; write one attempt-log
  row per try with status, latency and error.
- **Done when:** the sink verifies the signature and counts the delivery; a sink set to slow responses produces a
  timeout attempt row.
- **You can explain:** why the signature covers a timestamp; why the attempt log is append-only.

### Day 44 — M5: retries with backoff and jitter, dead state, replay API
- **Objective:** at-least-once delivery that gives up honestly and can be replayed.
- **Read:** ch 38, section 38.5.
- **Time:** 115 min (read 15 / build 90 / note 10).
- **Exercise:** Hookline M5, modules `worker` and `api`: on failure, set `next_attempt_at` from exponential backoff
  with full jitter (the chapter 14 function); after the maximum attempts, mark the delivery dead; a replay route in
  `api` that makes dead deliveries due again.
- **Done when:** the M5 acceptance list in section 38.5 passes; with the sink failing most requests, deliveries
  retry, some succeed, the rest go dead, and replay after fixing the sink delivers them.
- **You can explain:** why retry state lives in the delivery row and not in memory; what a receiver sees on a
  replay and why it must dedupe.

### Day 45 — Checkpoint 2: backend foundations and the delivery path (lighter day)
- **Objective:** prove Part C: an idempotent, durable event pipeline you can explain.
- **Read:** ch 38, sections 38.1–38.5.
- **Time:** 90 min (read 15 / build 65 / note 10).
- **Exercise:** work through the review list from a clean compose start: (1) M1–M5 acceptance lists pass; (2) a
  repeated `Idempotency-Key` creates one event; (3) killing the relay loses no event; (4) killing the dispatcher
  creates no duplicate deliveries; (5) a failing sink drives retries, dead state and a successful replay; (6) the
  isolation lab and job-queue tests pass; (7) answer the chapter 9–14 interview questions aloud.
- **Done when:** every item on the review list works; the note contains one end-to-end diagram of an event's path
  drawn from memory.
- **You can explain:** where each delivery guarantee comes from (unique constraint, outbox, idempotent fan-out,
  retry state).

---

## Week 10 (days 46–50)

### Day 46 — Partial failure, timeouts and time budgets, retries (lighter day)
- **Objective:** the failure model of distributed calls.
- **Read:** ch 15, sections 15.1–15.4.
- **Time:** 85 min (read 40 / build 35 / note 10).
- **Exercise:** `ch15-resilience-lab`, part 1: run the flaky upstream; add deadline propagation from the client's
  incoming request to its upstream call; run the lab script for the retry-storm scenario and record its call counts.
- **Done when:** the lab script output for the retry storm is saved from your run, and a test shows a call cut off
  when the remaining budget runs out.
- **You can explain:** why each hop needs a smaller timeout than its caller; how naive retries multiply load.

### Day 47 — Circuit breakers, bulkheads, graceful degradation
- **Objective:** stop calling what is down and contain the damage.
- **Read:** ch 15, sections 15.5–15.7.
- **Time:** 110 min (read 35 / build 65 / note 10).
- **Exercise:** `ch15-resilience-lab`, part 2: a retry budget; a circuit breaker with closed, open and half-open
  states; a bulkhead limiting concurrent calls; run the lab script with and without each mechanism.
- **Done when:** breaker state tests pass; the lab script's call counts for each configuration are recorded from
  your run.
- **You can explain:** the three breaker states and what moves between them; what a bulkhead isolates.

### Day 48 — Overload, token bucket, distributed rate limiting
- **Objective:** limit rate locally and across instances.
- **Read:** ch 16, sections 16.1–16.3.
- **Time:** 110 min (read 35 / build 65 / note 10).
- **Exercise:** `ch16-ratelimit`, part 1: a token bucket with an injected clock so tests run instantly; the Redis
  Lua limiter with tests against a real Redis; run the queue simulation and record its table; record the naive
  read-then-write overshoot; a first unprotected load run and per-tenant limiting through the server.
- **Done when:** clock-driven tests cover refill, burst and empty bucket; the Lua limiter test shows two clients
  sharing one limit.
- **You can explain:** why the clock is injected; why the Redis limiter is a Lua script and not two commands.

### Day 49 — Back-pressure, load shedding, bounded queues
- **Objective:** say no early instead of failing late.
- **Read:** ch 16, sections 16.4–16.6.
- **Time:** 100 min (read 35 / build 55 / note 10).
- **Exercise:** `ch16-ratelimit`, part 2: a bounded queue with blocking and rejecting policies and the pipeline run; a load-shedding middleware with a concurrency limit that returns 503 with
  `Retry-After`; run the small load script against it with and without shedding and record what you observe; a priority run with reserved
  slots.
- **Done when:** a test shows requests above the limit shed immediately; your load-script results are in the note.
- **You can explain:** rate limiting versus load shedding; why an unbounded queue only hides overload.

### Day 50 — Consistency models, replication, replica lag
- **Objective:** see replication lag and read-your-writes in practice.
- **Read:** ch 17, sections 17.1–17.4.
- **Time:** 115 min (read 40 / build 65 / note 10).
- **Exercise:** `ch17-replica-lag`, part 1: the compose file with a primary and a streaming replica; the Go probe
  that writes on the primary, reads back on the replica and reports the lag it observes; then the two read-your-writes fixes, session stickiness
  (`ForRead`) and an LSN token (`ReadNote`).
- **Done when:** the probe runs and its observed lag values from your machine are in the note; you can make a read
  miss its own write by reading from the replica.
- **You can explain:** eventual versus read-your-writes consistency; one way to give a user read-your-writes with a
  replica.

---

## Week 11 (days 51–55)

### Day 51 — Partitioning, hot partitions, CAP and PACELC (lighter day)
- **Objective:** choose partition keys and reason about trade-offs without myths.
- **Read:** ch 17, sections 17.5–17.7.
- **Time:** 85 min (read 40 / build 35 / note 10).
- **Exercise:** `ch17-replica-lag`, part 2: run the partition-key skew simulator with uniform, skewed and composite keys plus a
  range-partitioner run; complete `Ring.Node`; record the per-partition load it prints.
- **Done when:** the four simulator runs are saved and the note says which key you would choose for Hookline's
  events and why.
- **You can explain:** what makes a partition hot; what PACELC adds to CAP.

### Day 52 — Distributed locks, advisory locks, fencing tokens
- **Objective:** know why distributed locks are hard and how to make them safe enough.
- **Read:** ch 18, sections 18.1–18.4.
- **Time:** 110 min (read 45 / build 55 / note 10).
- **Exercise:** `ch18-leader`, part 1: two psql sessions contending for a session advisory lock and the `campaign`
  function; a Redis lock held by two at once; a fencing token checked by the resource on every write.
- **Done when:** a test shows a stale holder's write rejected by the fencing check.
- **You can explain:** why a lock with a TTL can be held by two processes at once; how a fencing token fixes it.

### Day 53 — Leader election, consensus idea, sagas
- **Objective:** a single active worker and a compensating multi-step operation.
- **Read:** ch 18, sections 18.5–18.7.
- **Time:** 110 min (read 40 / build 60 / note 10).
- **Exercise:** `ch18-leader`, part 2: leader election with the advisory lock (kill the leader, the other process
  takes over); an orchestrated saga with compensation and its tests.
- **Done when:** the takeover is shown in the logs of your run; saga tests cover a failure at each step with the
  right compensations.
- **You can explain:** why a session advisory lock is released when the process dies; when a saga is the right tool
  and when a transaction is.

### Day 54 — M6: per-endpoint circuit breaker and bulkhead
- **Objective:** one bad endpoint must not slow everyone else.
- **Read:** ch 38, section 38.6.
- **Time:** 115 min (read 15 / build 90 / note 10).
- **Exercise:** Hookline M6, module `worker`: a circuit breaker per endpoint (an open breaker defers the delivery to
  a later `next_attempt_at` without calling); a bulkhead limiting concurrent deliveries per endpoint.
- **Done when:** with one sink endpoint failing and a second healthy, the healthy endpoint's deliveries keep flowing
  and the failing one's breaker opens, as shown in the attempt log of your run.
- **You can explain:** why the breaker is per endpoint and not global; why an open breaker reschedules instead of
  dropping.

### Day 55 — M6: tenant rate limits, load shedding, single active relay
- **Objective:** protect ingest and keep the relay single.
- **Read:** ch 38, section 38.6.
- **Time:** 120 min (read 15 / build 95 / note 10).
- **Exercise:** Hookline M6, modules `api` and `relay`: a per-tenant Redis token bucket on ingest returning 429 with
  `Retry-After`; load shedding on ingest with a concurrency limit; the relay takes a PostgreSQL advisory lock so only
  one instance publishes.
- **Done when:** the M6 acceptance list in section 38.6 passes; a noisy tenant is limited while a quiet one is not;
  with two relays running, only one publishes, and killing it lets the other take over.
- **You can explain:** where each M6 mechanism sits on the event's path; why the relay needs exactly one active
  instance.

---

## Week 12 (days 56–60)

### Day 56 — Checkpoint 3: distributed-systems resilience (lighter day)
- **Objective:** prove Part D: Hookline isolates failures and bounds load.
- **Read:** ch 38, section 38.6.
- **Time:** 90 min (read 15 / build 65 / note 10).
- **Exercise:** work through the review list from a clean compose start: (1) M1–M6 acceptance lists pass; (2) one
  failing endpoint does not stall a healthy one; (3) a noisy tenant gets 429 and a quiet tenant does not; (4) ingest
  sheds under the concurrency limit; (5) two relays run, one publishes, takeover works; (6) the resilience,
  rate-limit, replica-lag and leader labs pass; (7) answer the chapter 15–18 interview questions aloud.
- **Done when:** every item on the review list works; the note lists each failure mode of section 37.6 with the
  mechanism that handles it.
- **You can explain:** how Hookline behaves when one receiver is down for an hour.

### Day 57 — Installations for Part E: kind, kubectl, Helm, Terraform, k6
- **Objective:** install and verify the Part E tools a day before first use.
- **Read:** ch 2, section 2.10.
- **Time:** 75 min (read 20 / install 45 / note 10).
- **Exercise:** `ch02-workbench`: install kind, kubectl, Helm, Terraform and k6 inside WSL2 at the versions in
  appendix C, verifying each download as the chapter shows; extend `doctor.sh` with their version checks. Do not
  create a cluster yet; that is day 60.
- **Done when:** `doctor.sh` reports every Day 0 and Part E tool at its pinned version.
- **You can explain:** why the kind node image is pinned by digest; why install day and first-use day are kept
  apart.

### Day 58 — Docker images: layers, multi-stage, distroless, Python with uv
- **Objective:** small, non-root images for Go and Python.
- **Read:** ch 19, sections 19.1–19.4.
- **Time:** 110 min (read 40 / build 60 / note 10).
- **Exercise:** `ch19-images`, part 1: a multi-stage Dockerfile for shorty on a distroless base running as non-root;
  an image for the chapter 8 receiver built with uv; record the sizes you measure for each, and for a naive
  single-stage build of shorty.
- **Done when:** both images run as a non-root user and serve requests; your measured sizes are in the note.
- **You can explain:** how layer order affects the build cache; what distroless leaves out and what that costs when
  debugging.

### Day 59 — Compose for development, scanning, SBOM, tags
- **Objective:** know what is inside an image and name it properly.
- **Read:** ch 19, sections 19.5–19.7.
- **Time:** 100 min (read 30 / build 60 / note 10).
- **Exercise:** `ch19-images`, part 2: run both images together with compose, health-checked; scan them with Trivy
  and with Grype and generate an SBOM with Syft, set up as the chapter describes; tag with a version and the commit.
- **Done when:** `check_images.sh` ends in `PASS`; scan reports and SBOM files are saved; the note records at least
  one finding and what you decided about it.
- **You can explain:** why `latest` is not a deploy tag; what an SBOM is for after an incident.

### Day 60 — Kubernetes on kind: pods, deployments, ReplicaSets
- **Objective:** a first real cluster and the objects that run workloads.
- **Read:** ch 20, sections 20.1–20.4.
- **Time:** 100 min (read 40 / build 50 / note 10).
- **Exercise:** `ch20-k8s-basics`, part 1: create a kind cluster with the pinned node image; load the shorty and
  receiver images into it; write a Deployment; delete a pod and watch the ReplicaSet replace it.
- **Done when:** the Deployment reports all replicas ready and a deleted pod is replaced automatically.
- **You can explain:** pod versus Deployment versus ReplicaSet; why the image must be loaded into kind.

---

## Week 13 (days 61–65)

### Day 61 — Services, DNS, ConfigMaps, Secrets, probes (lighter day)
- **Objective:** make shorty reachable, configurable and health-checked.
- **Read:** ch 20, sections 20.5–20.7.
- **Time:** 85 min (read 30 / build 45 / note 10).
- **Exercise:** `ch20-k8s-basics`, part 2: a Service; configuration from a ConfigMap and a Secret; startup,
  liveness and readiness probes; port-forward and call it; a one-off pod that reaches the receiver by DNS name.
- **Done when:** calls through the port-forward succeed; breaking the readiness route removes the pod from the
  Service endpoints without restarting it; the sender gets 200 from the receiver; `run_all.sh` ends in `PASS`.
- **You can explain:** readiness versus liveness; why a Kubernetes Secret is not encryption by itself.

### Day 62 — Requests and limits, HPA, rollouts and rollbacks
- **Objective:** run workloads with resources, scale and safe releases.
- **Read:** ch 21, sections 21.1–21.3.
- **Time:** 110 min (read 35 / build 65 / note 10).
- **Exercise:** `ch21-ops-helm`, part 1: requests and limits for shorty and the three QoS classes; throttling
  measured under three CPU limits and an OOM kill; metrics-server and an HPA under in-cluster load; a deliberately
  broken rollout and its undo.
- **Done when:** the three QoS classes appear and the three throttling times are in the note; a shorty pod shows
  `OOMKilled` with exit code 137; the HPA reports current utilization, scales up under load and back down; the
  broken rollout stalls with old pods still serving, and the undo restores the previous revision.
- **You can explain:** what happens when a container exceeds its memory limit versus its CPU limit; why a rolling
  update with readiness probes protects users.

### Day 63 — Ingress, debugging drill, Helm chart
- **Objective:** debug a cluster methodically and package an app with Helm.
- **Read:** ch 21, sections 21.4–21.7.
- **Time:** 120 min (read 35 / build 75 / note 10).
- **Exercise:** `ch21-ops-helm`, part 2: an ingress on kind; the debugging drill with four planted faults, using
  describe, logs, events and an ephemeral debug container; a Helm chart for shorty with a values schema, validated
  with lint and template; install, upgrade, a failed upgrade that rolls itself back, rollback and uninstall; the
  smoke test and cleanup.
- **Done when:** all four planted faults are found and fixed with the evidence that pointed to each in the note;
  the chart lints, renders and installs; `check_static.sh` and `smoke.sh` pass and the cluster is deleted.
- **You can explain:** your order of checks when a pod will not start; what Helm values and templates separate.

### Day 64 — Terraform: providers, resources, state, plan and apply
- **Objective:** infrastructure as code against the local cluster.
- **Read:** ch 22, sections 22.1–22.3.
- **Time:** 100 min (read 40 / build 50 / note 10).
- **Exercise:** `ch22-terraform`, part 1: a configuration with the kubernetes provider against kind that creates a
  namespace and a ConfigMap; plan, apply, change one value, plan again, drift made by hand with `kubectl edit`,
  destroy.
- **Done when:** the second plan shows exactly the one change you made; the plan reveals the manual edit; destroy
  leaves the namespace gone.
- **You can explain:** what state records and why losing it is dangerous; what a plan shows before anything changes.

### Day 65 — Variables, modules, remote state, the Helm provider
- **Objective:** reusable Terraform that installs a chart.
- **Read:** ch 22, sections 22.4–22.7.
- **Time:** 110 min (read 40 / build 60 / note 10).
- **Exercise:** `ch22-terraform`, part 2: variables and outputs; a module; the shorty chart released through the helm
  provider; read the remote-state and locking section. The AWS free-tier section is optional and can be skipped.
- **Done when:** one apply creates the namespace and the shorty release through the module, and destroy removes
  both; `check_static.sh` and `smoke.sh` pass, the cluster is deleted, and no state file is in Git.
- **You can explain:** why teams keep state remotely with locking; when Terraform should and should not manage Helm
  releases.

---

## Week 14 (days 66–70)

### Day 66 — GitHub Actions: workflows, build and test, service containers, images
- **Objective:** CI that tests with a real database and builds an image.
- **Read:** ch 23, sections 23.1–23.4.
- **Time:** 110 min (read 35 / build 65 / note 10).
- **Exercise:** `ch23-ci`, part 1: workflows for the chapter 6 and chapter 8 exercises; the chapter 11 workflow
  with a PostgreSQL service container and fail-on-skip; an image job that publishes to GHCR from `main` only; the
  local run of every step with the image build (`run_ci_locally.sh --image`).
- **Done when:** `tests/lint.sh` (actionlint, shellcheck, permissions) is clean on all three workflows and the local
  run is green with no skipped tests in ch11.
- **You can explain:** jobs versus steps; how a service container is reached from a job; why green without a
  database is a failure.

### Day 67 — Deploying to kind in CI, environments, secrets, releases (lighter day)
- **Objective:** a release pipeline with least privilege.
- **Read:** ch 23, sections 23.5–23.8.
- **Time:** 85 min (read 35 / build 40 / note 10).
- **Exercise:** `ch23-ci`, part 2: a job that creates kind, deploys the image and runs a smoke test before the image
  is published; a release workflow on tag that goes through the `production` environment, with every action pinned
  to a full SHA; the actionlint drill on the broken workflow (six findings); optionally, a first run in your own
  GitHub repository.
- **Done when:** the static checks are clean and every workflow declares its permissions; the local `--kind` run
  ends with its summary and leaves no cluster; the six broken-workflow findings and their fixes are in the note.
- **You can explain:** why permissions are written in the file and start from read-only; what an environment
  protects; major-version tags versus SHA pinning.

### Day 68 — M7: distroless images for every Hookline service
- **Objective:** Hookline built as small, non-root images.
- **Read:** ch 38, section 38.7.
- **Time:** 110 min (read 15 / build 85 / note 10).
- **Exercise:** Hookline M7, module `deploy` with `api`, `relay`, `dispatcher`, `worker` and `sink`: multi-stage
  distroless non-root images for the Go services, the uv image for `sink`; compose switched to the built images.
- **Done when:** the whole compose stack runs from the built images and the M1–M6 acceptance lists still pass; your
  measured image sizes are in the note.
- **You can explain:** why each service gets its own image; how configuration reaches each one.

### Day 69 — M7: Helm chart and deploy to kind
- **Objective:** Hookline running on Kubernetes.
- **Read:** ch 38, section 38.7.
- **Time:** 120 min (read 15 / build 95 / note 10).
- **Exercise:** Hookline M7, module `deploy`: a Helm chart with Deployments, Services, probes, resources and
  configuration for every service; dependencies (PostgreSQL, Redis, Redpanda) as section 38.7 specifies; migrations
  run as a Job before the services start; install on kind.
- **Done when:** all pods become ready on kind and an event ingested through a port-forward is counted by the sink.
- **You can explain:** why migrations run as a separate Job; how the relay's single-active lock behaves with two
  replicas.

### Day 70 — M7: Terraform and GitHub Actions for Hookline
- **Objective:** reproducible provisioning and CI for the whole project.
- **Read:** ch 38, section 38.7.
- **Time:** 115 min (read 15 / build 90 / note 10).
- **Exercise:** Hookline M7, module `deploy`: Terraform that creates the namespace and the Hookline release on kind
  through the helm provider; GitHub Actions workflows for tests (with service containers), image builds and a kind
  deploy smoke test.
- **Done when:** the M7 acceptance list in section 38.7 passes; one Terraform apply on a fresh kind cluster brings
  Hookline up; actionlint passes on all workflows.
- **You can explain:** what CI proves on every change and what it cannot; how you would roll back a bad Hookline
  release.

---

## Week 15 (days 71–75)

### Day 71 — Checkpoint 4: cloud-native delivery (lighter day)
- **Objective:** prove Part E: Hookline builds, deploys and rolls back reproducibly.
- **Read:** ch 38, section 38.7.
- **Time:** 90 min (read 15 / build 65 / note 10).
- **Exercise:** work through the review list starting from no cluster: (1) `doctor.sh` green; (2) all images build
  and run as non-root, scans and SBOMs saved; (3) Terraform apply on a fresh kind cluster brings Hookline up; (4) an
  event flows end to end to the sink; (5) a Helm upgrade and a rollback work; (6) the debugging drill faults can be
  found again; (7) actionlint passes; (8) answer the chapter 19–23 interview questions aloud.
- **Done when:** every item on the review list works; destroy leaves nothing behind.
- **You can explain:** the path from a commit to a running pod on kind, every step named.

### Day 72 — The three signals; OpenTelemetry in Go and Python
- **Objective:** instrument services with traces, metrics and logs.
- **Read:** ch 24, sections 24.1–24.3.
- **Time:** 110 min (read 40 / build 60 / note 10).
- **Exercise:** `ch24-telemetry`, part 1: the compose stack with the collector, Prometheus, Grafana, Loki and Tempo,
  datasources provisioned; OpenTelemetry in shorty and in the receiver.
- **Done when:** one request to shorty produces a trace visible in Tempo through Grafana, with spans from the
  handler.
- **You can explain:** what each signal answers; why the collector sits between services and backends.

### Day 73 — Prometheus, PromQL, RED and USE, Grafana
- **Objective:** metrics that answer questions, and a dashboard.
- **Read:** ch 24, sections 24.4–24.6.
- **Time:** 110 min (read 40 / build 60 / note 10).
- **Exercise:** `ch24-telemetry`, part 2: traffic from the load generator; PromQL for rate and a latency percentile,
  including the deliberate mistake without `le`; the three USE instruments of `linkworker`; the provisioned "RED
  overview" dashboard plus one panel of your own exported back to the file.
- **Done when:** the dashboard loads from the provisioned file and its panels move while traffic runs; the USE
  metrics test is green; your measured request rate and p99 are in the note.
- **You can explain:** counters versus histograms; RED for services versus USE for resources.

### Day 74 — Loki, Tempo, trace context through Kafka
- **Objective:** connect logs and traces, and carry a trace across a message.
- **Read:** ch 24, sections 24.7–24.9.
- **Time:** 115 min (read 40 / build 65 / note 10).
- **Exercise:** `ch24-telemetry`, part 3: inject trace context into a Kafka record header on produce and extract it
  on consume, with `kotel`; jump from a log line in Loki to its trace and back; run `tests/check_stack.sh`, then
  take the stack down.
- **Done when:** one trace in Tempo shows the producer, the consumer and the call to the receiver; the Redpanda test
  ran (not skipped) and passed; `tests/check_stack.sh` ends with `ALL CHECKS PASSED`.
- **You can explain:** how trace context travels in a message; why logs should carry the trace ID.

### Day 75 — SLIs, SLOs, error budgets, burn-rate alerts
- **Objective:** alert on what users feel, at the right speed.
- **Read:** ch 25, sections 25.1–25.3.
- **Time:** 110 min (read 45 / build 55 / note 10).
- **Exercise:** `ch25-slo`, part 1: recording rules for shorty's availability and latency SLIs; multi-window
  burn-rate alert rules; unit tests for the rules with promtool; the rules loaded into the chapter 24 stack.
- **Done when:** promtool rule tests pass, including a case where a fast burn fires and a short blip does not;
  Prometheus in the stack has loaded the 16 rules.
- **You can explain:** what an error budget is for; why multi-window burn-rate alerts beat a fixed error threshold.

---

## Week 16 (days 76–80)

### Day 76 — Runbooks, on-call, incident response, blameless postmortems (lighter day)
- **Objective:** the human half of operations.
- **Read:** ch 25, sections 25.4–25.7.
- **Time:** 85 min (read 40 / build 35 / note 10).
- **Exercise:** `ch25-slo`, part 2: a runbook for the burn-rate alert; run the staged incident of the chapter and
  write its postmortem from the template.
- **Done when:** the runbook says what to check first and how to mitigate; the staged incident fired the alert and
  it resolved after the rollback; the postmortem has a timeline, impact,
  contributing factors and action items, with no blame.
- **You can explain:** the incident roles; what makes a postmortem action item useful.

### Day 77 — M8: traces across API, Kafka and worker; metrics
- **Objective:** see one event's whole journey in one trace.
- **Read:** ch 38, section 38.8.
- **Time:** 120 min (read 15 / build 95 / note 10).
- **Exercise:** Hookline M8, modules `api`, `relay`, `dispatcher` and `worker`: OpenTelemetry in each; trace context
  stored with the outbox row and carried in the Kafka header so one trace runs from ingest to the delivery attempt;
  Prometheus metrics for ingest, delivery attempts by outcome, delivery latency and due-deliveries backlog.
- **Done when:** one trace shows ingest, relay publish, dispatch and delivery attempt; the metrics are scraped.
- **You can explain:** why the trace context is stored in the outbox row; which metric shows a growing backlog.

### Day 78 — M8: dashboards, SLOs and burn-rate alerts
- **Objective:** know Hookline's health at a glance and be alerted on user impact.
- **Read:** ch 38, section 38.8.
- **Time:** 110 min (read 15 / build 85 / note 10).
- **Exercise:** Hookline M8, module `deploy`: provisioned Grafana dashboards (ingest RED, delivery outcomes and
  latency, backlog, breaker states); SLOs for ingest availability and delivery latency from section 37.7, with
  recording rules, burn-rate alerts and promtool tests.
- **Done when:** dashboards load from files and move under traffic; promtool tests for all Hookline alert rules pass.
- **You can explain:** how you chose the SLIs; what a delivery SLO means when the receiver itself is down.

### Day 79 — M8: a deliberately broken release and its postmortem
- **Objective:** detect, mitigate and learn from an incident you caused on purpose.
- **Read:** ch 38, section 38.8.
- **Time:** 110 min (read 15 / build 75 / note 20).
- **Exercise:** Hookline M8: ship the broken release described in section 38.8 to kind while traffic runs; detect it
  from the dashboards and alerts; follow your runbook; roll back; write the postmortem.
- **Done when:** the M8 acceptance list in section 38.8 passes; the postmortem's timeline uses timestamps from your
  own dashboards and alerts.
- **You can explain:** how long detection took and why; which signal would have caught it sooner.

### Day 80 — Testing strategy, testcontainers, contract and property tests
- **Objective:** the right test at the right level.
- **Read:** ch 26, sections 26.1–26.4.
- **Time:** 110 min (read 35 / build 65 / note 10).
- **Exercise:** `ch26-testing`, part 1: integration tests for the chapter 10 repository with testcontainers-go and a
  test data builder; a JSON Schema contract for the webhook body, checked from Go and from Python; property tests
  for the backoff function and the signature with rapid.
- **Done when:** the integration tests start their own PostgreSQL, pass and leave no containers behind; the schema
  rejects every invalid example on both sides; the property tests pass with 2000 inputs per property.
- **You can explain:** what a contract test protects that an integration test does not; what a property test finds
  that table tests miss.

---

## Week 17 (days 81–85)

### Day 81 — Fuzzing, k6, reading load results
- **Objective:** find edge cases with fuzzing and measure under load.
- **Read:** ch 26, sections 26.5–26.7.
- **Time:** 110 min (read 35 / build 65 / note 10).
- **Exercise:** `ch26-testing`, part 2: a fuzz target for the signature-header parser, and a fix for the crash it
  finds; a constant-arrival-rate k6 script with thresholds against shorty; run it at three rates and read the
  results as the chapter teaches.
- **Done when:** the fuzz target runs for its time box and any crasher is fixed and kept as a test; the k6 summary
  from your run is saved with your reading of it.
- **You can explain:** why percentiles matter more than averages; what a k6 threshold turns a load test into.

### Day 82 — Code review by risk and static analysis (lighter day)
- **Objective:** review like a senior and let tools catch the rest.
- **Read:** ch 27, sections 27.1–27.6.
- **Time:** 90 min (read 35 / build 45 / note 10).
- **Exercise:** `ch27-review`: review the provided diff and write your comments; only then compare with the answer
  key in `solution/`; install golangci-lint (the exercise's step); run vet, staticcheck, golangci-lint, govulncheck,
  ruff and mypy on the seeded Go module and Python package and fix what they report.
- **Done when:** your review comments are saved before you open the key, with the missed issues listed afterwards;
  every linter passes on the seeded module and package, and the tests pass.
- **You can explain:** how you decide where to look first in a large diff; how govulncheck differs from a plain
  dependency scanner.

### Day 83 — OWASP API Top 10, OIDC, JWT checks, BOLA
- **Objective:** authenticate and authorize correctly.
- **Read:** ch 28, sections 28.1–28.4.
- **Time:** 110 min (read 45 / build 55 / note 10).
- **Exercise:** `ch28-security`, part 1: JWT validation middleware with tests for algorithm, issuer, audience,
  expiry and key id; a BOLA fix with a test.
- **Done when:** a test exists and passes for each JWT check rejecting a bad token; the BOLA test shows another
  tenant's object refused.
- **You can explain:** why the algorithm must be pinned; why BOLA tops the API list.

### Day 84 — SSRF, secrets, SBOM, threat modelling
- **Objective:** defend outbound calls and the supply chain; model threats.
- **Read:** ch 28, sections 28.5–28.8.
- **Time:** 110 min (read 40 / build 60 / note 10).
- **Exercise:** `ch28-security`, part 2: an SSRF-safe URL validator and a dialer that re-checks the resolved address;
  hashed API keys; secrets redacted from logs; an SBOM and a dependency scan; fill the STRIDE table template for the
  exercise service.
- **Done when:** tests reject private, loopback and link-local targets including a hostname that resolves to one;
  API keys are stored only as hashes and no secret reaches the log; the SBOM exists and the scan results are saved
  with their date; the STRIDE table is filled.
- **You can explain:** why validating the URL string alone is not enough against SSRF; why API keys are hashed.

### Day 85 — M9: integration, contract, property and load tests
- **Objective:** Hookline's test suite at senior level.
- **Read:** ch 38, section 38.9.
- **Time:** 120 min (read 15 / build 95 / note 10).
- **Exercise:** Hookline M9, modules `api`, `dispatcher` and `worker`: testcontainers integration tests for
  repositories and the ingest-to-delivery path; a contract test of the webhook payload shared with `sink`; property
  tests for backoff and signing; a k6 load test on ingest with thresholds.
- **Done when:** the integration and contract tests pass; property tests pass; the k6 run's summary from your
  machine is saved.
- **You can explain:** what the payload contract test protects receivers from; which Hookline behaviour you trust
  only because of an integration test.

---

## Week 18 (days 86–90)

### Day 86 — M9: API keys, SSRF guard, govulncheck and SBOM, STRIDE
- **Objective:** Hookline secured at the points an interviewer will probe.
- **Read:** ch 38, section 38.9.
- **Time:** 120 min (read 15 / build 85 / note 20).
- **Exercise:** Hookline M9, modules `api`, `worker` and `deploy`: hashed tenant API keys; the SSRF guard on endpoint
  URLs at registration and in the worker's dialer; govulncheck and SBOM generation in CI; the STRIDE threat model for
  Hookline.
- **Done when:** the M9 acceptance list in section 38.9 passes; registering an endpoint that points at a private
  address is refused, and so is a delivery to a hostname that resolves to one; the threat model names a mitigation
  or an accepted risk for every row.
- **You can explain:** why the SSRF check runs at registration and again at delivery; Hookline's top three threats.

### Day 87 — Checkpoint 5: operations, quality and security (lighter day)
- **Objective:** prove Parts F and G: Hookline is observable, tested and secured.
- **Read:** ch 38, section 38.9.
- **Time:** 90 min (read 15 / build 65 / note 10).
- **Exercise:** work through the review list on kind: (1) M1–M9 acceptance lists pass; (2) one trace spans ingest to
  delivery; (3) dashboards load and alert rules pass promtool; (4) the postmortem is complete; (5) integration,
  contract, property and fuzz tests pass; (6) the k6 thresholds hold on your machine or the gap is written down; (7)
  hashed keys, SSRF guard, govulncheck and SBOM are in place; (8) answer the chapter 24–28 interview questions aloud.
- **Done when:** every item on the review list works or has an honest written gap with a plan.
- **You can explain:** how you would know, and prove, that Hookline is meeting its SLOs today.

### Day 88 — System design method and design documents
- **Objective:** a repeatable method and a written design.
- **Read:** ch 29, sections 29.1–29.6.
- **Time:** 115 min (read 50 / build 55 / note 10).
- **Exercise:** `ch29-design-docs`: finish the sizing calculator script and its tests; write one design doc with
  the template (a Hookline extension of your choice), with one ADR.
- **Done when:** calculator tests pass and ruff and mypy are clean; the design doc has goals, non-goals, sizing, alternatives and risks, and
  the ADR records one decision with its consequences.
- **You can explain:** the six steps of the method; how a quick estimate changes a design.

### Day 89 — Evolving systems: expand and contract
- **Objective:** change a live schema without downtime.
- **Read:** ch 30, sections 30.1–30.5.
- **Time:** 115 min (read 40 / build 65 / note 10).
- **Exercise:** `ch30-expand-contract`: rename a column in four goose migrations while the Go writer and reader run;
  the test proves no failed request across the steps; the same rename for the `user.renamed` event.
- **Done when:** the rollout test passes with `-race` across all four steps with zero failed requests; the
  compatibility matrix and lock-queue tests pass; the event tests pass and the note says when `fullname` may be
  dropped from the event.
- **You can explain:** why expand and contract needs a deploy between steps; one use of a feature flag in the same
  change.

### Day 90 — Working as a senior
- **Objective:** ownership, estimation, mentoring and written communication.
- **Read:** ch 31, sections 31.1–31.6.
- **Time:** 90 min (read 40 / build 40 / note 10).
- **Exercise:** `ch31-senior-craft` (non-coding): an estimate with ranges for a Hookline feature; a written status
  update for it; a mentoring plan for a junior engineer joining the project.
- **Done when:** the estimate states ranges and assumptions; the status update fits on one screen and leads with
  status, date and risk; the mentoring plan has weekly goals.
- **You can explain:** why estimates are ranges; how you lead without authority.

---

## Week 19 (days 91–95)

### Day 91 — AI-assisted engineering: reviewing AI output
- **Objective:** delegate to AI tools and review their output like any other author's.
- **Read:** ch 32, sections 32.1–32.6.
- **Time:** 100 min (read 35 / build 55 / note 10).
- **Exercise:** `ch32-review-ai-output`: write tests that expose the four planted defects in the AI-written Go
  module, fix them, and record the review.
- **Done when:** each planted defect has a test that failed before the fix and passes after; the review note lists
  each defect and how it was found.
- **You can explain:** what you delegate to AI and what you never do; how AI use is treated in interviews.

### Day 92 — The Rust edge (lighter day)
- **Objective:** turn Rust experience into a clear interview asset.
- **Read:** ch 33, sections 33.1–33.5.
- **Time:** 85 min (read 35 / build 40 / note 10).
- **Exercise:** `ch33-rust-edge` (optional if Rust is not installed): port the chapter 16 token bucket to Rust with
  `cargo test`; write the one-page Rust-versus-Go comparison.
- **Done when:** the comparison page exists with at least one case for each language; the Rust tests pass if you did
  the port; the interview sentence is written.
- **You can explain:** when you would choose Rust over Go for an infrastructure component, and when not.

### Day 93 — The senior interview loop
- **Objective:** be ready for every stage of a senior loop in Israel.
- **Read:** ch 34, sections 34.1–34.7.
- **Time:** 110 min (read 45 / build 55 / note 10).
- **Exercise:** `ch34-interview-loop` (non-coding): five behavioural stories on the worksheet, one of them about
  Hookline; the CV checklist applied to your CV; the negotiation worksheet.
- **Done when:** each story has situation, task, action, result and what you would do differently, each told aloud
  in under three minutes; each of the five signals is covered by at least two stories; the CV passes the checklist;
  the negotiation sheet has a range with source and date and a written answer to the expectations question.
- **You can explain:** the stages of a senior loop; how you tell a story where things went wrong honestly.

### Day 94 — Coding drills: rate limiter, scheduler, TTL cache
- **Objective:** practical coding rounds, with tests as the specification.
- **Read:** ch 35, sections 35.1–35.4.
- **Time:** 120 min (read 25 / build 85 / note 10).
- **Exercise:** `ch35-drills`, part 1: the rate limiter, the task scheduler (pause, resume, retry, cancel) and the TTL
  cache, each in the time box the chapter gives, talking aloud as in an interview.
- **Done when:** the given tests pass with `-race` for all three; the note records your time per drill and where you stalled.
- **You can explain:** the trade-offs you named in each drill; how you test concurrency in an interview.

### Day 95 — Coding drills: retry-safe webhook handler, booking, review drill
- **Objective:** the drills closest to Hookline, plus a review round.
- **Read:** ch 35, sections 35.5–35.7.
- **Time:** 115 min (read 20 / build 85 / note 10).
- **Exercise:** `ch35-drills`, part 2: the retry-safe webhook handler, booking without double booking, and the code
  review drill.
- **Done when:** the given tests pass with `-race` for both drills, the double-booking test included against
  PostgreSQL; your review is committed before you open the key, and the comparison with the key is recorded.
- **You can explain:** two ways to prevent double booking and their costs.

---

## Week 20 (days 96–100)

### Day 96 — System design interviews: mock 1 (Hookline and rate limiting)
- **Objective:** design under interview conditions.
- **Read:** ch 36, sections 36.1–36.4.
- **Time:** 115 min (read 40 / build 65 / note 10).
- **Exercise:** `ch36-question-bank`, part 1: a timed mock design of the notification-system case aloud on a
  whiteboard or paper (alone or with a friend), then compare with the chapter's case; quick-pass the Hookline and
  rate-limiting cases.
- **Done when:** the mock is self-scored against the chapter's method with three concrete fixes written down.
- **You can explain:** how you spend the minutes of a design round; Hookline's design in five minutes.

### Day 97 — System design interviews: mock 2 and the question bank
- **Objective:** a second mock and the full question bank.
- **Read:** ch 36, sections 36.5–36.8 (36.8: its intro only; the bank is used through the quiz).
- **Time:** 120 min (read 25 / build 85 / note 10).
- **Exercise:** `ch36-question-bank`, part 2: a timed mock of a case you have not practised (job scheduler, ticket
  booking or metrics ingestion); render the question bank, run two 10-question self-quiz rounds and list the weak topics.
- **Done when:** the second mock is self-scored with two fixes written; the self-quiz is done and the weak topics are
  listed.
- **You can explain:** for each weak topic, the alternative, its cost and your evidence.

### Day 98 — M10: evaluation under load and chaos
- **Objective:** prove Hookline's SLOs and delivery guarantee with your own measurements.
- **Read:** ch 38, section 38.10; ch 39, sections 39.1–39.4.
- **Time:** 120 min (read 25 / build 85 / note 10).
- **Exercise:** Hookline M10, all modules on kind: run the k6 load test and read the SLOs from your dashboards; chaos
  runs that kill workers and the relay under load and set the sink to fail; the delivery-guarantee proof that every
  accepted event reached the sink at least once, from the sink's counts against the database; estimate the running
  cost as section 39.4 describes.
- **Done when:** the evaluation note holds your measured SLO results, each chaos run's outcome and the sink-count
  reconciliation, with any gap stated plainly.
- **You can explain:** what your sink counts prove and what they do not; which SLO was hardest to hold and why.

### Day 99 — M10: demo script and publishing
- **Objective:** Hookline ready to show and to defend.
- **Read:** ch 38, section 38.10; ch 39, sections 39.5–39.7.
- **Time:** 120 min (read 25 / build 85 / note 10).
- **Exercise:** Hookline M10: write and rehearse the demo script; finish the README with architecture, how to run it
  and your evaluation results; audit for secrets and anything that is not yours; tag a release and publish the
  repository with CI green; answer the twenty project questions aloud.
- **Done when:** the M10 acceptance list in section 38.10 passes; a public repository link exists; the demo runs
  start to finish within the length the script sets.
- **You can explain:** Hookline in three minutes: problem, design, evidence, limits.

### Day 100 — Checkpoint 6: the capstone, checklists and retrospective (lighter day)
- **Objective:** prove the whole course and plan what comes next.
- **Read:** ch 39, section 39.8.
- **Time:** 90 min (read 15 / build 50 / note 25).
- **Exercise:** work through the review list: (1) the published repository builds and passes CI; (2) a fresh
  Terraform apply on kind brings Hookline up and the demo runs; (3) the evaluation note's results are reproducible;
  (4) the M1–M10 acceptance lists pass; (5) fill the appendix E learning checklist and interview-readiness checklist
  honestly; then write the retrospective (built, measured, would do differently) and a job-search plan.
- **Done when:** every item on the review list works, both appendix E checklists are filled with three weak areas
  named and scheduled for revisits, and the retrospective exists.
- **You can explain:** what you can now do that you could not on day 1, with evidence for each claim.

---

## Checkpoints at a glance

| Checkpoint | Day | What it proves |
| --- | --- | --- |
| 1 | 14 | You can write, test, profile and gracefully operate a production Go HTTP service (Part A) |
| 2 | 45 | You can build an idempotent, durable event pipeline with PostgreSQL, Redis and Redpanda, and deliver signed webhooks with retries (Part C, M1–M5) |
| 3 | 56 | You can isolate failures and bound load with breakers, bulkheads, rate limits, shedding and a single active relay (Part D, M6) |
| 4 | 71 | You can build, deploy, provision and roll back a multi-service system on Kubernetes with Helm, Terraform and CI (Part E, M7) |
| 5 | 87 | You can observe, test and secure the system: traces, SLO alerts, postmortem, layered tests, SSRF and key hygiene (Parts F–G, M8–M9) |
| 6 | 100 | A published, evaluated capstone with measured SLOs, chaos results and a delivery-guarantee proof, plus an honest readiness assessment |
