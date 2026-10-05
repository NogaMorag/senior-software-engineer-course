#!/usr/bin/env bash
# Milestone M8: checks the telemetry of a running Hookline end to end through the backends' own APIs (the
# pattern of chapter 24's tests/check_stack.sh). Starts the four binaries with OTLP export, sends one
# event with a trace ID chosen here, and asks Tempo for the whole path API -> relay -> Redpanda ->
# dispatcher -> worker -> sink, Prometheus for Hookline's series, Loki for the trace's log lines, and
# Grafana for the two dashboards.
#   make obs-up && scripts/check-telemetry.sh
# shellcheck disable=SC2329 # the check functions are called through retry
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1
# shellcheck source=scripts/lib-local.sh
. scripts/lib-local.sh
PROM=${PROM:-http://localhost:9090}
LOKI=${LOKI:-http://localhost:3100}
TEMPO=${TEMPO:-http://localhost:3200}
GRAFANA=${GRAFANA:-http://localhost:3000}
WAIT=${WAIT:-90}
trap stop_all EXIT

failures=0
pass() { printf 'PASS  %s\n' "$1"; }
fail() { printf 'FAIL  %s\n' "$1"; failures=$((failures + 1)); }
json() { python3 -c "import json,sys; d=json.load(sys.stdin); print($1)"; }
retry() {
  local deadline=$((SECONDS + $1)); shift
  until "$@"; do
    [ "$SECONDS" -ge "$deadline" ] && return 1
    sleep 3
  done
}
up() { curl -fs -o /dev/null --max-time 3 "$1"; }

echo "== 1. the backends answer"
for url in "$PROM/-/ready" "$LOKI/ready" "$TEMPO/ready" "$GRAFANA/api/health" "http://localhost:13133/" "$SINK/stats"; do
  if retry "$WAIT" up "$url"; then pass "$url"; else fail "$url does not answer"; fi
done

echo "== 2. Hookline, and one event in a trace chosen here"
start_hookline || exit 1
KEY=$(new_tenant telemetry)
TRACE_ID=$(python3 -c 'import secrets; print(secrets.token_hex(16))')
EVENT=$(send "$KEY" "check-$TRACE_ID" "00-$TRACE_ID-$(python3 -c 'import secrets; print(secrets.token_hex(8))')-01")
if [ -n "$EVENT" ]; then pass "POST /v1/events -> event $EVENT in trace $TRACE_ID"; else fail "ingest failed"; fi
for i in $(seq 1 20); do send "$KEY" "more-$TRACE_ID-$i" > /dev/null; done
pass "20 more events"

echo "== 3. Tempo has the trace, across every hop"
trace_summary() {
  curl -fsS "$TEMPO/api/traces/$TRACE_ID" | python3 -c '
import json, sys
d = json.load(sys.stdin)
for b in d.get("batches", d.get("resourceSpans", [])):
    svc = next((a["value"]["stringValue"] for a in b["resource"]["attributes"] if a["key"] == "service.name"), "?")
    for ss in b.get("scopeSpans", b.get("instrumentationLibrarySpans", [])):
        for s in ss["spans"]:
            print(svc + ":" + s["name"])'
}
has_path() {
  local s; s=$(trace_summary 2> /dev/null) || return 1
  grep -q '^hookline-api:POST /v1/events$' <<< "$s" && grep -q '^hookline-relay:hookline.events publish$' <<< "$s" &&
    grep -q '^hookline-dispatcher:hookline.events process$' <<< "$s" && grep -q '^hookline-worker:hookline.deliver$' <<< "$s" &&
    grep -q '^hookline-sink:POST /hooks/{name}$' <<< "$s"
}
if retry "$WAIT" has_path; then
  pass "Tempo: trace $TRACE_ID has $(trace_summary | wc -l | tr -d ' ') spans:"
  trace_summary | sort | uniq -c | sed 's/^/        /'
else
  fail "Tempo: trace $TRACE_ID is missing a hop; got: $(trace_summary 2> /dev/null | tr '\n' ' ')"
fi

echo "== 4. Prometheus has Hookline's series and rules"
prom_value() {
  curl -fsS --get "$PROM/api/v1/query" --data-urlencode "query=$1" |
    json "d['data']['result'][0]['value'][1] if d['data']['result'] else 'none'"
}
has_series() { [ "$(prom_value "$1")" != none ]; }
for q in \
  'sum(http_server_request_duration_seconds_count{job="hookline-api"})' \
  'sum(hookline_ingest_events_total)' \
  'max(hookline_outbox_pending)' \
  'sum(hookline_dispatch_deliveries_total)' \
  'sum(hookline_delivery_attempts_total{outcome="delivered"})' \
  'sum(hookline_delivery_lag_seconds_count)' \
  'sum(http_server_request_duration_seconds_count{job="hookline-sink"})' \
  'job:slo_late_deliveries:ratio_rate5m' \
  'job:slo_delivery:objective'; do
  if retry "$WAIT" has_series "$q"; then pass "$q = $(prom_value "$q")"; else fail "no data for $q"; fi
done

echo "== 5. Loki has the log lines of that trace"
loki_lines() {
  curl -fsS --get "$LOKI/loki/api/v1/query_range" \
    --data-urlencode "query={service_name=~\"hookline-.+\"} | trace_id=\"$TRACE_ID\"" \
    --data-urlencode "start=$(($(date +%s) - 600))000000000" --data-urlencode "limit=100" |
    json "sum(len(s['values']) for s in d['data']['result'])"
}
has_lines() { [ "$(loki_lines)" -ge 1 ]; }
if retry "$WAIT" has_lines; then pass "Loki: $(loki_lines) lines carry trace_id=$TRACE_ID"; else fail "Loki: no lines for the trace"; fi

echo "== 6. Grafana loaded the dashboards"
for uid in hookline-api hookline-pipeline; do
  if curl -fs "$GRAFANA/api/dashboards/uid/$uid" > /dev/null; then pass "dashboard $uid"; else fail "dashboard $uid missing"; fi
done

echo
if [ "$failures" -eq 0 ]; then echo "ALL CHECKS PASSED"; else echo "$failures CHECK(S) FAILED"; fi
exit $((failures > 0))
