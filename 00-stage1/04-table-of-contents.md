# Table of contents (Stage 1b) — the contract for chapters, schedule and checker

Written 2026-10-05. Every chapter file uses exactly these numbered sections (`## N.k <Hebrew title>`); the schedules
reference them as "פרק N, סעיפים N.a עד N.b" and `tools/consistency_check.py` verifies every reference. A writer may
add `###` subsections freely but may not add, drop or renumber `##` sections without updating this file and both
schedules in the same change. "Days" are study days (Day 0 = installations). "Capstone reuses" names what Hookline
takes from the chapter's exercise.

Exercise slugs: `exercises/chNN-<slug>/`. Go exercises are one module per package (`go.mod` with exact pins);
Python exercises are uv projects (`pyproject.toml` + `uv.lock`).

---

## Part 0 — לפני שמתחילים

### 00 `00-how-to-study.md` — פרק 0: איך ללמוד מהספר הזה
Objective: the method, the daily rhythm, Obsidian, honesty, cost, what to do when stuck; Day 0 and the WSL2/Mac note.
Sections: 0.1 למי הספר ומה יוצא ממנו · 0.2 מבנה הספר · 0.3 יום לימוד טיפוסי · 0.4 יום 0, WSL2 ו־Mac · 0.5 Obsidian
והפתקים · 0.6 תרגילים, פתרונות ובדיקות · 0.7 כלל הכנות · 0.8 עלות · 0.9 כשנתקעים · 0.10 נקודות ביקורת ומעקב.
Exercise: none (chapter 0 has no apparatus). Days: 1.

### 01 `01-senior-in-2026.md` — פרק 1: מה זה „בכירה” ב־2026
Objective: what senior means in 2026 Israeli postings (cites `01-market-research.md`), the lanes, how to read a
posting, her Rust/C++ edge, the senior loop in one page, the capstone in one page.
Sections: 1.1 מה המודעות אומרות · 1.2 המסלולים: backend, platform, SRE, תשתיות נתונים · 1.3 איך קוראים מודעה ·
1.4 מה מבדיל בכירה: היקף, בעלות, שיקול דעת · 1.5 היתרון שלך: Rust ו־C++ · 1.6 תהליך הגיוס הבכיר בעמוד אחד ·
1.7 פרויקט הגמר בעמוד אחד · 1.8 איך הספר ממפה למודעות.
Exercise `ch01-reading-postings`: save five real postings as text, run a small Python tally script (given) and fill a
skills-gap matrix. Days: 2. Capstone reuses: nothing (the gap matrix feeds chapter 34).

### 02 `02-workbench-wsl2.md` — פרק 2: סביבת העבודה: WSL2, Docker, Go ו־uv
Objective: a working, verified environment; the install-chapter pattern (what, why, where, version, commands,
expected output, verify, restart, remove, troubleshooting) for every tool.
Sections: 2.1 מה מתקינים ולמה · 2.2 WSL2 ו־Ubuntu · 2.3 הטרמינל ו־bash · 2.4 Docker Desktop עם WSL2 · 2.5 Git
ומפתחות SSH · 2.6 VS Code עם הרחבת WSL · 2.7 Go · 2.8 uv ו־Python · 2.9 מאגר הקורס ובדיקת עשן · 2.10 התקנות של חלק ה:
kind, kubectl, Helm, Terraform, k6 · 2.11 פתרון תקלות · 2.12 עצירה, הפעלה מחדש וניקוי · 2.13 אם את עובדת על Mac.
Exercise `ch02-workbench`: `doctor.sh` (checks every tool and version), a hello-world Go module with one test, a
pytest smoke test, `docker run hello-world`, first commit. Days: 0 (2.1–2.8, and 2.13 on a Mac), 3 (2.9, 2.11–2.12),
57 (2.10). Capstone reuses: `doctor.sh`.

---

## Part A — Go

