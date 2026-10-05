# Terminology policy (book 2)

Written 2026-10-05 (Stage 1b). Applies to every chapter, exercise README, appendix and Obsidian note. It extends the
rules in `CONVENTIONS.md` ("Hebrew first, English in parentheses on first use") with a domain decision: **which words
stay English and which are translated**, and the one rendering each translated concept uses. Where book 1
(`ai-engineering-course`, appendix A) already has a rendering, this book reuses it unless the brief fixed another.

## 1. Rule of thumb

- **Product nouns and their own vocabulary stay English**, with a Hebrew gloss on first use in a chapter:
  "pod (היחידה הקטנה ביותר שרצה ב־Kubernetes: קונטיינר אחד או יותר)". Afterwards the English word, unformatted in
  prose when it is a concept (pod, topic), in backticks when it is an identifier or a command (`kubectl`, `Deployment`
  as a manifest `kind`).
- **Concepts are translated**, one rendering each, listed below. On first use per chapter: Hebrew, then English in
  parentheses. Never alternate between two Hebrew renderings of the same concept in one book.
- **Metrics and acronyms stay English** as in book 1: p99, RED, USE, SLI, MTTR, QPS, ACID, CAP, PACELC, RPO/RTO.
  SLO is introduced once as "יעד רמת שירות (SLO)" and then written SLO.
- **Hebrew grammar attaches with a maqaf**: ה־pod, ב־Kafka, ל־goroutine. Plurals of English nouns: "שלושה pods",
  not "פודים"; in prose about a kind of object prefer the English plural.
- A new term not in this file: pick one rendering, use it consistently, add it to the chapter's terms table (appendix
  A is generated from those tables) and, if it will recur across chapters, add it here in the same change.

## 2. Stays English (gloss on first use)

| Area | Words |
| --- | --- |
| Go | goroutine, channel, `select`, context, slice, map, struct, receiver, module, package, generics, `defer`, panic, `go vet`, pprof (interface and method are translated: see §3) |
| Python | asyncio, coroutine, `asyncio.Task`, `TaskGroup`, FastAPI, pydantic, fixture, uv, ruff, mypy |
| Containers | Docker, image, container, layer, Dockerfile, registry, tag, distroless, compose |
| Kubernetes | Kubernetes, cluster, node, pod, Deployment, ReplicaSet, Service, Ingress, ConfigMap, Secret (the object), namespace, probe, HPA, Helm, chart, release, values, kind, kubectl |
| Kafka | Kafka, Redpanda, topic, partition, offset, consumer group, producer, consumer, broker, rebalance, `acks` |
| Infra & CI | Terraform, provider, resource, state, module, plan, apply, GitHub Actions, workflow, job, step, runner, action |
| Observability | OpenTelemetry, span, Prometheus, PromQL, Grafana, Loki, Tempo, exporter, collector, histogram, counter, gauge, label |
| Data | PostgreSQL, Redis, index (B-tree, GIN), `EXPLAIN`, WAL, replica, primary, pgx, goose, advisory lock |
| APIs | REST, gRPC, protobuf, OpenAPI, JSON, endpoint, header, webhook, cursor, JWT, OIDC, HMAC |

## 3. Translated concepts (one rendering each)

