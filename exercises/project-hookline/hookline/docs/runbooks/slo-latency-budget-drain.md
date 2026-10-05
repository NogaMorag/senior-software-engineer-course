# Runbook: SLOLatencyBudgetDrain

| Field | Value |
| --- | --- |
| Alert | `SLOLatencyBudgetDrain` (`severity=ticket`) |
| Service | `hookline-api` |
| SLO or signal | 99% of API requests finish within 250 ms, over 30 days |
| Rules | `deploy/observability/rules/slo-alerts.yml` |
| Owner | you (Hookline, milestone M8) |
| Last reviewed | `<date>` |

## What it means
Over 3 days, more requests than the SLO allows were slower than 250 ms: a slow drift.

## Impact
Slower ingest for tenants; less budget for risky changes.

## First five minutes
1. the "Hookline API" dashboard (http://localhost:3000, folder Hookline) over 3 days: when does p99 rise? 2. Compare with traffic and with deploys.

## Diagnose
- A growing table without the right index (chapter 10).
- A tenant's traffic pattern.
- Memory or GC pressure after a release.

## Mitigate
Fix in working hours: an index, a query, a resource limit.

## Escalate
Database-side causes: ticket to its owner.

## After
Keep the timeline (alert, acknowledgement, mitigation, recovery) and write the postmortem from
`docs/postmortem-TEMPLATE.md` within a few days.