### 03 `03-go-first-steps.md` — פרק 3: צעדים ראשונים ב־Go
Objective: write, test and lint small Go programs; understand values, pointers, GC versus ownership, slices and maps.
Sections: 3.1 למה Go, ומה היא לא · 3.2 שרשרת הכלים: run, build, fmt, vet · 3.3 מודולים וחבילות · 3.4 התחביר
למתכנתת Rust · 3.5 ערכים, מצביעים, ו־GC מול בעלות · 3.6 slices מבפנים · 3.7 maps, מחרוזות ו־runes · 3.8 בדיקות
טבלאיות · 3.9 תוכנית קטנה: `logstat` · 3.10 staticcheck ו־gofmt כהרגל.
Exercise `ch03-logstat`: CLI that parses access-log lines (stdin or file), counts status classes and top paths,
handles malformed lines, table-driven tests. Days: 4 (3.1–3.5), 5 (3.6–3.10).

### 04 `04-go-types-interfaces-errors.md` — פרק 4: טיפוסים, ממשקים ושגיאות
Objective: structs and methods, implicit interfaces, errors as values, wrapping, generics, composition.
Sections: 4.1 structs, מתודות ו־receivers · 4.2 ממשקים מול traits · 4.3 ממשקים קטנים: `io.Reader` ו־`io.Writer` ·
4.4 שגיאות כערכים מול `Result` · 4.5 עטיפה: `%w`, `errors.Is` ו־`errors.As` · 4.6 panic ו־recover, ולמה כמעט אף פעם ·
4.7 generics · 4.8 הטמעה (embedding) והרכבה · 4.9 תוכנית קטנה: `ledger`.
Exercise `ch04-ledger`: an account ledger library + CLI: `Store` interface with memory and JSON-file implementations,
sentinel and typed errors, a generic `Page[T]`, tests. Days: 6 (4.1–4.5), 7 (4.6–4.9).

### 05 `05-go-concurrency.md` — פרק 5: מקביליות ב־Go
Objective: goroutines, channels, select, context, sync, errgroup, race detection, leak avoidance.
Sections: 5.1 goroutines מול async/await ו־threads · 5.2 channels · 5.3 select וזמנים קצובים · 5.4 context: ביטול
ומועדים · 5.5 sync: Mutex, WaitGroup, Once · 5.6 errgroup · 5.7 דפוסים: worker pool, pipeline, fan-out ו־fan-in ·
5.8 מרוצי נתונים ו־`-race` · 5.9 דליפות goroutines · 5.10 תוכנית קטנה: `linkcheck`.
Exercise `ch05-linkcheck`: bounded-concurrency URL checker with per-request timeouts, global cancellation, ordered
report; tested with `httptest` and `-race`. Days: 8 (5.1–5.4), 9 (5.5–5.7), 10 (5.8–5.10, lighter day).

### 06 `06-go-http-service.md` — פרק 6: שירות HTTP לייצור
Objective: ship a tested HTTP service with config, structured logs, graceful shutdown, a profile and a benchmark.
Sections: 6.1 `net/http` והנתב של Go · 6.2 handlers, middleware ו־JSON · 6.3 תצורה ממשתני סביבה · 6.4 `slog`: לוגים
מובנים · 6.5 כיבוי מסודר · 6.6 בדיקות עם `httptest` · 6.7 מבנה פרויקט ו־Makefile · 6.8 pprof: פרופיל ראשון · 6.9
בנצ'מרקים · 6.10 נקודת ביקורת 1.
Exercise `ch06-shorty`: URL-shortener service (create, resolve, stats) with in-memory store behind an interface,
env config, request-ID + logging + recovery middleware, graceful shutdown, `/debug/pprof`, tests, one benchmark.
Days: 11 (6.1–6.3), 12 (6.4–6.6), 13 (6.7–6.9), 14 (6.10, checkpoint 1, lighter day).
Capstone reuses: the server skeleton, middleware, config and shutdown (M1).

---

## Part B — Python בכירה

### 07 `07-python-typed-packaged.md` — פרק 7: Python עם טיפוסים, אריזה ובדיקות
Objective: senior-level Python project hygiene.
Sections: 7.1 מה „Python בכירה” אומרת · 7.2 פרויקט עם uv: pyproject, lock ו־scripts · 7.3 ruff ו־mypy · 7.4 טיפוסים
מתקדמים: Protocol, TypedDict ו־generics · 7.5 pydantic v2 · 7.6 pytest: fixtures, parametrize ו־monkeypatch · 7.7 מבנה
חבילה ו־entry points · 7.8 ביצועים: מודדים לפני שמשפרים.
Exercise `ch07-eventcheck`: a typed, packaged CLI that validates a JSONL file of webhook events with pydantic, reports
errors with line numbers, ruff+mypy clean, parametrized tests, one cProfile measurement. Days: 15 (7.1–7.4), 16 (7.5–7.8).

