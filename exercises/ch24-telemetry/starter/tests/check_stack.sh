#!/usr/bin/env bash
# Checks the running chapter 24 stack end to end, through the backends' own
# HTTP APIs: sends traffic with a trace ID it chose, then asks Prometheus for
# the series, Loki for the log lines of that trace, and Tempo for the trace,
# which must cross the Kafka hop (shorty, linkworker, receiver).
#
#   bash tests/check_stack.sh            (with the stack up; see the README)
#
# Exits 0 when every check passes. Needs bash, curl and python3 (for JSON).
set -uo pipefail

SHORTY=${SHORTY:-http://localhost:8080}
PROM=${PROM:-http://localhost:9090}
LOKI=${LOKI:-http://localhost:3100}
TEMPO=${TEMPO:-http://localhost:3200}
GRAFANA=${GRAFANA:-http://localhost:3000}
COLLECTOR_HEALTH=${COLLECTOR_HEALTH:-http://localhost:13133/}
WAIT=${WAIT:-90} # seconds to wait for data to arrive in each backend

failures=0
pass() { printf 'PASS  %s\n' "$1"; }
fail() { printf 'FAIL  %s\n' "$1"; failures=$((failures + 1)); }

# json EXPR: read JSON on stdin, print the Python expression EXPR over it (d).
json() { python3 -c "import json,sys; d=json.load(sys.stdin); print($1)"; }

# retry SECONDS CMD...: run CMD every 3 s until it succeeds or time runs out.
retry() {
  local deadline=$((SECONDS + $1)); shift
  until "$@"; do
    [ "$SECONDS" -ge "$deadline" ] && return 1
    sleep 3
  done
}

up() { curl -fs -o /dev/null --max-time 3 "$1"; }

echo "== 1. everything answers"
for url in "$SHORTY/healthz" "$PROM/-/ready" "$LOKI/ready" "$TEMPO/ready" "$GRAFANA/api/health" "$COLLECTOR_HEALTH"; do
  if retry "$WAIT" up "$url"; then pass "$url"; else fail "$url does not answer"; fi
done

echo "== 2. traffic with a trace ID chosen here"
# A W3C traceparent: version 00, 16-byte trace ID, 8-byte parent span ID, flags 01 (sampled).
TRACE_ID=$(python3 -c 'import secrets; print(secrets.token_hex(16))')
CODE="chk$(python3 -c 'import secrets; print(secrets.token_hex(4))')"
status=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$SHORTY/v1/links" \
  -H "traceparent: 00-$TRACE_ID-$(python3 -c 'import secrets; print(secrets.token_hex(8))')-01" \
  -H 'Content-Type: application/json' -d "{\"url\": \"https://example.com/check\", \"code\": \"$CODE\"}")
if [ "$status" = 201 ]; then pass "POST /v1/links -> 201 in trace $TRACE_ID"; else fail "POST /v1/links -> $status"; fi
for _ in $(seq 1 20); do
  curl -s -o /dev/null "$SHORTY/$CODE"            # 302
  curl -s -o /dev/null "$SHORTY/no-such-code-$RANDOM" # 404
done
pass "40 more requests sent (resolves and misses)"

echo "== 3. Prometheus has the series"
prom_value() { # prom_value QUERY: the value of the first sample, or 0
  curl -fsS --get "$PROM/api/v1/query" --data-urlencode "query=$1" |
    json "d['data']['result'][0]['value'][1] if d['data']['result'] else 0"
}
has_series() { [ "$(prom_value "$1")" != 0 ]; }
for q in \
  'sum(http_server_request_duration_seconds_count{job="shorty"})' \
  'sum(http_server_request_duration_seconds_count{job="receiver"})' \
  'sum(linkworker_deliveries_total{outcome="ok"})' \
  'sum(go_goroutine_count{job="shorty"})'; do
  if retry "$WAIT" has_series "$q"; then pass "$q = $(prom_value "$q")"; else fail "no data for $q"; fi
done
has_exemplars() {
  curl -fsS --get "$PROM/api/v1/query_exemplars" --data-urlencode 'query=http_server_request_duration_seconds_bucket{job="shorty"}' \
    --data-urlencode "start=$(($(date +%s) - 600))" --data-urlencode "end=$(date +%s)" |
    json "sum(len(s['exemplars']) for s in d['data'])" | grep -qv '^0$'
}
if retry "$WAIT" has_exemplars; then pass "exemplars (trace IDs on histogram samples) are stored"; else fail "no exemplars"; fi

echo "== 4. Loki has the log lines of that trace"
loki_lines() {
  curl -fsS --get "$LOKI/loki/api/v1/query_range" \
    --data-urlencode "query={service_name=~\"shorty|linkworker|receiver\"} | trace_id=\"$TRACE_ID\"" \
    --data-urlencode "start=$(($(date +%s) - 600))000000000" --data-urlencode "limit=100" |
    json "sum(len(s['values']) for s in d['data']['result'])"
}
has_lines() { [ "$(loki_lines)" -ge 3 ]; } # shorty's two lines, linkworker's, the receiver's
if retry "$WAIT" has_lines; then pass "Loki: $(loki_lines) lines carry trace_id=$TRACE_ID"; else fail "Loki: $(loki_lines) lines for the trace, want at least 3"; fi

echo "== 5. Tempo has the trace, across the Kafka hop"
trace_summary() { # prints "service:span" pairs, one per line
  curl -fsS "$TEMPO/api/traces/$TRACE_ID" | python3 -c '
import json, sys
d = json.load(sys.stdin)
for b in d.get("batches", d.get("resourceSpans", [])):
    svc = next((a["value"]["stringValue"] for a in b["resource"]["attributes"] if a["key"] == "service.name"), "?")
    for ss in b.get("scopeSpans", b.get("instrumentationLibrarySpans", [])):
        for s in ss["spans"]:
            print(svc + ":" + s["name"])'
}
has_hop() {
  local s; s=$(trace_summary 2>/dev/null) || return 1
  grep -q '^shorty:links.created publish$' <<<"$s" && grep -q '^linkworker:links.created process$' <<<"$s" &&
    grep -q '^receiver:POST /hooks/{name}$' <<<"$s"
}
if retry "$WAIT" has_hop; then
  pass "Tempo: trace $TRACE_ID has $(trace_summary | wc -l | tr -d ' ') spans:"
  trace_summary | sort | uniq -c | sed 's/^/        /'
else
  fail "Tempo: trace $TRACE_ID is missing the hop; got: $(trace_summary 2>/dev/null | tr '\n' ' ')"
fi

echo "== 6. Grafana loaded the provisioned files"
for uid in prometheus loki tempo; do
  if curl -fs "$GRAFANA/api/datasources/uid/$uid" >/dev/null; then pass "datasource $uid"; else fail "datasource $uid missing"; fi
done
title=$(curl -fs "$GRAFANA/api/dashboards/uid/red-overview" | json "d['dashboard']['title']" 2>/dev/null)
if [ "$title" = "RED overview" ]; then pass "dashboard red-overview"; else fail "dashboard red-overview missing"; fi

echo
if [ "$failures" -eq 0 ]; then echo "ALL CHECKS PASSED"; else echo "$failures CHECK(S) FAILED"; fi
exit $((failures > 0))