| English | Hebrew | Note |
| --- | --- | --- |
| consistency | עקביות | consistency model: מודל עקביות |
| eventual consistency | עקביות בסופו של דבר | |
| replication | שכפול | replica lag: פיגור שכפול |
| partitioning (data) | חלוקה למחיצות | a Kafka partition stays "partition" |
| idempotency | אידמפוטנטיות | book 1 |
| idempotency key | מפתח אידמפוטנטיות | book 1 |
| retry | ניסיון חוזר | book 1 |
| backoff | המתנה גדלה | exponential backoff with jitter: המתנה גדלה מעריכית עם פיזור אקראי (jitter) |
| jitter | פיזור אקראי (jitter) | |
| timeout | זמן קצוב | book 1 |
| deadline | גבול זמן כולל | book 1; a context deadline in Go is the same concept |
| circuit breaker | מפסק זרם | book 1 |
| bulkhead | מחיצה אטומה | |
| rate limiting | הגבלת קצב | a rate limit (the number): מגבלת קצב (book 1) |
| back-pressure | לחץ חוזר | |
| load shedding | השלת עומס | |
| graceful degradation | הידרדרות מכובדת | book 1 |
| graceful shutdown | כיבוי מסודר | |
| observability | תצפיתיות | book 1 |
| trace | מעקב | book 1; span stays English |
| metric | מדד | book 1 |
| log | לוג | book 1; structured log: לוג מובנה |
| SLO | יעד רמת שירות (SLO) | SLI: מדד רמת שירות (SLI) |
| error budget | תקציב שגיאות | burn rate: קצב שריפה |
| alert | התראה | |
| on-call | תורנות | |
| incident | תקרית | |
| postmortem | תחקיר | blameless: ללא האשמה |
| runbook | runbook | stays English (no settled Hebrew term) |
| migration (schema) | מיגרציה | |
| migration (system) | הגירה | |
| transaction | טרנזקציה | isolation level: רמת בידוד |
| lock | נעילה | deadlock: קיפאון |
| leader election | בחירת מנהיג | |
| consensus | קונצנזוס | |
| saga | סאגה | compensating action: פעולה מפצה |
| outbox pattern | תבנית ה־outbox | |
| dead-letter queue | תור מתים | |
| at-least-once / exactly-once | לפחות פעם אחת / בדיוק פעם אחת | in quotes on first use |
| cache | מטמון | cache stampede: התפרצות מטמון; invalidation: אי־תוקף |
| throughput | תפוקה | |
| latency | זמן תגובה | tail latency: זמן התגובה בזנב (p99) |
| secret | סוד | book 1 (the Kubernetes object stays `Secret`) |
| least privilege | מינימום הרשאות | book 1 |
| threat model | מודל איומים | book 1 |
| authentication / authorization | אימות / הרשאה | |
| infrastructure as code | תשתית כקוד | |
| design doc | מסמך תכנון | ADR: רשומת החלטה (ADR) |
| code review | סקירת קוד | |
| static analysis | ניתוח סטטי | |
| ownership | בעלות | in Rust boxes, "ownership" of memory is בעלות as well; the context disambiguates |
| technical debt | חוב טכני | |
| feature flag | דגל תכונה | |
| race condition / data race | מרוץ / מרוץ נתונים | |
| garbage collector | אוסף זבל (GC) | |
| property-based test | בדיקת תכונה | |
| contract test | בדיקת חוזה | |
| load test | בדיקת עומס | |
| table-driven test | בדיקה טבלאית | |
| benchmark | בנצ'מרק | |
| cache hit / miss | פגיעה / פספוס | hit ratio: יחס פגיעות |
| eviction | פינוי | |
| TTL | זמן חיים (TTL) | |
| negative caching | מטמון שלילי | |
| router | נתב | |
| fan-out / fan-in | פיזור / איסוף (fan-out, fan-in) | never פריסה, which is deploy |
| expand and contract | הרחבה ואז כיווץ | |
| backfill | מילוי לאחור | |
| dual write | כתיבה כפולה | |
| rolling deploy | פריסה מתגלגלת | |
| strangler fig | strangler fig (תאנה חונקת) | English in headings |
| tolerant reader | קורא סובלני | |
| non-goal | לא־מטרה | |
| region | אזור | |
| fencing token | fencing token (אסימון גידור) | |
| tenant | דיירת (plural דיירות) | Hookline's customer account; feminine throughout, with agreeing verbs and adjectives |
| subscription | מנוי | endpoint stays English |
| request ID | מזהה בקשה | |
| access log | לוג גישה | |
| grace period | זמן חסד | |
| cooperative cancellation | ביטול שיתופי | |
| snapshot | תמונת מצב | |
| sequential scan | סריקה מלאה | index-only scan: סריקת אינדקס בלבד; keyset stays English |
| write skew | write skew | glossed once as הטיית כתיבה |
| interface (Go, and the general concept) | ממשק | the keyword stays `interface`; Rust's trait stays trait |
| method | מתודה | receiver stays English |
| event loop | לולאת האירועים | asyncio's loop; the identifier stays English |
| task (asyncio, background job) | משימה | the class `asyncio.Task` stays an identifier |
| continuous integration | אינטגרציה רציפה (CI) | chapter 23; CI stays English afterwards |
| smoke test | בדיקת עשן | chapter 2 and chapter 23 |
| composite action | פעולה מורכבת | GitHub Actions; "action" itself stays English |
| reusable workflow | workflow לשימוש חוזר | |
| required reviewer (environment) | מבקר חובה | |
| estimate | הערכת זמנים | chapter 31; given as a range (טווח) with a confidence (רמת ביטחון) |
| mentoring | חניכה | chapter 1 and 31 |
| agent (AI coding) | סוכן | book 1; chapter 32 |
| prompt | הנחיה | book 1; chapter 32 |
| spec | מפרט | chapter 32: what an agent or a person is asked to build, with its edges |

Appendices are named with Hebrew letters in prose: נספח א … נספח ו (never the Latin letter, as in "נספח C").

## 4. Words to avoid

| Avoid | Use |
| --- | --- |
| „פודים”, „קונטיינרים” as Hebrew plurals in technical prose | pods, containers |
| „ריטריי”, „טיימאאוט”, „קאש” | ניסיון חוזר, זמן קצוב, מטמון |
| „דיפלוי” | פריסה (deploy as a verb: לפרוס) |
| „בק־אנד” in Hebrew letters | backend |
| „מימוש” for deployment | פריסה |
| „לרנדר” | להפיק (a template), להציג (a page) |
