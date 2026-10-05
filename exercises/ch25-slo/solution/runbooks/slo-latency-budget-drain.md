# Runbook: SLOLatencyBudgetDrain

| Field | Value |
| --- | --- |
| Alert | `SLOLatencyBudgetDrain` (`severity=ticket`) |
| Service | the `job` label of the alert (for chapter 25: `shorty`) |
| SLO | `job:slo_latency:objective` of the requests finish within 250 ms, over 30 days (shorty: 99%) |
| Owner | you (chapter 25) |
| Last reviewed | `<date>` |

## What it means
Over 3 days, and still over 6 hours, more requests took longer than 250 ms than the SLO allows. Not
urgent; it will exhaust the latency budget before the 30 days end.

## First steps (working hours)
1. Duration panel over 7 days: a step (a release, a dependency) or a slope (growing data, growing
   traffic)?
2. Slowest routes with the p99 query of `slo-latency-budget-burn.md`, using `[1d]`.
3. Compare a slow trace with a fast one of the same route.

## Mitigate
A performance fix through the normal release process: a profile (chapter 6), an index (chapter 10), a
cache (chapter 12). Measure before and after.

## Escalate
If the slowness is in another team's dependency, open a ticket with them and the numbers.

## After
Record the cause. If 250 ms is the wrong promise for this service, change the SLO deliberately, with
the people who depend on it, rather than silencing the alert.
