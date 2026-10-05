# Proposal — "Senior Software Engineer in Industry" (book 2) — reconstructed 2026-10-05

> **Provenance.** The approved `02-proposal.md` did not transfer from the authoring machine. This reconstruction
> follows the handoff brief section by section (reader §1, decisions §3, content per part §10, schedule rules §11).
> The chapter list (§4) and the capstone shape (§6) are rebuilt to satisfy every constraint the brief states; they are
> the author's proposal and the reader can amend them. Evidence: `01-market-research.md` (fresh pass, 2026-10-05).

## 1. Reader
See `03-decisions-2026-10-04.md` D1. The book assumes she can program well and reads code fluently; it assumes
nothing about Go, Kubernetes, Kafka, Terraform, observability stacks, or how senior loops are run in 2026.

## 2. Language decision
Go is the primary language: it is the language of the cloud-native infrastructure the lane works on (Kubernetes,
Terraform and Prometheus are written in it) and appears in roughly a third of the sampled Israeli senior postings,
usually next to Python (`01-market-research.md` §1). Python leads the same sample, which is why it is the second
production language rather than an afterthought. Because the
reader already knows Python, the book raises it to senior production level instead of re-teaching it. Rust is her edge:
the book does not teach it but tells her where it is valued and how to present it, and uses it as the comparison
point when teaching Go.

## 3. Stack
| Area | Choice | Why this one |
| --- | --- | --- |
| Primary language | Go (stdlib `net/http`, `log/slog`, `pgx`, `go-redis`, `franz-go`) | Cloud-native default; small surface |
| Second language | Python 3.14 with uv, ruff, mypy, pytest, FastAPI, pydantic | Senior Python expected in postings |
| Database | PostgreSQL (container) | The default OLTP store; deep enough for isolation and locking labs |
| Cache | Redis (container) | Caching patterns, rate limiting, locks |
| Log/stream | Redpanda (Kafka API, single binary container) | Kafka semantics without a ZooKeeper/KRaft cluster to run |
| Containers | Docker Desktop (WSL2 backend) | What postings assume |
| Kubernetes | kind + kubectl + Helm | Free, local, real API server |
| IaC | Terraform with the Kubernetes and Helm providers; optional AWS free-tier section | No cloud account required |
| CI | GitHub Actions (validated with actionlint; run in her repository) | The common CI in postings |
| Observability | OpenTelemetry SDKs, Prometheus, Grafana, Loki, Tempo via compose | Vendor-neutral, free |
| Load and tests | k6, testcontainers-go, rapid (property tests) | Standard, scriptable |
| Security | govulncheck, Trivy or Grype, Syft SBOM, OIDC/JWT libraries | Free and current |

Exact versions are pinned in `07-tech-stack.md`.

## 4. Structure: Part 0 + ten parts, 40 chapters
| Part | Chapters |
| --- | --- |
| 0. לפני שמתחילים | 00 איך ללמוד מהספר; 01 מה זה „בכיר” ב־2026; 02 סביבת העבודה: WSL2, Docker, Go, uv |
| A. Go | 03 צעדים ראשונים ב־Go; 04 טיפוסים, ממשקים ושגיאות; 05 מקביליות; 06 שירות HTTP לייצור |
| B. Python בכירה | 07 Python עם טיפוסים, אריזה ובדיקות; 08 Python אסינכרונית ו־FastAPI |
| C. יסודות backend | 09 עיצוב API ו־gRPC; 10 PostgreSQL: שאילתות ואינדקסים; 11 PostgreSQL: טרנזקציות, נעילות ומיגרציות; 12 Redis ומטמון; 13 Kafka עם Redpanda; 14 עבודות רקע וניסיונות חוזרים |
| D. מערכות מבוזרות | 15 זמנים קצובים, ניסיונות חוזרים ומפסקי זרם; 16 עומס: הגבלת קצב והשלת עומס; 17 עקביות, שכפול וחלוקה למחיצות; 18 נעילות, בחירת מנהיג וסאגות |
| E. Cloud-native | 19 Docker כמו שצריך; 20 Kubernetes על kind; 21 הפעלת Kubernetes ו־Helm; 22 Terraform; 23 GitHub Actions |
| F. תצפיתיות ותפעול | 24 טלמטריה: OpenTelemetry, Prometheus, Grafana, Loki; 25 SLO, התראות, תורנות ותחקיר |
| G. איכות ואבטחה | 26 אסטרטגיית בדיקות ובדיקות עומס; 27 סקירת קוד וניתוח סטטי; 28 אבטחה לשירותי backend |
| H. אומנות הבכירות | 29 שיטת תכנון ומסמכי תכנון; 30 מערכות שמשתנות; 31 לעבוד כבכירה; 32 הנדסה בעזרת AI; 33 היתרון של Rust |
| I. קריירה | 34 תהליך הגיוס הבכיר; 35 תרגילי קוד מעשיים; 36 תכנון מערכות בראיון ומאגר השאלות |
| J. פרויקט הגמר | 37 Hookline: מפרט וארכיטקטורה; 38 אבני הדרך; 39 הערכה, הדגמה וראיון על הפרויקט |

Part F is two chapters and Part G three (instead of 3+3) to keep chapter 33 as the Rust chapter, as the brief
fixes; SLOs, alerting, on-call and the postmortem share chapter 25 because they form one loop (budget, alert,
incident, learning).

## 5. Program
100 study days in 20 weeks plus Day 0; 75–120 minutes a day including debugging; six checkpoints (end of Parts A,
C, D, E, G and the capstone); install-heavy days separate (Day 0 and a second slot before Part E); one lighter day
a week; milestones M1–M10 two to three days each, interleaved after the part that teaches them; the last week is mock
interviews, polish, publishing, checklists and a retrospective. Day-by-day: `05-daily-schedule.md` (English source)
and appendix F (Hebrew).

## 6. Capstone: Hookline
A webhook delivery platform: tenants register endpoints and subscriptions, submit events through an idempotent API,
and Hookline delivers each event to every subscribed endpoint at least once, signed (HMAC), with retries and
exponential backoff with jitter, per-endpoint circuit breaking, per-tenant rate limiting, a dead-letter state with
replay, and a full attempt log. Built from Go services (API, dispatcher, delivery workers), PostgreSQL (state, outbox),
Redpanda (delivery queue), Redis (rate limits), a Python FastAPI receiver simulator that misbehaves on purpose, and
the observability stack; deployed to kind with Helm, provisioned with Terraform, built and tested by GitHub Actions.
Full specification: `06-final-project.md`. It demonstrates exactly the senior topics an interviewer probes:
idempotency, delivery guarantees, back-pressure, failure isolation, SLOs and operating under failure.
