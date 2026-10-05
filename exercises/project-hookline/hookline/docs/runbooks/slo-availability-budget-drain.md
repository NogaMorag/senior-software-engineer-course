# Runbook: SLOAvailabilityBudgetDrain

| Field | Value |
| --- | --- |
| Alert | `SLOAvailabilityBudgetDrain` (`severity=ticket`) |
| Service | `hookline-api` |
| SLO or signal | 99.9% of API requests are not answered 5xx, over 30 days |
| Rules | `deploy/observability/rules/slo-alerts.yml` |
| Owner | you (Hookline, milestone M8) |
| Last reviewed | `<date>` |

## What it means
Over 3 days (and still over 6 hours) the API failed more requests than the SLO allows: a slow leak, not an
outage. At this rate the 30-day budget runs out before the month ends.

## Impact
Some tenants see occasional 5xx. Nothing is on fire; the budget for risky changes is shrinking.

## First five minutes
1. Open the "Hookline API" dashboard (http://localhost:3000, folder Hookline) over the last 3 days.
2. Find the route and the code that fail most, with the query of `slo-availability-budget-burn.md`, over `[3d]`.

## Diagnose
- One route, steady: a bug or a slow query on that path.
- Spikes at certain hours: load, a batch job, a noisy tenant (check 429 and 503).

## Mitigate
Fix in working hours. Until the budget recovers, slow down risky releases (that is what the budget is for).

## Escalate
If the cause is outside Hookline (the database, the network), open a ticket with its owner.

## After
Keep the timeline (alert, acknowledgement, mitigation, recovery) and write the postmortem from
`docs/postmortem-TEMPLATE.md` within a few days.
