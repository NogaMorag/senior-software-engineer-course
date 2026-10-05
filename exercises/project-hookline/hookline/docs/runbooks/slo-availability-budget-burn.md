# Runbook: SLOAvailabilityBudgetBurn

| Field | Value |
| --- | --- |
| Alert | `SLOAvailabilityBudgetBurn` (`severity=page`) |
| Service | `hookline-api` |
| SLO or signal | 99.9% of API requests are not answered 5xx, over 30 days (`job:slo_availability:objective`) |
| Rules | `deploy/observability/rules/slo-recording.yml`, `slo-objectives.yml`, `slo-alerts.yml` |
| Owner | you (Hookline, milestone M8) |
| Last reviewed | `<date>` |

## What it means
The API answers 5xx at 6 to 14.4 times the rate the SLO allows, over a long and a short window (1h and 5m at
14.4, or 6h and 30m at 6). At 14.4 the 30-day budget lasts about 2 days. The short window says it is
happening now. 503 from load shedding counts: shedding protects the service but users still see failures.

## Impact
Tenants cannot send events (POST /v1/events) or manage endpoints. Their own retries with the same
Idempotency-Key are safe, but events they give up on are never delivered.

## First five minutes
1. Acknowledge the page; say in the incident channel that you are looking.
2. Open the "Hookline API" dashboard (http://localhost:3000, folder Hookline). Errors panel: 5xx share (normal 0) and 503 per second (shedding).
3. Which route and which code:

   ```bash
   curl -s --get http://localhost:9090/api/v1/query --data-urlencode \
     'query=sum by (http_route, http_response_status_code) (rate(http_server_request_duration_seconds_count{job="hookline-api", http_response_status_code=~"5.."}[5m]))'
   ```

4. Error logs of the last 15 minutes in Grafana, Explore, Loki: `{service_name="hookline-api"} | detected_level="error"`.
5. Open one failing request's trace from a log line ("Open trace") or an exemplar on the Duration panel.

## Diagnose
- **503 only:** the shedder (`HOOKLINE_MAX_IN_FLIGHT`) is refusing work. Why are requests slow to
  finish? Usually the database: check the Duration panel and PostgreSQL's connections.
- **500 on POST /v1/events:** the ingest transaction fails. The log line's error names PostgreSQL
  (connection refused, a constraint, a missing column after a bad migration).
- **Did something change?** A deploy, a migration, a config value. Most incidents follow a change.

## Mitigate
- A recent change: roll it back first (`helm rollback hookline` in Kubernetes; restart the previous
  binary locally).
- The database is down or full: restore it; the API recovers by itself, and tenants' retries with the same
  keys create no duplicates.
- Shedding under a real overload: add API replicas, or raise `HOOKLINE_MAX_IN_FLIGHT` only if the
  database has room; never remove the limit.
- It worked when the 5xx share is back at 0; the alert resolves within about 5 minutes.

## Escalate
No improvement 15 minutes after a rollback, or PostgreSQL itself is failing: call in whoever owns the database.

## After
Keep the timeline (alert, acknowledgement, mitigation, recovery) and write the postmortem from
`docs/postmortem-TEMPLATE.md` within a few days.
