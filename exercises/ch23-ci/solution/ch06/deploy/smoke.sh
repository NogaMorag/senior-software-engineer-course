#!/usr/bin/env bash
# Smoke test for shorty on a cluster: port-forward the Service, create a link
# with a known code, and check that it resolves with a redirect.
# Usage: deploy/smoke.sh            (kubectl must point at the cluster)
set -euo pipefail

port="${SMOKE_PORT:-18080}"
base="http://127.0.0.1:${port}"
log="$(mktemp)"

kubectl port-forward service/shorty "${port}:80" >"${log}" 2>&1 &
forward_pid=$!
trap 'kill "${forward_pid}" 2>/dev/null || true; rm -f "${log}"' EXIT

# port-forward needs a moment before it accepts connections.
for _ in $(seq 1 30); do
  if curl -fsS "${base}/healthz" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
curl -fsS "${base}/healthz" >/dev/null || { cat "${log}"; exit 1; }
echo "healthz: ok"

created="$(curl -fsS -X POST "${base}/v1/links" \
  -H 'Content-Type: application/json' \
  -d '{"url": "https://example.com/from-ci", "code": "ci-smoke"}')"
echo "created: ${created}"

status="$(curl -s -o /dev/null -w '%{http_code}' "${base}/ci-smoke")"
if [ "${status}" != "302" ]; then
  echo "expected 302 from /ci-smoke, got ${status}" >&2
  exit 1
fi
echo "resolve: 302"
echo "smoke test passed"