### 08 `08-python-async-fastapi.md` — פרק 8: Python אסינכרונית ו־FastAPI
Objective: asyncio done right, FastAPI service with lifespan, async tests, performance limits.
Sections: 8.1 asyncio: לולאת האירועים · 8.2 משימות, TaskGroup וביטול · 8.3 זמנים קצובים ו־Semaphore · 8.4 FastAPI:
נתיבים, מודלים ותלויות · 8.5 lifespan ומשאבים · 8.6 בדיקות אסינכרוניות עם httpx · 8.7 ביצועים: GIL, עבודה חוסמת
ו־workers · 8.8 תוכנית: מקבל webhooks.
Exercise `ch08-receiver`: FastAPI webhook receiver: verifies an HMAC-SHA256 signature, dedupes by event id, records
deliveries, has configurable failure rate and latency; async tests. Days: 17 (8.1–8.3), 18 (8.4–8.8).
Capstone reuses: becomes `hookline/sink` (the receiver simulator, M3).

---

## Part C — יסודות backend

### 09 `09-api-design-grpc.md` — פרק 9: עיצוב API ו־gRPC
Objective: contract-first HTTP APIs and when gRPC fits.
Sections: 9.1 REST כחוזה: משאבים ופעלים · 9.2 OpenAPI כמקור אמת · 9.3 חוזה שגיאות: problem details · 9.4 עימוד:
offset מול cursor · 9.5 מפתחות אידמפוטנטיות · 9.6 גרסאות ותאימות לאחור · 9.7 gRPC ו־protobuf · 9.8 מתי לא gRPC ·
9.9 אימות קלט וגבולות.
Exercise `ch09-catalog-api`: Go catalog API from an OpenAPI 3.1 spec: problem-details errors, cursor pagination,
idempotency-key middleware (in-memory), request size limits; plus a tiny gRPC `Ping/GetItem` service generated with
buf. Days: 19 (9.1–9.3), 20 (9.4–9.6), 21 (9.7–9.9). Capstone reuses: error contract, pagination, idempotency (M1–M2).

### 10 `10-postgres-queries-indexes.md` — פרק 10: PostgreSQL: שאילתות ואינדקסים
Objective: read plans on realistic data and choose indexes; use pgx from Go.
Sections: 10.1 PostgreSQL בקונטיינר ו־psql · 10.2 סכמה ונתונים מציאותיים · 10.3 EXPLAIN ו־EXPLAIN ANALYZE ·
10.4 אינדקסי B-tree · 10.5 אינדקסים מרוכבים, חלקיים ומכסים · 10.6 N+1 ועימוד מהיר · 10.7 pgx מ־Go.
Exercise `ch10-orders-db`: compose file, a seed script generating a realistic orders dataset with `generate_series`,
five slow queries to fix with indexes (the reader records her own before/after timings), a pgx repository with tests.
Days: 24 (10.1–10.4), 25 (10.5–10.7).

### 11 `11-postgres-transactions-migrations.md` — פרק 11: PostgreSQL: טרנזקציות, נעילות ומיגרציות
Objective: isolation levels with demonstrated anomalies, row locks, SKIP LOCKED, deadlocks, migrations, pooling.
Sections: 11.1 טרנזקציות ו־ACID בפועל · 11.2 רמות בידוד · 11.3 אנומליות מודגמות: עדכון אבוד ו־write skew · 11.4 נעילות
שורה: `FOR UPDATE` ו־`SKIP LOCKED` · 11.5 קיפאון (deadlock) · 11.6 מיגרציות עם goose · 11.7 מאגר חיבורים · 11.8
ניסיון חוזר על כשל סריאליזציה.
Exercise `ch11-isolation-lab`: Go tests that open two connections and reproduce lost update and write skew, then fix
them (`FOR UPDATE`, `SERIALIZABLE` + retry); a `SKIP LOCKED` work queue; goose migrations. Days: 26 (11.1–11.3), 27
(11.4–11.5), 28 (11.6–11.8). Capstone reuses: migrations, retry-on-serialization helper, SKIP LOCKED (M2, M5).

