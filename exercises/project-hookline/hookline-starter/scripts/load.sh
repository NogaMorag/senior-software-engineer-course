#!/usr/bin/env bash
# The load tests of milestone M9, against Hookline running on this machine with the whole compose stack
# (the observability one too, so you can watch the RED dashboard while k6 runs).
#
#   scripts/load.sh ingest          # spec N1: loadtest/ingest.js, RATE events per second (default 100)
#   scripts/load.sh two-tenants     # spec N5: loadtest/two-tenants.js, a noisy tenant and a quiet one
#   scripts/load.sh overload        # spec N6: loadtest/overload.js, RATE (default 800) against an API
#                                   # whose in-flight limit is MAX_IN_FLIGHT (default 4)
#
# k6 exits non-zero when a threshold fails, and so does this script. Its summary is saved to
# $OUT/<test>-summary.txt; copy the numbers you got, with the date and the machine, to docs/EVALUATION.md.
# Needs: docker, go, curl, python3, k6 v2. Leaves the stack up unless DOWN=1.
set -euo pipefail
cd "$(dirname "$0")/.."
OUT=${OUT:-/tmp/hookline-load}
# shellcheck source=scripts/lib-local.sh
. scripts/lib-local.sh

TEST=${1:-}
case "$TEST" in
  ingest | two-tenants | overload) ;;
  *) echo "usage: scripts/load.sh ingest | two-tenants | overload" >&2; exit 2 ;;
esac
# Overload: a small in-flight limit, so a laptop can push the API past it (read by start_hookline's api).
if [ "$TEST" = overload ]; then export HOOKLINE_MAX_IN_FLIGHT=${MAX_IN_FLIGHT:-4}; fi
command -v k6 > /dev/null || { echo "k6 is not on PATH (chapter 26 installs it)" >&2; exit 1; }

cleanup() {
  stop_all
  if [ "${DOWN:-0}" = 1 ]; then $COMPOSE down -v > /dev/null 2>&1; fi
}
trap cleanup EXIT

echo "== stack and binaries (logs in $OUT)"
$COMPOSE up -d --wait --build > "$OUT/compose.log" 2>&1 || { echo "compose failed; see $OUT/compose.log" >&2; exit 1; }
start_hookline
curl -sf -X POST "$SINK/reset" > /dev/null

echo "== $TEST"
args=(-e BASE_URL="$API" --summary-export "$OUT/$TEST-summary.json")
if [ "$TEST" = ingest ]; then
  args+=(-e API_KEY="$(new_tenant load)" -e RATE="${RATE:-100}")
elif [ "$TEST" = overload ]; then
  # A limit far above the flood, so every refusal comes from the shedder.
  args+=(-e API_KEY="$(new_tenant overload 100000)" -e RATE="${RATE:-800}")
else
  NOISY_RATE=${NOISY_RATE:-100}
  # The noisy tenant's limit is a fifth of what it sends; the quiet one's is far above its rate.
  args+=(-e NOISY_KEY="$(new_tenant noisy $((NOISY_RATE / 5)))" -e QUIET_KEY="$(new_tenant quiet 1000)")
  args+=(-e NOISY_RATE="$NOISY_RATE" -e QUIET_RATE="${QUIET_RATE:-20}")
fi
k6 run "${args[@]}" "loadtest/$TEST.js" 2>&1 | tee "$OUT/$TEST-summary.txt"
