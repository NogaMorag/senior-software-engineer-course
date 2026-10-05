# Runbook: DeliveryAttemptsFailing

| Field | Value |
| --- | --- |
| Alert | `DeliveryAttemptsFailing` (`severity=page`) |
| Service | `hookline-worker` |
| SLO or signal | share of attempts without a 2xx (`job:delivery_attempt_failures:ratio_rate5m`) above 50% for 3 minutes |
| Rules | `deploy/observability/rules/delivery-slo.yml`, `operations.yml` |
| Owner | you (Hookline, milestone M8) |
| Last reviewed | `<date>` |

## What it means
Most delivery attempts, across all receivers, fail. One broken receiver cannot do that, so it is more
likely Hookline: a bad worker release, DNS, the network, an expired TLS root. The delivery SLO stays
silent meanwhile (nothing arrives, so nothing is late yet); this alert is what pages.

## Impact
Webhooks stop arriving for everyone. Deliveries wait and retry, and become dead after the last attempt; the
longer it lasts, the more need a replay.

## First five minutes
1. Acknowledge; open the "Hookline delivery pipeline" dashboard (http://localhost:3000, folder Hookline): attempts by outcome, receiver latency.
2. Read the worker's error lines in Loki: `{service_name="hookline-worker"} |= "delivery attempt failed"`.
   The `error` field names the cause (`no response within 1ms` is a timeout setting, `connection refused`
   a network, `x509` a certificate).
3. What changed in the worker in the last hour? Its startup log line prints its whole configuration.

## Diagnose
- **Every attempt times out at the same short duration:** `HOOKLINE_ATTEMPT_TIMEOUT` is wrong (the staged
  incident of milestone M8).
- **DNS or connection errors to many hosts:** the network or a NetworkPolicy.
- **Only one tenant's receivers:** not this alert's case; that is the breaker's job.

## Mitigate
Roll back the worker release (Kubernetes: `helm rollback hookline`; locally: restart the worker with the
previous configuration). Retries resume by themselves; check the dead deliveries afterwards and replay
them (`POST /v1/deliveries/{id}/replay`).

## Escalate
The cause is outside Hookline (network, DNS): call in its owner at once.

## After
Keep the timeline (alert, acknowledgement, mitigation, recovery) and write the postmortem from
`docs/postmortem-TEMPLATE.md` within a few days.