### 12 `12-redis-caching.md` — פרק 12: Redis ומטמון
Objective: caching patterns, invalidation, stampedes.
Sections: 12.1 Redis בקונטיינר ומבני הנתונים · 12.2 cache-aside · 12.3 TTL, פינוי ואי־תוקף · 12.4 התפרצות מטמון
ו־singleflight · 12.5 מונים ופעולות אטומיות · 12.6 מתי לא לשים מטמון.
Exercise `ch12-cache-aside`: cache-aside over the chapter 10 repository with go-redis; a stampede demo (N concurrent
misses counted) and the singleflight fix; tests against a real Redis container (skipped with a message if absent).
Days: 31 (12.1–12.3), 32 (12.4–12.6). Capstone reuses: subscription cache (M3).

### 13 `13-kafka-redpanda.md` — פרק 13: Kafka עם Redpanda
Objective: Kafka semantics in practice, the outbox pattern, and why exactly-once is a contract.
Sections: 13.1 יומן מול תור · 13.2 Redpanda בקונטיינר ו־rpk · 13.3 topics ומחיצות · 13.4 producers: מפתחות, סדר ו־acks ·
13.5 consumer groups ו־offsets · 13.6 rebalance ועיבוד „לפחות פעם אחת” · 13.7 תבנית ה־outbox · 13.8 למה „בדיוק פעם אחת”
הוא חוזה · 13.9 franz-go מ־Go.
Exercise `ch13-outbox`: a Go producer/consumer pair with franz-go; an outbox table written in the same transaction as
the business row and a relay that publishes and marks rows; a consumer that dedupes; a crash-in-the-middle test.
Days: 36 (13.1–13.3), 37 (13.4–13.6), 38 (13.7–13.9). Capstone reuses: outbox relay and consumer (M4).

### 14 `14-background-jobs.md` — פרק 14: עבודות רקע וניסיונות חוזרים
Objective: durable jobs, scheduled retries with backoff and jitter, consumer idempotency, dead letters, schedulers.
Sections: 14.1 עבודות רקע: מתי ולמה · 14.2 תור עבודות על PostgreSQL · 14.3 ניסיון חוזר, המתנה גדלה ו־jitter · 14.4
אידמפוטנטיות אצל הצרכן · 14.5 תור מתים והפעלה חוזרת · 14.6 מתזמנים ו־cron · 14.7 כיבוי מסודר של workers.
Exercise `ch14-jobqueue`: a Postgres-backed job queue (`SKIP LOCKED`, `run_at`, attempts, dead state, replay), a
backoff function with tests, idempotent handler, graceful drain. Days: 41 (14.1–14.4), 42 (14.5–14.7).
Capstone reuses: backoff, attempt bookkeeping, dead/replay (M5).

---

## Part D — מערכות מבוזרות

### 15 `15-timeouts-retries-breakers.md` — פרק 15: זמנים קצובים, ניסיונות חוזרים ומפסקי זרם
Sections: 15.1 כשל חלקי · 15.2 זמנים קצובים ותקציב זמן לאורך השרשרת · 15.3 ניסיון חוזר עם jitter ותקציב ניסיונות ·
15.4 סערת ניסיונות חוזרים · 15.5 מפסק זרם · 15.6 מחיצות אטומות (bulkheads) · 15.7 הידרדרות מכובדת.
Exercise `ch15-resilience-lab`: a flaky upstream (configurable latency/errors) and a client; implement deadline
propagation, a retry budget and a circuit breaker; a lab script prints call counts with and without each mechanism.
Days: 46 (15.1–15.4), 47 (15.5–15.7). Capstone reuses: breaker per endpoint, bulkhead (M6).

### 16 `16-overload-rate-limiting.md` — פרק 16: עומס: הגבלת קצב, לחץ חוזר והשלת עומס
Sections: 16.1 מה קורה לשירות בעומס יתר · 16.2 token bucket ב־Go · 16.3 הגבלת קצב מבוזרת ב־Redis · 16.4 לחץ חוזר ·
16.5 השלת עומס ועדיפויות · 16.6 תורים חסומים ומגבלות מקביליות.
Exercise `ch16-ratelimit`: token bucket with an injected clock (tests run instantly), a Redis Lua limiter, a
load-shedding middleware with a concurrency limit; a small load script. Days: 48 (16.1–16.3), 49 (16.4–16.6).
Capstone reuses: per-tenant limiter, ingest shedding (M6).

