#!/usr/bin/env bash
# Shared by check-telemetry.sh and stage-incident.sh (milestone M8): start the four Hookline binaries on
# this machine, exporting to the collector of deploy/compose.observability.yaml, and create a tenant with
# one endpoint on the compose sink. Source it; do not run it.
# shellcheck shell=bash

OUT=${OUT:-/tmp/hookline-m8}
API=${API:-http://127.0.0.1:18080}
SINK=${SINK:-http://localhost:8091}
COMPOSE="docker compose -p hookline -f deploy/compose.yaml -f deploy/compose.observability.yaml"
mkdir -p "$OUT"

export DATABASE_URL=postgres://hookline:change-me@localhost:5438/hookline
export REDIS_ADDR=localhost:6438 KAFKA_BROKERS=localhost:19438
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 OTEL_METRIC_EXPORT_INTERVAL=5000
export OTEL_RESOURCE_ATTRIBUTES=deployment.environment.name=local
# Dev mode for every binary: the compose sink is plain http on localhost (the SSRF policy, M9).
export HOOKLINE_DEV_MODE=true
export HOOKLINE_BACKOFF_BASE=${HOOKLINE_BACKOFF_BASE:-1s} HOOKLINE_BACKOFF_CAP=${HOOKLINE_BACKOFF_CAP:-30s}

sql() { $COMPOSE exec -T postgres psql -U hookline -d hookline -At -c "$1"; }

declare -A PIDS=()
# start NAME [VAR=value ...]: run one binary in the background, logging to $OUT/NAME.log.
start() {
  local name=$1; shift
  env "$@" "$OUT/bin/$name" >> "$OUT/$name.log" 2>&1 &
  PIDS[$name]=$!
}
stop_one() { kill "${PIDS[$1]}" 2> /dev/null && wait "${PIDS[$1]}" 2> /dev/null; unset "PIDS[$1]"; }
stop_all() { for n in "${!PIDS[@]}"; do stop_one "$n"; done; }

start_hookline() {
  go build -o "$OUT/bin/" ./cmd/...
  "$OUT/bin/hookline-admin" migrate up > /dev/null
  "$OUT/bin/hookline-admin" create-topic -partitions 6 > /dev/null
  start api HOOKLINE_ADDR=127.0.0.1:18080 HOOKLINE_DEBUG_ADDR=off
  start relay
  start dispatcher
  start worker
  for _ in $(seq 50); do curl -sf "$API/healthz" > /dev/null && return 0; sleep 0.2; done
  echo "the API did not start; see $OUT/api.log" >&2
  return 1
}

# new_tenant HOOK [RATE BURST]: prints the API key of a new tenant whose one endpoint (every event type)
# is the compose sink's hook HOOK, verified with the sink's practice secret. The ingest limit is RATE per
# second with bursts of BURST (both 1000 unless given).
new_tenant() {
  local key id
  key=$("$OUT/bin/hookline-admin" create-tenant -name "m8-$1-$(date +%s)" -rate "${2:-1000}" -burst "${3:-${2:-1000}}" |
    sed -n 's/^api key: *//p')
  id=$(curl -sf "$API/v1/endpoints" -H "Authorization: Bearer $key" -H 'Content-Type: application/json' \
    -d "{\"url\":\"$SINK/hooks/$1\"}" | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
  sql "UPDATE endpoints SET secret = 'whsec_change-me' WHERE id = '$id'" > /dev/null
  curl -sf "$API/v1/subscriptions" -H "Authorization: Bearer $key" -H 'Content-Type: application/json' \
    -d "{\"endpoint_id\":\"$id\",\"event_type\":\"*\"}" > /dev/null
  echo "$key"
}

# send KEY IDEMPOTENCY_KEY [TRACEPARENT]: one event; prints its id.
send() {
  local tp=()
  if [ -n "${3:-}" ]; then tp=(-H "traceparent: $3"); fi
  curl -sf "$API/v1/events" -H "Authorization: Bearer $1" -H 'Content-Type: application/json' \
    -H "Idempotency-Key: $2" "${tp[@]}" -d '{"type":"m8.check","payload":{"hello":"telemetry"}}' |
    python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])'
}
