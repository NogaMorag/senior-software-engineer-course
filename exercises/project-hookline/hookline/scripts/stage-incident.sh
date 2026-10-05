#!/usr/bin/env bash
# Milestone M8: stages the deliberately broken release and prints a timeline for the postmortem
# (the pattern of chapter 25's data/stage_incident.sh). Steady traffic with a healthy worker, then a
# worker "release" whose attempt timeout is 1ms instead of 1s (a unit typo), then the rollback. Every
# delivery attempt times out; ingest stays green; DeliveryAttemptsFailing should page, and after the
# rollback the backlog arrives late, which the delivery SLO counts.
#   make obs-up && scripts/stage-incident.sh
# Defaults: 2 minutes of baseline, the bad release until the page fires (at most 10 minutes), then the
# rollback and up to 10 minutes of watching. Tune with BASELINE, MAX_BAD, AFTER (seconds).
# shellcheck disable=SC2329 # paged and recovered are called through watch_alerts
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1
# shellcheck source=scripts/lib-local.sh
. scripts/lib-local.sh
PROM=${PROM:-http://localhost:9090}
BASELINE=${BASELINE:-120}
MAX_BAD=${MAX_BAD:-600}
AFTER=${AFTER:-600}
BAD_TIMEOUT=${BAD_TIMEOUT:-1ms}
trap 'kill "$LOAD" 2> /dev/null; stop_all' EXIT

now() { date -u +%H:%M:%S; }
say() { printf '%s  %s\n' "$(now)" "$*"; }
alert_state() { # alert_state NAME: firing, pending or none
  curl -fsS "$PROM/api/v1/alerts" | python3 -c "
import json, sys
a = [x for x in json.load(sys.stdin)['data']['alerts'] if x['labels']['alertname'] == '$1']
print(a[0]['state'] if a else 'none')"
}
q() { # q QUERY: the first value, rounded, or n/a
  curl -fsS --get "$PROM/api/v1/query" --data-urlencode "query=$1" |
    python3 -c "import json,sys; r=json.load(sys.stdin)['data']['result']; print(f'{float(r[0][\"value\"][1]):.3f}' if r else 'n/a')"
}
status() { say "$1 (attempt failure share $(q 'job:delivery_attempt_failures:ratio_rate5m') | late share $(q 'job:slo_late_deliveries:ratio_rate5m') | outbox oldest $(q 'max(hookline_outbox_oldest_age_seconds)')s)"; }
watch_alerts() { # watch_alerts SECONDS CHECK: print every state change of the two alerts; stop when CHECK succeeds
  local deadline=$((SECONDS + $1)) last=""
  while [ "$SECONDS" -lt "$deadline" ]; do
    local cur
    cur="DeliveryAttemptsFailing=$(alert_state DeliveryAttemptsFailing) DeliveryLagBudgetBurn=$(alert_state DeliveryLagBudgetBurn)"
    if [ "$cur" != "$last" ]; then status "$cur"; last=$cur; fi
    if "$2"; then return 0; fi
    sleep 10
  done
  return 1
}

paged() { [ "$(alert_state DeliveryAttemptsFailing)" = firing ]; }
recovered() {
  [ "$(alert_state DeliveryAttemptsFailing)" = none ] && [ "$(alert_state DeliveryLagBudgetBurn)" = none ] &&
    [ "$(sql "SELECT count(*) FROM deliveries d JOIN tenants t ON t.id = d.tenant_id
              WHERE t.name LIKE 'm8-incident-%' AND d.status <> 'delivered'")" = 0 ]
}

start_hookline || exit 1
KEY=$(new_tenant incident)
say "traffic starts: about 5 events per second"
( i=0; while true; do i=$((i + 1)); send "$KEY" "incident-$$-$i" > /dev/null; sleep 0.2; done ) &
LOAD=$!
sleep "$BASELINE"
status "baseline done"

say "CHANGE: worker redeployed with HOOKLINE_ATTEMPT_TIMEOUT=$BAD_TIMEOUT"
stop_one worker
start worker HOOKLINE_ATTEMPT_TIMEOUT="$BAD_TIMEOUT"
if watch_alerts "$MAX_BAD" paged; then
  say "PAGE: DeliveryAttemptsFailing is firing"
else
  say "the page did not fire within $MAX_BAD s"
fi

say "MITIGATION: worker rolled back to the default attempt timeout"
stop_one worker
start worker
watch_alerts "$AFTER" recovered ||
  say "still recovering after $AFTER s"
kill "$LOAD" 2> /dev/null

echo
echo "== numbers for the postmortem (this run)"
sql "SELECT 'deliveries=' || count(*) || ' delivered=' || count(*) FILTER (WHERE status = 'delivered') ||
            ' dead=' || count(*) FILTER (WHERE status = 'dead') ||
            ' late_over_30s=' || count(*) FILTER (WHERE d.delivered_at - e.created_at > interval '30 seconds') ||
            ' max_lag_s=' || round(max(extract(epoch FROM d.delivered_at - e.created_at))::numeric, 1) ||
            ' attempts=' || sum(d.attempts)
     FROM deliveries d JOIN events e ON e.id = d.event_id
     JOIN tenants t ON t.id = e.tenant_id WHERE t.name LIKE 'm8-incident-%'"
say "done; logs in $OUT"
