#!/usr/bin/env bash
# Where did the delivery lag come from? (milestone M10, chapter 39 section 39.2.) Splits the time from
# ingest to delivery into the pipeline's three stages, averaged over buckets of BUCKET seconds by ingest
# time, for the deliveries to endpoints whose URL contains ENDPOINT:
#   relay     outbox row written -> published to Redpanda
#   dispatch  published -> delivery row created by the dispatcher
#   deliver   delivery row created -> delivered (retries and breaker waits included)
#
#   SINCE='2026-10-05 15:00:00+00' ENDPOINT=chaos-healthy scripts/stage-lag.sh
#
# Read-only. Needs the compose stack (docker).
set -euo pipefail
cd "$(dirname "$0")/.."
SINCE=${SINCE:-1970-01-01 00:00:00+00}
ENDPOINT=${ENDPOINT:-}
BUCKET=${BUCKET:-15}
COMPOSE=${COMPOSE:-docker compose -p hookline -f deploy/compose.yaml}

$COMPOSE exec -T postgres psql -U hookline -d hookline -c "
SELECT to_timestamp(floor(extract(epoch FROM e.created_at) / $BUCKET) * $BUCKET) AS ingested,
       count(*) AS deliveries,
       round(avg(extract(epoch FROM o.published_at - o.created_at))::numeric, 1) AS relay_s,
       round(avg(extract(epoch FROM d.created_at - o.published_at))::numeric, 1) AS dispatch_s,
       round(avg(extract(epoch FROM d.delivered_at - d.created_at))::numeric, 1) AS deliver_s
FROM events e
JOIN outbox o ON o.event_id = e.id
JOIN deliveries d ON d.event_id = e.id
JOIN endpoints ep ON ep.id = d.endpoint_id
WHERE e.created_at >= '$SINCE' AND ep.url LIKE '%' || '$ENDPOINT' || '%' AND d.status = 'delivered'
GROUP BY 1 ORDER BY 1"
