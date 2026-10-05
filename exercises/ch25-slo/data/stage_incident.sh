#!/usr/bin/env bash
# Stages the chapter 25 incident on the chapter 24 stack and prints a timeline
# for the postmortem: steady traffic, then a "config change" that points
# shorty at a topic that does not exist, then the rollback.
#
#   export CH24=~/senior-course/ch24 RULES_DIR=~/senior-course/ch25/rules
#   bash data/stage_incident.sh
#
# The stack must already be up with RULES_DIR set (see the README). Defaults:
# 5 minutes of baseline, the bad config until the page fires (at most 15
# minutes), then the rollback and up to 15 minutes for the alert to resolve.
set -uo pipefail

: "${CH24:?set CH24 to the folder that holds data/compose.yaml of chapter 24}"
: "${RULES_DIR:?set RULES_DIR to the chapter 25 rules folder (the stack must use it too)}"
BASELINE=${BASELINE:-300}
PROM=${PROM:-http://localhost:9090}
BAD_TOPIC=${BAD_TOPIC:-link.created} # the typo: the real topic is links.created
compose() { docker compose -p ch24 -f "$CH24/data/compose.yaml" "$@"; }
now() { date -u +%H:%M:%S; }
say() { printf '%s  %s\n' "$(now)" "$*"; }

alert_state() { # alert_state NAME: firing, pending or none
  curl -fsS "$PROM/api/v1/alerts" | python3 -c "
import json, sys
alerts = [a for a in json.load(sys.stdin)['data']['alerts'] if a['labels']['alertname'] == '$1' and a['labels'].get('job') == 'shorty']
print(alerts[0]['state'] if alerts else 'none')"
}
error_ratio() {
  curl -fsS --get "$PROM/api/v1/query" --data-urlencode 'query=job:slo_errors_per_request:ratio_rate5m{job="shorty"}' |
    python3 -c "import json,sys; r=json.load(sys.stdin)['data']['result']; print(f'{float(r[0][\"value\"][1]):.3f}' if r else 'n/a')"
}

curl -fsS --get "$PROM/api/v1/query" --data-urlencode 'query=job:slo_availability:objective' | grep -q shorty ||
  { echo "Prometheus has no job:slo_availability:objective: is the stack running with RULES_DIR?"; exit 1; }

say "traffic starts (20 requests per second, 40 minutes at most)"
compose run -d --rm --name ch24-incident-load loadgen -rps 20 -d 40m >/dev/null 2>&1
sleep "$BASELINE"
say "baseline done, 5m error ratio $(error_ratio)"

say "CHANGE: shorty redeployed with SHORTY_KAFKA_TOPIC=$BAD_TOPIC"
SHORTY_KAFKA_TOPIC=$BAD_TOPIC compose up -d --no-deps --wait shorty >/dev/null 2>&1
deadline=$((SECONDS + 900)); last=""
while [ "$SECONDS" -lt "$deadline" ]; do
  state=$(alert_state SLOAvailabilityBudgetBurn)
  if [ "$state" != "$last" ]; then say "SLOAvailabilityBudgetBurn: $state (5m error ratio $(error_ratio))"; last=$state; fi
  [ "$state" = firing ] && break
  sleep 10
done
[ "$state" = firing ] || say "the page did not fire within 15 minutes"

say "MITIGATION: shorty rolled back to the default topic"
compose up -d --no-deps --wait shorty >/dev/null 2>&1
say "rollback done, 5m error ratio $(error_ratio)"
deadline=$((SECONDS + 900))
while [ "$SECONDS" -lt "$deadline" ]; do
  state=$(alert_state SLOAvailabilityBudgetBurn)
  if [ "$state" != "$last" ]; then say "SLOAvailabilityBudgetBurn: $state (5m error ratio $(error_ratio))"; last=$state; fi
  [ "$state" = none ] && break
  sleep 10
done
say "latency page: $(alert_state SLOLatencyBudgetBurn)"
docker stop ch24-incident-load >/dev/null 2>&1
say "traffic stopped; incident staged. Copy these lines into the postmortem timeline."