### 17 `17-consistency-replication-partitioning.md` — פרק 17: עקביות, שכפול וחלוקה למחיצות
Sections: 17.1 מודלי עקביות במילים פשוטות · 17.2 שכפול ב־PostgreSQL · 17.3 פיגור שכפול, מודגם · 17.4 לקרוא את מה
שכתבת · 17.5 חלוקה למחיצות ובחירת מפתח · 17.6 מחיצות חמות · 17.7 CAP ו־PACELC בלי מיתוסים.
Exercise `ch17-replica-lag`: compose with a primary and a streaming replica; a Go probe that writes and reads back
through the replica and reports observed lag; a partition-key skew simulator. Days: 50 (17.1–17.4), 51 (17.5–17.7).

### 18 `18-locks-leaders-sagas.md` — פרק 18: נעילות, בחירת מנהיג וסאגות
Sections: 18.1 למה נעילה מבוזרת קשה · 18.2 advisory locks ב־PostgreSQL · 18.3 נעילות ב־Redis וגבולותיהן · 18.4 fencing
tokens · 18.5 בחירת מנהיג · 18.6 קונצנזוס: הרעיון בלי Raft · 18.7 סאגות ופעולות מפצות.
Exercise `ch18-leader`: leader election with a Postgres session advisory lock (two processes, kill one, the other takes
over), fencing token check, an orchestrated saga with compensation and tests. Days: 52 (18.1–18.4), 53 (18.5–18.7).
Capstone reuses: single active outbox relay (M6).

---

## Part E — Cloud-native

### 19 `19-docker-done-right.md` — פרק 19: Docker כמו שצריך
Sections: 19.1 מה זה image: שכבות ומטמון · 19.2 Dockerfile רב־שלבי ל־Go · 19.3 distroless ומשתמש שאינו root · 19.4
image ל־Python עם uv · 19.5 compose לפיתוח · 19.6 סריקה ו־SBOM · 19.7 תגיות וגרסאות.
Exercise `ch19-images`: images for the chapter 6 service and the chapter 8 receiver; compare sizes (measured by the
reader); run as non-root; Trivy/Grype scan and Syft SBOM. Days: 58 (19.1–19.4), 59 (19.5–19.7).

### 20 `20-kubernetes-on-kind.md` — פרק 20: Kubernetes על kind
Sections: 20.1 מה Kubernetes פותר · 20.2 kind: אשכול על המחשב שלך · 20.3 pods · 20.4 deployments ו־ReplicaSets · 20.5
services ו־DNS · 20.6 ConfigMaps ו־Secrets · 20.7 probes.
Exercise `ch20-k8s-basics`: manifests for shorty (chapter 6): Deployment, Service, ConfigMap, Secret, probes; load the
image into kind; port-forward and call it. Days: 60 (20.1–20.4), 61 (20.5–20.7).

### 21 `21-kubernetes-operations-helm.md` — פרק 21: הפעלת Kubernetes ו־Helm
Sections: 21.1 requests ו־limits · 21.2 HPA · 21.3 rollouts ו־rollbacks · 21.4 ingress · 21.5 דיבוג: describe, logs,
events ו־exec · 21.6 Helm: chart ראשון · 21.7 values ו־templates.
Exercise `ch21-ops-helm`: resources, metrics-server and an HPA, a broken rollout and `rollout undo`, an ingress on kind,
a debugging drill (four planted faults), and a Helm chart for shorty with `helm lint` and `helm template`.
Days: 62 (21.1–21.3), 63 (21.4–21.7).

### 22 `22-terraform.md` — פרק 22: Terraform
Sections: 22.1 תשתית כקוד · 22.2 providers, resources ו־state · 22.3 plan ו־apply מול kind · 22.4 משתנים, outputs
ומודולים · 22.5 state מרוחק ונעילות · 22.6 ה־provider של Helm · 22.7 אופציונלי: AWS free tier.
Exercise `ch22-terraform`: Terraform config with the kubernetes and helm providers against kind: a namespace, a
ConfigMap, the shorty chart release; a module; `plan`/`apply`/`destroy`. The AWS section is optional and unexecuted
by the author. Days: 64 (22.1–22.3), 65 (22.4–22.7).

