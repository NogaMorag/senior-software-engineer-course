# Runbook: SLOLatencyBudgetBurn

| Field | Value |
| --- | --- |
| Alert | `SLOLatencyBudgetBurn` (`severity=page`) |
| Service | `hookline-api` |
| SLO or signal | 99% of API requests finish within 250 ms, over 30 days (`job:slo_latency:objective`) |
| Rules | `deploy/observability/rules/slo-recording.yml`, `slo-alerts.yml` |
| Owner | you (Hookline, milestone M8) |
| Last reviewed | `<date>` |

## What it means
More requests than the SLO allows take longer than 250 ms, at 6 to 14.4 times the budget rate. (N1 asks
200 ms for ingest at a fixed load; k6 measures that, and the alert uses the nearest bucket, 0.25 s.)

## Impact
Tenants wait; their clients may time out and retry, which adds load. Retries are safe (idempotency keys).

## First five minutes
1. Acknowledge; open the "Hookline API" dashboard (http://localhost:3000, folder Hookline), Duration panel: p99 of POST /v1/events.
2. Goroutines and memory panel: a saturated process shows here first.
3. Open a slow request's trace from an exemplar: which span takes the time? (The ingest transaction is the
   usual one.)

## Diagnose
- **Database:** slow `INSERT` into events and outbox; check locks and connections.
- **Redis:** the rate limiter call; a slow Redis slows every ingest.
- **Load:** requests per second per tenant; is one tenant above its limit and still costing work?

## Mitigate
Roll back a recent change. Under real load add API replicas. Shedding (503) trades latency for errors;
do not raise its limit while latency burns.

## Escalate
No improvement after 15 minutes, or the database is the slow span: call in the database owner.

## After
Keep the timeline (alert, acknowledgement, mitigation, recovery) and write the postmortem from
`docs/postmortem-TEMPLATE.md` within a few days.
