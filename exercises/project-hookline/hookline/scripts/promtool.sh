#!/usr/bin/env bash
# Checks and unit-tests Hookline's Prometheus rules with promtool, from the Prometheus image, so nothing
# needs installing (chapter 25's tests/run_promtool.sh, for deploy/observability).
#   scripts/promtool.sh
set -euo pipefail
cd "$(dirname "$0")/../deploy/observability"
IMAGE=prom/prometheus:v3.15.0
docker run --rm -v "$PWD":/work -w /work --entrypoint promtool "$IMAGE" check rules rules/*.yml
docker run --rm -v "$PWD":/work -w /work/tests --entrypoint promtool "$IMAGE" test rules rules_test.yml
# Every alert's runbook_url points at a runbook that exists.
cd ../..
missing=0
while read -r f; do
  if [ -f "$f" ]; then echo "runbook ok: $f"; else echo "runbook MISSING: $f"; missing=1; fi
done < <(grep -ho 'runbook_url: [^ ]*' deploy/observability/rules/*.yml | awk '{print $2}' | sort -u)
exit "$missing"