### 23 `23-github-actions.md` — פרק 23: GitHub Actions
Sections: 23.1 workflow, jobs ו־steps · 23.2 בנייה ובדיקה ל־Go ול־Python · 23.3 service containers · 23.4 בניית image
ופרסום ל־GHCR · 23.5 פריסה ל־kind ב־CI · 23.6 environments, סודות והרשאות · 23.7 releases · 23.8 actionlint.
Exercise `ch23-ci`: workflows for the chapter 6 and chapter 11 exercises (tests with a Postgres service container),
image build, kind deploy smoke test, release on tag; validated with actionlint here, run in her own repository later.
Days: 66 (23.1–23.4), 67 (23.5–23.8). Capstone reuses: the CI workflows (M7).

---

## Part F — תצפיתיות ותפעול

### 24 `24-telemetry.md` — פרק 24: טלמטריה: OpenTelemetry, Prometheus, Grafana ו־Loki
Sections: 24.1 שלושת האותות · 24.2 OpenTelemetry ב־Go · 24.3 OpenTelemetry ב־Python · 24.4 Prometheus: מדדים
ו־PromQL · 24.5 RED ו־USE · 24.6 Grafana · 24.7 Loki ולוגים · 24.8 Tempo ומעקבים · 24.9 הפצת הקשר דרך Kafka.
Exercise `ch24-telemetry`: compose stack (collector, Prometheus, Grafana, Loki, Tempo) with provisioned datasources;
shorty and the receiver instrumented; a RED dashboard JSON; trace context through a Kafka message.
Days: 72 (24.1–24.3), 73 (24.4–24.6), 74 (24.7–24.9). Capstone reuses: telemetry setup and dashboards (M8).

### 25 `25-slo-alerting-incidents.md` — פרק 25: SLO, התראות, תורנות ותחקיר
Sections: 25.1 SLI ו־SLO · 25.2 תקציב שגיאות · 25.3 התראות על קצב שריפה · 25.4 runbooks · 25.5 תורנות · 25.6 תגובה
לתקרית · 25.7 תחקיר ללא האשמה.
Exercise `ch25-slo`: recording and multi-window burn-rate alert rules for shorty with `promtool test rules`, a runbook,
and a postmortem for a staged incident. Days: 75 (25.1–25.3), 76 (25.4–25.7). Capstone reuses: SLO rules (M8).

---

## Part G — איכות ואבטחה

### 26 `26-testing-strategy-load.md` — פרק 26: אסטרטגיית בדיקות ובדיקות עומס
Sections: 26.1 מה לבדוק ובאיזו רמה · 26.2 testcontainers-go · 26.3 בדיקות חוזה · 26.4 בדיקות תכונה עם rapid · 26.5
fuzzing ב־Go · 26.6 k6: בדיקת עומס ראשונה · 26.7 לקרוא תוצאות עומס.
Exercise `ch26-testing`: integration tests for the chapter 10 repository with testcontainers, a JSON-schema contract
test, property tests for the backoff function, a fuzz target, a k6 script with thresholds against shorty.
Days: 80 (26.1–26.4), 81 (26.5–26.7). Capstone reuses: test patterns (M9).

### 27 `27-code-review-static-analysis.md` — פרק 27: סקירת קוד וניתוח סטטי
Sections: 27.1 סקירה לפי סיכון · 27.2 מה מחפשים · 27.3 לכתוב הערות טובות · 27.4 לקבל סקירה · 27.5 ניתוח סטטי:
vet, staticcheck, golangci-lint, ruff ו־mypy · 27.6 govulncheck.
Exercise `ch27-review`: review a provided diff with planted issues (answer key in `solution/`), then run the linters on a
seeded module and fix what they report. Days: 82 (27.1–27.6).

