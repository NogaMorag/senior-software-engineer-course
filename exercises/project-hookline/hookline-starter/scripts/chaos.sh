#!/usr/bin/env bash
# The chaos run of spec N3 and N4 (milestone M10, chapter 39): k6 sends events at RATE per second for
# DURATION seconds while this script kills a relay, a dispatcher or a worker with SIGKILL at random
# moments and restarts it, and the second sink misbehaves in phases (failing, then hanging, then flaky).
# Afterwards it waits for the backlog and runs scripts/verify-delivery.sh, which must report 0 missing.
#
#   scripts/chaos.sh                        # RATE=50, DURATION=120, a kill every 3 to 12 s
#   RATE=100 DURATION=300 SEED=42 scripts/chaos.sh
#
# Two instances of each pipeline binary run, so a kill is also a failover: the relays elect a leader, the
# dispatchers rebalance the consumer group, and a dead worker's leases expire. The API is not killed: an
# event it never accepted is not one Hookline owes anyone. One tenant has two endpoints: chaos-healthy on
# the sink at :8091 (its lag is N3's number) and chaos-flaky on the sink at :8092 (the misbehaving one).
# The timeline, k6's summary, a resource sample and the verification land in $OUT. Needs: docker, go,
# curl, python3, k6 v2. Leaves the stack up unless DOWN=1.
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1
OUT=${OUT:-/tmp/hookline-chaos}
RATE=${RATE:-50}
DURATION=${DURATION:-120}
KILL_EVERY=${KILL_EVERY:-10}
SETTLE=${SETTLE:-300}
SEED=${SEED:-$$}
RANDOM=$SEED
# Short timeouts and leases, so a killed worker's deliveries come back within seconds, and a breaker
# that probes again soon; the lease must stay longer than the attempt timeout.
export HOOKLINE_ATTEMPT_TIMEOUT=${HOOKLINE_ATTEMPT_TIMEOUT:-2s} HOOKLINE_LEASE=${HOOKLINE_LEASE:-10s}
export HOOKLINE_MAX_ATTEMPTS=${HOOKLINE_MAX_ATTEMPTS:-50} HOOKLINE_BREAKER_OPEN=${HOOKLINE_BREAKER_OPEN:-5s}
export HOOKLINE_BACKOFF_BASE=${HOOKLINE_BACKOFF_BASE:-500ms} HOOKLINE_BACKOFF_CAP=${HOOKLINE_BACKOFF_CAP:-5s}
# shellcheck source=scripts/lib-local.sh
. scripts/lib-local.sh
SLOW=http://localhost:8092
VICTIMS=(relay-1 relay-2 dispatcher-1 dispatcher-2 worker-1 worker-2)

