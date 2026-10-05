#!/usr/bin/env bash
# The isolation lab of spec N3 (milestone M6). Two receivers: "healthy" answers at once, "slow" holds
# every request until the worker's attempt timeout. Both are subscribed to every event. The lab sends
# EVENTS events and prints the delivery lag of the healthy endpoint: the time from ingest to delivery.
#
#   scripts/lab-n3.sh                 # with the per-endpoint breaker and bulkhead
#   GUARDS=off scripts/lab-n3.sh      # the same run without them, for comparison
#
# The numbers depend on your machine. Record yours, with the date and the machine, in docs/EVALUATION.md.
# Needs: docker, go, curl, python3. Leaves the stack up unless DOWN=1.
set -euo pipefail
cd "$(dirname "$0")/.."

EVENTS=${EVENTS:-300}
GUARDS=${GUARDS:-on}
WORKERS=${WORKERS:-8}
OUT=${OUT:-/tmp/hookline-lab-n3}
API=http://127.0.0.1:18080
COMPOSE="docker compose -p hookline -f deploy/compose.yaml"
mkdir -p "$OUT"

export DATABASE_URL=postgres://hookline:change-me@localhost:5438/hookline
export REDIS_ADDR=localhost:6438 KAFKA_BROKERS=localhost:19438
export HOOKLINE_WORKERS=$WORKERS HOOKLINE_ATTEMPT_TIMEOUT=${ATTEMPT_TIMEOUT:-5s} HOOKLINE_LEASE=30s
export HOOKLINE_BACKOFF_BASE=500ms HOOKLINE_BACKOFF_CAP=5s HOOKLINE_MAX_ATTEMPTS=20
if [ "$GUARDS" = off ]; then
  # A bulkhead as wide as the pool and a breaker that never opens: no protection at all.
  export HOOKLINE_ENDPOINT_CONCURRENCY=$WORKERS HOOKLINE_BREAKER_FAILURES=1000000
else
  export HOOKLINE_ENDPOINT_CONCURRENCY=${CONCURRENCY:-2} HOOKLINE_BREAKER_FAILURES=5 HOOKLINE_BREAKER_OPEN=10s
fi

sql() { $COMPOSE exec -T postgres psql -U hookline -d hookline -At -c "$1"; }

echo "== stack and binaries"
$COMPOSE up -d --wait --build > "$OUT/compose.log" 2>&1
go build -o "$OUT/bin/" ./cmd/...
"$OUT/bin/hookline-admin" migrate up > /dev/null
"$OUT/bin/hookline-admin" create-topic -partitions 6 > /dev/null
KEY=$("$OUT/bin/hookline-admin" create-tenant -name "lab-n3-$(date +%s)" -rate 1000 -burst 1000 | sed -n 's/^api key: *//p')

pids=()
cleanup() {
  kill "${pids[@]}" 2> /dev/null || true
  wait 2> /dev/null || true
  if [ "${DOWN:-0}" = 1 ]; then $COMPOSE down -v > /dev/null 2>&1; fi
}
trap cleanup EXIT
# Dev mode: the lab receivers are plain http on localhost.
HOOKLINE_DEV_MODE=true HOOKLINE_ADDR=127.0.0.1:18080 HOOKLINE_DEBUG_ADDR=off "$OUT/bin/api" > "$OUT/api.log" 2>&1 & pids+=($!)
"$OUT/bin/relay" > "$OUT/relay.log" 2>&1 & pids+=($!)
"$OUT/bin/dispatcher" > "$OUT/dispatcher.log" 2>&1 & pids+=($!)
"$OUT/bin/worker" > "$OUT/worker.log" 2>&1 & pids+=($!)
for _ in $(seq 50); do curl -sf "$API/healthz" > /dev/null && break; sleep 0.2; done
curl -sf "$API/healthz" > /dev/null || { echo "the API did not start; see $OUT/api.log" >&2; exit 1; }

call() { curl -sf "$API$1" -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' "${@:2}"; }
endpoint() { # url -> id; gives the endpoint the lab sinks' secret, which the API did not generate
  local id
  id=$(call /v1/endpoints -d "{\"url\":\"$1\"}" | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])') ||
    { echo "creating the endpoint $1 failed; see $OUT/api.log" >&2; exit 1; }
  sql "UPDATE endpoints SET secret = 'whsec_change-me' WHERE id = '$id'" > /dev/null
  call /v1/subscriptions -d "{\"endpoint_id\":\"$id\",\"event_type\":\"*\"}" > /dev/null
  echo "$id"
}
curl -sf -X POST http://localhost:8091/reset > /dev/null
curl -sf -X POST http://localhost:8092/reset > /dev/null
curl -sf -X POST http://localhost:8092/control -H 'Content-Type: application/json' -d '{"timeout_mode":true}' > /dev/null
HEALTHY=$(endpoint http://localhost:8091/hooks/healthy)
SLOW=$(endpoint http://localhost:8092/hooks/slow)

echo "== sending $EVENTS events (guards: $GUARDS, workers: $WORKERS, endpoint concurrency: $HOOKLINE_ENDPOINT_CONCURRENCY)"
for i in $(seq "$EVENTS"); do
  call /v1/events -H "Idempotency-Key: lab-$i-$RANDOM$RANDOM" -d "{\"type\":\"lab.tick\",\"payload\":{\"n\":$i}}" > /dev/null
done

echo "== waiting for the healthy endpoint (at most ${WAIT:-180}s)"
for _ in $(seq "${WAIT:-180}"); do
  done_n=$(sql "SELECT count(*) FROM deliveries WHERE endpoint_id = '$HEALTHY' AND status = 'delivered'")
  [ "$done_n" -ge "$EVENTS" ] && break
  sleep 1
done

echo "== delivery lag of the healthy endpoint, seconds from ingest to delivery"
sql "SELECT 'delivered=' || count(*) ||
            '  p50=' || round(percentile_cont(0.50) WITHIN GROUP (ORDER BY lag)::numeric, 3) ||
            '  p95=' || round(percentile_cont(0.95) WITHIN GROUP (ORDER BY lag)::numeric, 3) ||
            '  p99=' || round(percentile_cont(0.99) WITHIN GROUP (ORDER BY lag)::numeric, 3) ||
            '  max=' || round(max(lag)::numeric, 3)
     FROM (SELECT extract(epoch FROM d.delivered_at - e.created_at) AS lag
           FROM deliveries d JOIN events e ON e.id = d.event_id
           WHERE d.endpoint_id = '$HEALTHY' AND d.status = 'delivered') x"
echo "== the slow endpoint: deliveries by status, attempts spent"
sql "SELECT status || '=' || count(*) || ' (attempts ' || sum(attempts) || ')' FROM deliveries
     WHERE endpoint_id = '$SLOW' GROUP BY status ORDER BY status"
echo "== breaker transitions in the worker log"
grep -c '"msg":"endpoint breaker"' "$OUT/worker.log" || true
echo "logs: $OUT"