### 28 `28-security-backend.md` — פרק 28: אבטחה לשירותי backend
Sections: 28.1 OWASP API Security Top 10 · 28.2 אימות ו־OIDC · 28.3 JWT: מה בודקים · 28.4 הרשאות: BOLA · 28.5 SSRF ·
28.6 ניהול סודות · 28.7 SBOM וסריקת תלויות · 28.8 מודל איומים.
Exercise `ch28-security`: JWT validation middleware with tests (alg, iss, aud, exp, kid), a BOLA fix with a test, an
SSRF-safe URL validator and dialer, hashed API keys; a STRIDE table template. Days: 83 (28.1–28.4), 84 (28.5–28.8).
Capstone reuses: API-key hashing, SSRF guard, threat model (M9).

---

## Part H — אומנות הבכירות

### 29 `29-system-design-method.md` — פרק 29: שיטת תכנון ומסמכי תכנון
Sections: 29.1 שיטה בשישה צעדים · 29.2 הערכות גודל מהירות · 29.3 מסמך תכנון · 29.4 ADR · 29.5 מסמך לדוגמה: הפעלה
חוזרת ב־Hookline · 29.6 מסמך לדוגמה: מעבר לשני אזורים.
Exercise `ch29-design-docs`: templates plus a sizing calculator script (Python) with tests; write one design doc.
Days: 88.

### 30 `30-evolving-systems.md` — פרק 30: מערכות שמשתנות
Sections: 30.1 מיגרציה ללא השבתה: expand ו־contract · 30.2 תאימות לאחור ולפנים · 30.3 דגלי תכונה · 30.4 הגירת מערכת:
strangler fig · 30.5 חוב טכני: לזהות, לתמחר, לשלם.
Exercise `ch30-expand-contract`: rename a column with zero downtime in four goose migrations while a Go writer and
reader run; a test proves no failed request across the steps. Days: 89.

### 31 `31-working-as-senior.md` — פרק 31: לעבוד כבכירה
Sections: 31.1 בעלות · 31.2 הערכת זמנים · 31.3 חניכה · 31.4 תקשורת בכתב · 31.5 הובלה בלי סמכות · 31.6 עבודה עם
מנהלת ועם מוצר.
Exercise `ch31-senior-craft` (non-coding): an estimate with ranges, a written status update, a mentoring plan.
Days: 90.

### 32 `32-ai-assisted-engineering.md` — פרק 32: הנדסה בעזרת AI
Sections: 32.1 מה הכלים עושים ב־2026 · 32.2 האצלה: מה כן ומה לא · 32.3 לסקור פלט של AI · 32.4 כשהוא טועה · 32.5 AI
בראיונות · 32.6 מה בספר הראשון.
Exercise `ch32-review-ai-output`: a plausible AI-written Go module with four planted defects; tests that expose them;
fix and record the review. Days: 91.

### 33 `33-rust-edge.md` — פרק 33: היתרון של Rust
Sections: 33.1 איפה יש משרות Rust · 33.2 Rust מול Go: שיקול דעת · 33.3 להציג את הניסיון · 33.4 תרגום דפוסים · 33.5
Rust בראיון.
Exercise `ch33-rust-edge`: port the chapter 16 token bucket to Rust (`cargo test`) and write a one-page Rust-vs-Go
comparison; optional if Rust is not installed. Days: 92.

---

## Part I — קריירה

### 34 `34-senior-interview-loop.md` — פרק 34: תהליך הגיוס הבכיר
Sections: 34.1 מבנה התהליך ב־2026 · 34.2 סיפורים התנהגותיים · 34.3 חמישה סיפורים שלך · 34.4 מטלת בית · 34.5 קורות
חיים ו־LinkedIn · 34.6 משא ומתן בישראל · 34.7 כלל הכנות.
Exercise `ch34-interview-loop` (non-coding): story worksheet, CV checklist, negotiation worksheet. Days: 93.

### 35 `35-coding-drills.md` — פרק 35: תרגילי קוד מעשיים
Sections: 35.1 איך נראה ראיון קוד מעשי · 35.2 מגביל קצב · 35.3 מתזמן משימות · 35.4 מטמון עם TTL · 35.5 מקבל
webhooks שבטוח לניסיונות חוזרים · 35.6 הזמנה בלי הזמנה כפולה · 35.7 תרגיל סקירת קוד.
Exercise `ch35-drills`: five Go drills (rate limiter; task scheduler with pause, resume, retry, cancel; TTL cache;
retry-safe webhook handler; booking without double booking) with tests as the specification, plus a review drill.
Days: 94 (35.1–35.4), 95 (35.5–35.7).

