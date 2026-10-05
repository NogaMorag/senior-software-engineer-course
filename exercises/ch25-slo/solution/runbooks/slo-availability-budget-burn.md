# Runbook: SLOAvailabilityBudgetBurn

| Field | Value |
| --- | --- |
| Alert | `SLOAvailabilityBudgetBurn` (`severity=page`) |
| Service | the `job` label of the alert (for chapter 25: `shorty`) |
| SLO | `job:slo_availability:objective` of the requests are not answered 5xx, over 30 days (shorty: 99.5%) |
| Rules | `rules/slo-recording.yml`, `rules/slo-objectives.yml`, `rules/slo-alerts.yml` |
| Owner | you (chapter 25) |
| Last reviewed | `<date>` |

## What it means
Users are getting 5xx answers at more than 6 times the rate the SLO allows, measured over both a long
and a short window (1h and 5m at burn rate 14.4, or 6h and 30m at burn rate 6). At burn rate 14.4 the
whole 30-day budget is gone in about 2 days; 2% of it went in the last hour. The short window means it
is still happening now.

## Impact
For shorty: creates, resolves or stats fail with 5xx. A failing `POST /v1/links` means a user cannot
create a link; a failing `GET /{code}` means a link someone clicked does not redirect.

## First five minutes
1. Acknowledge the page and say in the incident channel that you are looking.
2. Open the "RED overview" dashboard in Grafana (http://localhost:3000, folder Course), service =
   the alert's `job`. Read the Errors panel (normal: 0) and the Status codes panel: which code?
3. Find which route fails:

   ```bash
   curl -s --get http://localhost:9090/api/v1/query --data-urlencode \
     'query=sum by (http_route, http_response_status_code) (rate(http_server_request_duration_seconds_count{job="shorty", http_response_status_code=~"5.."}[5m]))'
   ```

4. Read the error logs of the last 15 minutes in Grafana, Explore, Loki:
   `{service_name="shorty"} | detected_level="error"`. Open a line: the `exception_message` field
   (the slog `error` attribute, as the OpenTelemetry bridge stores it) names the cause.
5. Open one failing request's trace: from a log line, "Open trace"; from the Duration panel, an
   exemplar dot. The red span is where it failed.

## Diagnose
- **Did something change?** `docker compose -p ch24 -f data/compose.yaml ps` shows when each container
  was created; in a real system, the deploy log and the config history. Most incidents follow a change.
- **One route or all?** One route (only `POST /v1/links`) points at what that route alone depends on
  (for shorty, the Kafka publish). All routes point at the process or the host.
- **What does the error say?** `publish to <topic>: ...` names the topic: compare it with the topics
  that exist (`rpk topic list` in the redpanda container). `UNKNOWN_TOPIC_OR_PARTITION`, or a
  `context deadline exceeded` for a topic name you do not recognise, is a configuration problem; a
  deadline on the right topic is a slow or unreachable broker.

## Mitigate
- **A recent change:** roll it back first, investigate after. For the staged incident of chapter 25:

  ```bash
  docker compose -p ch24 -f data/compose.yaml up -d --no-deps shorty   # without SHORTY_KAFKA_TOPIC: the default topic
  ```

- **The broker is down:** shorty cannot announce links; there is no safe degraded mode in this version
  (the outbox of chapter 13 is the fix). Restart the broker container and watch the Errors panel.
- **It worked when:** the Errors panel returns to 0 and, within about 5 minutes, the alert resolves
  (the 5m window falls below the threshold first; that is what the short window is for).

## Escalate
If the error ratio has not started to fall 15 minutes after a rollback, or if more than one service
alerts at once, call in the owner of the shared dependency (the broker, the network).

## After
Keep the timeline (times of alert, acknowledgement, mitigation, recovery) and write the postmortem
from `postmortems/TEMPLATE.md` within a few days.
