# Runbook: DeliveryLagBudgetBurn

| Field | Value |
| --- | --- |
| Alert | `DeliveryLagBudgetBurn` (`severity=page`) |
| Service | `hookline-worker` |
| SLO or signal | 99% of delivered deliveries arrive within 30 s of ingest, over 30 days (spec N2, `job:slo_delivery:objective`) |
| Rules | `deploy/observability/rules/delivery-slo.yml` |
| Owner | you (Hookline, milestone M8) |
| Last reviewed | `<date>` |

## What it means
More deliveries arrive later than 30 s after their event was ingested than the SLO allows, at 6 to 14.4 times
the budget rate. The SLI counts only deliveries that arrived: it fires when the pipeline is slow, and also
right after an outage, when the backlog arrives late all at once.

## Impact
Tenants' customers get their webhooks late. Nothing is lost (deliveries wait as rows), but late is what the
SLO promises against.

## First five minutes
1. Acknowledge; open the "Hookline delivery pipeline" dashboard (http://localhost:3000, folder Hookline).
2. Is the lag recovering (a backlog draining after an outage) or still rising? The p99 lag panel says.
3. Outbox oldest row: if high, the delay is before the workers (relay); see `outbox-stalled.md`.
4. Attempts by outcome: a wall of retry means receivers fail; postponed by breaker means many endpoints are
   open-circuit.
5. Pending deliveries:

   ```bash
   docker compose -p hookline -f deploy/compose.yaml exec postgres psql -U hookline -d hookline -c \
     "SELECT status, count(*), min(next_attempt_at) FROM deliveries GROUP BY status"
   ```

## Diagnose
- **After an outage:** the alert follows the backlog; check that it shrinks.
- **Workers too few:** pending due rows grow while attempts succeed. Add workers.
- **A few receivers slow:** the bulkhead should keep them away from the others (N3); check postponed by bulkhead.

## Mitigate
Add worker replicas (`kubectl scale deploy/hookline-worker --replicas=N`). Roll back a recent worker or
dispatcher release. Do not raise the attempt timeout to make failures "succeed"; that holds workers longer.

## Escalate
Lag still rising 15 minutes after adding workers: the database is the likely limit; call in its owner.

## After
Keep the timeline (alert, acknowledgement, mitigation, recovery) and write the postmortem from
`docs/postmortem-TEMPLATE.md` within a few days.