### 36 `36-system-design-interviews.md` — פרק 36: תכנון מערכות בראיון ומאגר השאלות
Sections: 36.1 הלוח והזמן · 36.2 מקרה: Hookline · 36.3 מקרה: שירות הגבלת קצב · 36.4 מקרה: מערכת התראות · 36.5 מקרה:
מתזמן עבודות · 36.6 מקרה: הזמנת כרטיסים · 36.7 מקרה: קליטת מדדים · 36.8 מאגר השאלות.
Exercise `ch36-question-bank`: `data/questions.jsonl` (id, category, chapter, question, answer, tradeoff, follow_up)
rendered into §36.8 by `tools/render_question_bank.py`; a self-quiz script. Days: 96 (36.1–36.4), 97 (36.5–36.8).

---

## Part J — פרויקט הגמר

### 37 `37-hookline-spec.md` — פרק 37: Hookline: מפרט וארכיטקטורה
Sections: 37.1 סיפור המשתמשת · 37.2 דרישות ובדיקות · 37.3 ארכיטקטורה · 37.4 מודל נתונים · 37.5 החלטות וחלופות ·
37.6 מצבי כשל · 37.7 איך מודדים הצלחה · 37.8 איך עובדים עם ה־starter. Days: 22 (whole chapter, with M1).

### 38 `38-hookline-milestones.md` — פרק 38: אבני הדרך
Sections: 38.1 M1: שלד ה־API · 38.2 M2: PostgreSQL וקליטה אידמפוטנטית · 38.3 M3: מטמון והמקבל המדומה · 38.4 M4:
outbox ו־Redpanda · 38.5 M5: מסירה, חתימה וניסיונות חוזרים · 38.6 M6: חוסן ועומס · 38.7 M7: קונטיינרים, Kubernetes
ו־CI · 38.8 M8: תצפיתיות ו־SLO · 38.9 M9: בדיקות ואבטחה · 38.10 M10: הערכה והדגמה.
Days: M1 22–23 (38.1), M2 29–30 (38.2), M3 33–34 (38.3), M4 39–40 (38.4), M5 43–44 (38.5), M6 54–55 (38.6), M7 68–70
(38.7), M8 77–79 (38.8), M9 85–86 (38.9), M10 98–99 (38.10).

### 39 `39-hookline-evaluation.md` — פרק 39: הערכה, הדגמה וראיון על הפרויקט
Sections: 39.1 עמידה ב־SLO תחת עומס · 39.2 ניסויי כאוס · 39.3 הוכחת אחריות המסירה · 39.4 עלות ההרצה · 39.5 תסריט
הדגמה · 39.6 פרסום · 39.7 עשרים שאלות על הפרויקט · 39.8 היום האחרון. Days: 98 (39.1–39.4), 99 (39.5–39.7), 100 (39.8).

---

## Appendices
A מילון מונחים (generated) · B פתרון תקלות · C גרסאות כלים · D תבניות Obsidian · E רשימות בדיקה · F לוח הלימוד היומי.

## Day map (summary)
| Days | Content |
| --- | --- |
| 0 | Installations (ch 2) |
| 1–14 | ch 0–6, checkpoint 1 on day 14 |
| 15–18 | ch 7–8 |
| 19–23 | ch 9, ch 37, M1 |
| 24–30 | ch 10–11, M2 |
| 31–35 | ch 12, M3, review day 35 |
| 36–40 | ch 13, M4 |
| 41–45 | ch 14, M5, checkpoint 2 on day 45 |
| 46–56 | ch 15–18, M6, checkpoint 3 on day 56 |
| 57 | Installations for Part E (ch 2 §2.10) |
| 58–71 | ch 19–23, M7, checkpoint 4 on day 71 |
| 72–79 | ch 24–25, M8 |
| 80–87 | ch 26–28, M9, checkpoint 5 on day 87 |
| 88–92 | ch 29–33 |
| 93–97 | ch 34–36 |
| 98–100 | M10, ch 39, checkpoint 6 and retrospective on day 100 |
