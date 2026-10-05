# Runbook: OutboxStalled

| Field | Value |
| --- | --- |
| Alert | `OutboxStalled` (`severity=page`) |
| Service | `hookline-relay` |
| SLO or signal | age of the oldest unpublished outbox row (`hookline_outbox_oldest_age_seconds`) above 60 s for 2 minutes |
| Rules | `deploy/observability/rules/operations.yml` |
| Owner | you (Hookline, milestone M8) |
| Last reviewed | `<date>` |

## What it means
Events are accepted (202) but do not leave the outbox: no relay leads, the leader cannot publish to Redpanda,
or it cannot mark rows. Ingest keeps working, so no API SLO notices.

## Impact
Nothing is delivered for any tenant. Nothing is lost: every event waits in the outbox and goes out once the
relay publishes again.

## First five minutes
1. Acknowledge; the "Hookline delivery pipeline" dashboard (http://localhost:3000, folder Hookline): outbox rows waiting (rising).
2. Is any relay leading? Its log says `relay leading` with a term; the follower says `relay campaigning`.
3. The leader's warnings: `relay run failed; retrying` with the Kafka error.

## Diagnose
- **No leader:** every relay pod is down, or they cannot reach PostgreSQL (the lock lives there).
- **Publish fails:** Redpanda is down or unreachable (`rpk cluster info` in its container).
- **`ErrNotLeader` in a loop:** two relays fight over terms; one has a stale session.

## Mitigate
Restart the relay (`kubectl rollout restart deploy/hookline-relay`). Restore Redpanda. The backlog drains in
publish order; expect the delivery lag alert to follow for a while.

## Escalate
Redpanda or PostgreSQL itself is down: call in their owner.

## After
Keep the timeline (alert, acknowledgement, mitigation, recovery) and write the postmortem from
`docs/postmortem-TEMPLATE.md` within a few days.
