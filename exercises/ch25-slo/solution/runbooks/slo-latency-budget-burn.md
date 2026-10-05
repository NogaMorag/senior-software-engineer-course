# Runbook: SLOLatencyBudgetBurn

| Field | Value |
| --- | --- |
| Alert | `SLOLatencyBudgetBurn` (`severity=page`) |
| Service | the `job` label of the alert (for chapter 25: `shorty`) |
| SLO | `job:slo_latency:objective` of the requests finish within 250 ms, over 30 days (shorty: 99%) |
| Owner | you (chapter 25) |
| Last reviewed | `<date>` |

## What it means
Requests slower than 250 ms are more than 6 to 14.4 times as frequent as the SLO allows, over a long
and a short window. Users wait; some clients give up and retry, which adds load.

## Impact
Every slow `GET /{code}` is a person waiting for a redirect; slow creates hold connections open.

## First five minutes
1. Acknowledge; say in the incident channel that you are looking.
2. Dashboard "RED overview", Duration panel: p50, p95 and p99. Is everything slow (p50 up) or only a
   tail (p99 up)?
3. Which route is slow:

   ```bash
   curl -s --get http://localhost:9090/api/v1/query --data-urlencode \
     'query=histogram_quantile(0.99, sum by (le, http_route) (rate(http_server_request_duration_seconds_bucket{job="shorty"}[5m])))'
   ```

4. Click an exemplar dot near the top of the p99 line: the trace shows which span takes the time.

## Diagnose
- A slow child span (the Kafka publish, a call to another service) points at that dependency.
- No slow child span, but the server span is long: the process itself (CPU, garbage collection,
  lock contention). Check the Go runtime panels and, for Go services, a CPU profile (chapter 6).
- A change in traffic: the Rate panel; more requests than usual can make every request slower.

## Mitigate
Roll back a recent change; reduce load (rate limits, shedding, chapter 16); add capacity if the
service is saturated. Confirm on the Duration panel.

## Escalate
As for `SLOAvailabilityBudgetBurn`: no improvement 15 minutes after mitigation, or several services at
once.

## After
Timeline and postmortem, as for any page.
