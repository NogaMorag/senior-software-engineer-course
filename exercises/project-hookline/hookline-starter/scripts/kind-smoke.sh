#!/usr/bin/env bash
# Milestone M7: Hookline on a kind cluster, from nothing to "an event reached the in-cluster sink".
#   scripts/kind-smoke.sh            # create the cluster, build, load, install, smoke, delete
#   KEEP=1 scripts/kind-smoke.sh     # the same, but leave the cluster running afterwards
# Uses its own kubeconfig file, so your other clusters and contexts are never touched.
# Needs: docker, kind, kubectl, helm.
set -euo pipefail
cd "$(dirname "$0")/.."

CLUSTER=${CLUSTER:-hookline}
NS=hookline
TAG=${TAG:-dev}
export KUBECONFIG=${KUBECONFIG_FILE:-/tmp/hookline-kind.kubeconfig}

# shellcheck disable=SC2329 # called by the trap
finish() {
  local code=$?
  if [ -n "${PF:-}" ]; then kill "$PF" 2> /dev/null || true; fi
  if [ "$code" -ne 0 ]; then
    echo "== failed; what the cluster saw" >&2
    kubectl -n "$NS" get pods -o wide >&2 || true
    kubectl -n "$NS" get events --sort-by=.lastTimestamp 2> /dev/null | tail -20 >&2 || true
  fi
  if [ "${KEEP:-0}" != 1 ] && [ "${CREATED:-0}" = 1 ]; then
    kind delete cluster --name "$CLUSTER" > /dev/null 2>&1 || true
    rm -f "$KUBECONFIG"
  fi
  exit "$code"
}
trap finish EXIT

echo "== cluster"
if ! kind get clusters 2> /dev/null | grep -qx "$CLUSTER"; then
  kind create cluster --name "$CLUSTER" --config deploy/kind/kind-config.yaml --wait 120s
  CREATED=1
else
  kind export kubeconfig --name "$CLUSTER"
fi

echo "== images"
docker build -q -t "hookline:$TAG" . > /dev/null
docker build -q -t "hookline-sink:$TAG" sink > /dev/null
kind load docker-image "hookline:$TAG" "hookline-sink:$TAG" --name "$CLUSTER"

echo "== infrastructure"
kubectl apply -f deploy/kind/infra.yaml > /dev/null
for d in postgres redis redpanda sink; do kubectl -n "$NS" rollout status "deployment/$d" --timeout=180s; done
# The database URL lives in a Secret created here, outside the chart and outside Git.
kubectl -n "$NS" create secret generic hookline-database \
  --from-literal=url="postgres://hookline:change-me@postgres.$NS.svc.cluster.local:5432/hookline" \
  --dry-run=client -o yaml | kubectl apply -f - > /dev/null

echo "== helm install"
# devMode: the in-cluster sink is plain http. Small replica counts: the cluster shares a laptop.
helm upgrade --install hookline deploy/helm/hookline -n "$NS" --wait --timeout 300s \
  --set image.tag="$TAG" --set devMode=true \
  --set api.replicas=1 --set relay.replicas=2 --set dispatcher.replicas=1 --set worker.replicas=1
kubectl -n "$NS" get pods

echo "== smoke: one event to the in-cluster sink"
exec_api() { kubectl -n "$NS" exec deploy/hookline-api -- "$@"; }
KEY=$(exec_api /usr/local/bin/hookline-admin create-tenant -name smoke | sed -n 's/^api key: *//p')
kubectl -n "$NS" port-forward svc/hookline-api 18090:80 > /tmp/hookline-kind-pf.log 2>&1 &
PF=$!
for _ in $(seq 50); do curl -sf localhost:18090/healthz > /dev/null && break; sleep 0.2; done
call() { curl -sf "localhost:18090$1" -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' "${@:2}"; }
EP=$(call /v1/endpoints -d '{"url":"http://sink.hookline.svc.cluster.local:8081/hooks/smoke"}' |
  python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
# The sink knows one secret; give the endpoint that one instead of the generated one (a practice shortcut).
kubectl -n "$NS" exec deploy/postgres -- psql -U hookline -d hookline -qc \
  "UPDATE endpoints SET secret = 'whsec_change-me' WHERE id = '$EP'"
call /v1/subscriptions -d "{\"endpoint_id\":\"$EP\",\"event_type\":\"*\"}" > /dev/null
EV=$(call /v1/events -H 'Idempotency-Key: kind-smoke-1' -d '{"type":"smoke.test","payload":{"hello":"kind"}}' |
  python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
echo "event $EV"
stats() { kubectl -n "$NS" exec deploy/sink -- python -c \
  'import urllib.request; print(urllib.request.urlopen("http://127.0.0.1:8081/stats").read().decode())'; }
for _ in $(seq 60); do
  if stats | grep -q '"unique":1'; then
    echo "sink: $(stats)"
    echo "deliveries: $(call "/v1/events/$EV/deliveries" | python3 -c 'import json,sys; print([(d["status"], d["attempts"]) for d in json.load(sys.stdin)["items"]])')"
    echo "PASS: the event reached the in-cluster sink, signed"
    exit 0
  fi
  sleep 1
done
echo "FAIL: no delivery within 60s; sink: $(stats)" >&2
exit 1
