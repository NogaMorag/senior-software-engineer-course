#!/usr/bin/env bash
# Checks and unit-tests the rules with promtool, from the Prometheus image, so
# nothing needs installing. Run from the folder that holds rules/ and tests/.
#
#   bash tests/run_promtool.sh
set -euo pipefail
IMAGE=prom/prometheus:v3.15.0
docker run --rm -v "$PWD":/work -w /work --entrypoint promtool "$IMAGE" \
  check rules rules/slo-recording.yml rules/slo-objectives.yml rules/slo-alerts.yml
docker run --rm -v "$PWD":/work -w /work/tests --entrypoint promtool "$IMAGE" \
  test rules slo_test.yml
