# Postmortem: webhooks stopped arriving for 8 minutes after a worker release

> Blameless: this document describes what the system and the people in it did, and why each step made
> sense with the information available at the time. It names roles, not culprits.
>
> This is the deliberately broken release of milestone M8, staged with `scripts/stage-incident.sh`.
> The section "Author's run" records the run of the person who built the reference implementation, on
> their machine, so that you can see what a filled-in postmortem looks like. It is **not** a result to
> copy: your run has its own times and numbers, and they go in the section "Your run". The rest of the
> document (cause, factors, action items) should be rewritten from your run, not kept because it is
> already here.

| Field | Value |
| --- | --- |
| Date | 2026-10-05 (author's run) / `<your date>` |
| Authors | `<you>` |
| Status | draft |
| Severity | SEV2: every receiver affected, no data lost |
| Duration | author's run: 8 min 19 s from the release to the rollback / yours: `<minutes>` |
| Detection | `DeliveryAttemptsFailing`, author's run: 8 min 19 s after the release / yours: `<minutes>` |

## Summary
A worker release set the per-attempt timeout to 1 ms instead of 1 s (`HOOKLINE_ATTEMPT_TIMEOUT=1ms`, a unit
typo). From then on every delivery attempt timed out before any receiver could answer. Ingest kept answering
202, so the API's SLOs saw nothing; events piled up as pending deliveries. `DeliveryAttemptsFailing` paged
about 8 minutes later, the release was rolled back, and the backlog was delivered within seconds, late. No
event was lost and no delivery became dead.

## Impact
- Receivers affected: all (in the staged run, one tenant with one endpoint).
- Delivery SLO (N2): deliveries later than 30 s after ingest: author's run 1839 of 4526 delivered;
  yours `<n>` of `<m>`.
- Longest lag: author's run 523 s; yours `<s>`.
- Data: none lost; deliveries waited as rows and retried (`attempts` author's run 4549 for 4527
  deliveries); dead: author's run 0, yours `<n>`.
- API: unaffected (202 throughout).

## Timeline (UTC)

### Author's run (2026-10-05, the reference implementation's machine)

| Time | Event |
| --- | --- |
| 13:15:40 | Traffic starts, about 5 events per second, one endpoint on the compose sink. |
| 13:17:40 | Baseline: attempt failure share 0, late share 0. |
| 13:17:40 | CHANGE: worker restarted with `HOOKLINE_ATTEMPT_TIMEOUT=1ms`. |
| 13:17:42 | The endpoint's breaker opens after 5 consecutive timeouts (worker log `endpoint breaker`, closed to open). From now on most deliveries are postponed, not attempted. |
| 13:22:56 | `DeliveryAttemptsFailing` pending: the 5-minute window no longer holds the healthy baseline. |
| 13:25:59 | `DeliveryAttemptsFailing` firing (after its 3-minute `for`). PAGE. |
| 13:25:59 | MITIGATION: worker restarted with the default timeout. |
| 13:26:19 | `DeliveryAttemptsFailing` resolved; attempt failure share 0.011; late share 0.930 (the backlog arriving). |
| 13:26:30 | `DeliveryLagBudgetBurn` firing: the backlog arrived late, as the SLO counts it. |
| 13:36:01 | `DeliveryLagBudgetBurn` still firing: its 6h/30m pair keeps the late backlog in view for about 30 minutes. One delivery still pending (a scheduled retry). |

### Your run

| Time | Event |
| --- | --- |
| `hh:mm:ss` | Traffic starts. |
| `hh:mm:ss` | CHANGE. |
| `hh:mm:ss` | Breaker opens (worker log). |
| `hh:mm:ss` | `DeliveryAttemptsFailing` pending. |
| `hh:mm:ss` | `DeliveryAttemptsFailing` firing. |
| `hh:mm:ss` | MITIGATION. |
| `hh:mm:ss` | `DeliveryLagBudgetBurn` firing, then resolved. |

## Root cause and contributing factors
- The cause: a duration value with the wrong unit. `1ms` is a valid duration, so configuration validation
  accepted it (it only requires the lease to be longer than the timeout).
- No check compares the configured timeout with what receivers need; anything below about 100 ms is
  almost certainly a mistake.
- Ingest stayed green, so the API's dashboards and SLOs looked healthy throughout.
- The delivery SLO counts only deliveries that arrive: while nothing arrived, it had no data.
- Detection took about 8 minutes: the 5-minute rate window still held healthy attempts for 5 minutes, and
  the breaker (correctly) cut the attempt rate, so failures accumulated slowly; then the alert's 3-minute
  `for` added the rest.

## What went well
- The breaker opened within 2 seconds and stopped hammering the receiver with doomed requests.
- Nothing was lost: deliveries waited as rows, retried, and none became dead.
- The rollback took effect at once; the backlog was delivered within seconds.
- The trace of a failing delivery (`hookline.deliver`, error `no response within 1ms`) named the cause.

## What went badly
- 8 minutes to page for a total outage of deliveries.
- The release had no canary: one worker with the new setting, watched for a few minutes, would have shown
  the failure share at once.

## Where we were lucky
- Retries were capped well below the default 10 attempts' worth of backoff (the staged run uses a 30 s cap),
  so the backlog drained quickly; with the production cap of 15 minutes it would have taken longer.

## Action items

| Item | Kind | Owner | Due | Ticket |
| --- | --- | --- | --- | --- |
| Reject `HOOKLINE_ATTEMPT_TIMEOUT` below 100 ms in `config.Load` | prevent | `<you>` | `<date>` | `<id>` |
| Shorten detection: a 1-minute `for` on `DeliveryAttemptsFailing`, or a second window of 2 minutes | detect | `<you>` | `<date>` | `<id>` |
| Canary worker releases: one pod first, promoted after 5 minutes without a rise in failures | prevent | `<you>` | `<date>` | `<id>` |
| An SLI that also sees deliveries that have not arrived (age of the oldest due delivery) | detect | `<you>` | `<date>` | `<id>` |

## Lessons
An SLO built only from successes cannot see an outage in which nothing succeeds. Pair it with a signal
on failures, and test that pair with promtool (scenario 4 of `deploy/observability/tests/rules_test.yml`).
