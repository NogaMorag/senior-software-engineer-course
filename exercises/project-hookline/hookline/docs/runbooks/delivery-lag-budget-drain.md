# Runbook: DeliveryLagBudgetDrain

| Field | Value |
| --- | --- |
| Alert | `DeliveryLagBudgetDrain` (`severity=ticket`) |
| Service | `hookline-worker` |
| SLO or signal | 99% of delivered deliveries arrive within 30 s of ingest, over 30 days |
| Rules | `deploy/observability/rules/delivery-slo.yml` |
| Owner | you (Hookline, milestone M8) |
| Last reviewed | `<date>` |

## What it means
Over 3 days, more deliveries arrived late than the SLO allows.

## Impact
A steady share of webhooks arrives late; the budget for risky worker changes shrinks.

## First five minutes
1. the "Hookline delivery pipeline" dashboard (http://localhost:3000, folder Hookline) over 3 days: lag percentiles, attempts by outcome, postponements.

## Diagnose
- Retries of a few always-failing receivers count as late when they finally succeed: is it one tenant?
- Worker capacity at peak hours.

## Mitigate
Fix capacity in working hours; talk to the tenant whose receivers fail.

## Escalate
Capacity limits outside Hookline (database, network): ticket to their owner.

## After
Keep the timeline (alert, acknowledgement, mitigation, recovery) and write the postmortem from
`docs/postmortem-TEMPLATE.md` within a few days.
