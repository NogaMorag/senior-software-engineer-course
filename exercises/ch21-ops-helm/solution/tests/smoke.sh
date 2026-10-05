#!/usr/bin/env bash
# Smoke test against a running kind cluster that already has the image
# shorty:ch21 loaded: install the chart, wait, create and resolve one link,
# upgrade, roll back, uninstall.
#   bash tests/smoke.sh            # uses the chart in the directory above tests/
set -euo pipefail

ROOT="${1:-$(cd "$(dirname "$0")/.." && pwd)}"
NS=smoke
REL=smoke
PORT=18080

cleanup() {
  [ -n "${pf:-}" ] && kill "$pf" 2>/dev/null || true
  helm uninstall "$REL" -n "$NS" --wait >/dev/null 2>&1 || true
  kubectl delete namespace "$NS" --wait=false >/dev/null 2>&1 || true
}
trap cleanup EXIT

helm install "$REL" "$ROOT/chart/shorty" -n "$NS" --create-namespace --wait --timeout 120s
kubectl -n "$NS" port-forward "svc/$REL-shorty" "$PORT:80" >/dev/null 2>&1 &
pf=$!
for _ in $(seq 1 30); do curl -sf "localhost:$PORT/healthz" >/dev/null && break; sleep 1; done

curl -sf "localhost:$PORT/readyz" | grep -q ready
curl -sf -X POST "localhost:$PORT/v1/links" -H 'Content-Type: application/json' \
  -d '{"url": "https://go.dev/", "code": "smoke1"}' >/dev/null
location="$(curl -s -o /dev/null -w '%{redirect_url}' "localhost:$PORT/smoke1")"
[ "$location" = "https://go.dev/" ] || { echo "FAIL resolve: got '$location'"; exit 1; }
echo "PASS  install, create and resolve"

helm upgrade "$REL" "$ROOT/chart/shorty" -n "$NS" --reuse-values --set config.logLevel=debug --wait --timeout 120s >/dev/null
[ "$(helm history "$REL" -n "$NS" -o json | grep -o '"revision":' | wc -l | tr -d ' ')" = 2 ] || { echo "FAIL upgrade"; exit 1; }
echo "PASS  upgrade to revision 2"

helm rollback "$REL" 1 -n "$NS" --wait --timeout 120s >/dev/null
level="$(kubectl -n "$NS" get configmap "$REL-shorty" -o jsonpath='{.data.SHORTY_LOG_LEVEL}')"
[ "$level" = info ] || { echo "FAIL rollback: log level $level"; exit 1; }
echo "PASS  rollback to revision 1 (as revision 3)"
echo "smoke test passed"