command -v k6 > /dev/null || { echo "k6 is not on PATH (chapter 26 installs it)" >&2; exit 1; }
rm -f "$OUT"/*.log "$OUT/timeline.txt"
log() { printf '%s  %s\n' "$(date +%T)" "$*" | tee -a "$OUT/timeline.txt"; }
# spawn NAME: one more instance of the binary NAME names (relay-2 runs bin/relay), logging to NAME.log.
spawn() {
  "$OUT/bin/${1%-*}" >> "$OUT/$1.log" 2>&1 &
  PIDS[$1]=$!
}
control() { curl -sf -X POST "$SLOW/control" -H 'Content-Type: application/json' -d "$1" > /dev/null; }
# shellcheck disable=SC2329 # called by the EXIT trap
cleanup() {
  stop_all
  if [ "${DOWN:-0}" = 1 ]; then $COMPOSE down -v > /dev/null 2>&1; fi
}
trap cleanup EXIT

log "seed $SEED, rate $RATE/s for ${DURATION}s, a kill every 3 to $((KILL_EVERY + 2)) s"
$COMPOSE up -d --wait --build > "$OUT/compose.log" 2>&1 || { echo "compose failed; see $OUT/compose.log" >&2; exit 1; }
go build -o "$OUT/bin/" ./cmd/... || exit 1
"$OUT/bin/hookline-admin" migrate up > /dev/null
"$OUT/bin/hookline-admin" create-topic -partitions 6 > /dev/null
start api HOOKLINE_ADDR=127.0.0.1:18080 HOOKLINE_DEBUG_ADDR=off
for v in "${VICTIMS[@]}"; do spawn "$v"; done
for _ in $(seq 50); do curl -sf "$API/healthz" > /dev/null && break; sleep 0.2; done

for s in "$SINK" "$SLOW"; do curl -sf -X POST "$s/reset" > /dev/null; done
KEY=$(new_tenant chaos-healthy) || exit 1
FLAKY=$(curl -sf "$API/v1/endpoints" -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' \
  -d "{\"url\":\"$SLOW/hooks/chaos-flaky\"}" | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
sql "UPDATE endpoints SET secret = 'whsec_change-me' WHERE id = '$FLAKY'" > /dev/null
curl -sf "$API/v1/subscriptions" -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' \
  -d "{\"endpoint_id\":\"$FLAKY\",\"event_type\":\"*\"}" > /dev/null
SINCE=$(sql "SELECT now()")

log "load starts"
k6 run -e BASE_URL="$API" -e API_KEY="$KEY" -e RATE="$RATE" -e DURATION="${DURATION}s" \
  --summary-export "$OUT/k6-summary.json" loadtest/ingest.js > "$OUT/k6.txt" 2>&1 &
K6=$!
begin=$SECONDS
phase=0
kills=0
sampled=0
while kill -0 "$K6" 2> /dev/null; do
  t=$((SECONDS - begin))
  # The flaky sink's phases, by elapsed time: 503 half of the time, then hanging, then 503 a fifth.
  if [ "$phase" = 0 ] && [ "$t" -ge 10 ]; then control '{"failure_rate":0.5}'; log "SINK  chaos-flaky fails 50%"; phase=1; fi
  if [ "$phase" = 1 ] && [ "$t" -ge $((DURATION / 3)) ]; then control '{"failure_rate":0,"timeout_mode":true}'; log "SINK  chaos-flaky hangs"; phase=2; fi
  if [ "$phase" = 2 ] && [ "$t" -ge $((2 * DURATION / 3)) ]; then control '{"failure_rate":0.2,"timeout_mode":false}'; log "SINK  chaos-flaky fails 20%"; phase=3; fi
  if [ "$sampled" = 0 ] && [ "$t" -ge $((DURATION / 2)) ]; then
    { date +%T; docker stats --no-stream --format '{{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}' | grep hookline
      for n in "${!PIDS[@]}"; do printf '%s\t%s\n' "$n" "$(ps -o pcpu=,rss= -p "${PIDS[$n]}")"; done
    } > "$OUT/resources.txt" 2>&1
    sampled=1
  fi
  sleep $((RANDOM % KILL_EVERY + 3))
  kill -0 "$K6" 2> /dev/null || break
  v=${VICTIMS[RANDOM % ${#VICTIMS[@]}]}
  kill -9 "${PIDS[$v]}" 2> /dev/null
  wait "${PIDS[$v]}" 2> /dev/null
  log "KILL  $v"
  kills=$((kills + 1))
  sleep $((RANDOM % 4 + 1))
  spawn "$v"
  log "START $v"
done
wait "$K6"
k6_status=$?
control '{"failure_rate":0,"timeout_mode":false,"latency_ms":0}'
log "load done (k6 exit $k6_status, $kills kills); chaos-flaky healthy again; waiting up to ${SETTLE}s"

SINCE="$SINCE" WAIT="$SETTLE" OUT="$OUT/verify" COMPOSE="$COMPOSE" scripts/verify-delivery.sh | tee "$OUT/verify.txt"
verify_status=${PIPESTATUS[0]}
sql "SELECT status, count(*) FROM deliveries d JOIN events e ON e.id = d.event_id
     WHERE e.created_at >= '$SINCE' GROUP BY status ORDER BY status" | sed 's/^/deliveries  /'
grep -E "http_req_duration\{phase:measure|dropped_iterations\{|http_reqs\.\.\." "$OUT/k6.txt" | sed 's/^ */k6  /'
log "verify exit $verify_status"
exit "$verify_status"
