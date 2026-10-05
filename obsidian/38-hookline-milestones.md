---
direction: rtl
tags: [senior-course, book]
chapter: 38
---
# אבני הדרך

- **‏M1:** ממשקי מאגר; מימוש בזיכרון; אימות לפי גיבוב; 404 לכל משאב של דיירת אחרת.
- **‏M2:** מיגרציות לכל המודל; ‏`ON CONFLICT DO NOTHING` ואז `SELECT`; אירוע ו־outbox בטרנזקציה אחת.
- **‏M3:** מטמון לכל דיירת; מוחקים אחרי השמירה; זמן החיים הוא הגבול לרשימה ישנה; ה־sink עם worker אחד.
- **‏M4:** פרסום, אישור, ואז סימון; ‏offset אחרי ההכנסה; ‏`ON CONFLICT DO NOTHING` על אירוע ו־endpoint.
- **‏M5:** ‏`SKIP LOCKED` וחכירה; 408, 429 ו־5xx מנוסים שוב, שאר 4xx לא; ‏`attempt_base` בהפעלה חוזרת.
- **‏M6:** מחיצה בלי המתנה ומפסק לכל endpoint; דחייה בלי ניסיון; מגבלה לכל דיירת, נכשלת פתוח; השלה לפני האימות.
- **‏M7:** ‏image אחד עם כל הבינאריים; משתמש מספרי; מיגרציות ב־hook; ‏Secret מבחוץ; ‏compose גם ב־CI.
- **‏M8:** ‏`traceparent` בשורת האירוע; ‏SLI של פיגור על גבול 30; התראה על כשלים כי ה־SLI שותק; תחקיר עם מספרים שלי.
- **‏M9:** ‏SSRF נבדק ביצירה ובחיבור; הסתרה לפני כל פלט, וסריקה; סכמה משני הצדדים; ‏VUs לפי קצב כפול זמן תגובה.
- **‏M10:** הורגים ב־SIGKILL; סט צפוי מהמסד ולא משורות המסירה; פירוק הפיגור לשלבים; זמן סשן קצר לצרכנים.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
M1 tests  DATABASE_URL= go test -race -skip 'Invalidat|Deliver|Replay|ActiveSubscriptions|RateLimit|OverItsLimit|LimiterFailure|Shed|Trace|PrivateAddress' ./internal/api/ ./internal/store/... ./internal/admin/
M2 db     make db-up && export DATABASE_URL=postgres://hookline:change-me@localhost:5438/hookline
M2 tests  go test -race -count=1 -skip 'Invalidat|Deliver|Replay|ActiveSubscriptions|RateLimit|OverItsLimit|LimiterFailure|Shed|Trace|PrivateAddress' ./internal/api/ ./internal/store/... ./internal/admin/ ./internal/db/ ./migrations/
admin     go run ./cmd/hookline-admin migrate up | create-tenant -name <name> | create-topic -partitions 6
stack     make up   (postgres, redis, redpanda, console :8098, sink :8091)   make down
env       REDIS_ADDR=localhost:6438 KAFKA_BROKERS=localhost:19438 SINK_URL=http://localhost:8091
run       make run | make relay | make dispatcher | make worker
sink      curl -s localhost:8091/stats    curl -s -X POST localhost:8091/control -d '{"failure_rate":1}'
retry     408 429 5xx timeout refused -> retry      other 4xx 3xx -> dead      replay: dead -> pending
m7        make images | make chart-check | KEEP=1 make kind-smoke | make tf-check | make kind-down
kube      export KUBECONFIG=/tmp/hookline-kind.kubeconfig; kubectl -n hookline get pods
ci        cp -r ~/senior-course/hookline/.github/. ~/senior-course/.github/ && (cd ~/senior-course && actionlint)
m8        make obs-up | make promtool | make check-telemetry | make stage-incident | make obs-down
grafana   http://localhost:3000 (folder Hookline)   prometheus :9090   tempo :3200   loki :3100
m9        make load-ingest | make load-two-tenants | make images && make supply-chain   (RATE= NOISY_RATE= QUIET_RATE=)
m9 tests  DATABASE_URL= go test -race ./internal/ssrf/ ./internal/redact/ ./internal/page/ ./contract/ ./internal/api/
m10       make chaos (SEED= RATE= DURATION=) | make verify-delivery (SINCE= WAIT=) | make load-overload | scripts/stage-lag.sh
lab       make lab-n3      GUARDS=off make lab-n3      (record your p50/p99, date, machine)
guards    HOOKLINE_ENDPOINT_CONCURRENCY=2 HOOKLINE_BREAKER_FAILURES=5 HOOKLINE_BREAKER_OPEN=30s HOOKLINE_MAX_IN_FLIGHT=256
m5 tests  go test -race -count=1 -skip 'RateLimit|OverItsLimit|LimiterFailure|Shed|Breaker|Bulkhead|Trace|PrivateAddress' ./internal/backoff/ ./internal/delivery/ ./internal/api/ ./internal/store/... ./internal/e2e/
skip      -skip drops later milestones' tests: Invalidat ActiveSubscriptions (M3), Deliver Replay (M5), StaleRelay RateLimit OverItsLimit LimiterFailure Shed Breaker Bulkhead (M6), Trace (M8), PrivateAddress (M9)
```

</div>

קישורים: [[37-hookline-spec]] · [[39-hookline-evaluation]]
