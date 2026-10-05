#!/usr/bin/env bash
# Exercises the deployed chapter 20 objects end to end: rollout, calls through
# a port-forward, self-healing, the readiness drill, and cluster DNS. Run after
# the manifests are applied and the Secret exists:
#   bash tests/smoke.sh
set -euo pipefail
cd "$(dirname "$0")/.."

NS=ch20
pass=0
pids=()
ok()   { echo "ok   $*"; pass=$((pass + 1)); }
fail() { echo "FAIL $*"; exit 1; }
cleanup() { for p in "${pids[@]}"; do kill "$p" 2>/dev/null || true; done; }
trap cleanup EXIT

forward() { # target local:remote; returns when the local port answers
  kubectl -n "$NS" port-forward "$1" "$2" >/dev/null 2>&1 &
  pids+=($!)
  for _ in $(seq 1 30); do
    nc -z 127.0.0.1 "${2%%:*}" 2>/dev/null && return 0
    sleep 0.5
  done
  fail "port-forward $1 $2 did not come up"
}
ready_endpoints() {
  kubectl -n "$NS" get endpointslices -l kubernetes.io/service-name=shorty \
    -o jsonpath='{range .items[*].endpoints[*]}{.conditions.ready}{"\n"}{end}' | grep -c true || true
}
wait_for() { # seconds description command...
  local n=$1 what=$2; shift 2
  for _ in $(seq 1 "$n"); do "$@" && return 0; sleep 1; done
  fail "timed out waiting for $what"
}

kubectl -n "$NS" rollout status deployment/shorty --timeout=180s >/dev/null
kubectl -n "$NS" rollout status deployment/receiver --timeout=180s >/dev/null
ok "both Deployments rolled out"
[ "$(ready_endpoints)" = "2" ] && ok "Service shorty has 2 ready endpoints" || fail "ready endpoints: $(ready_endpoints)"

forward svc/shorty 18080:80
code=$(curl -fsS -X POST localhost:18080/v1/links -H 'Content-Type: application/json' \
  -d '{"url":"https://kubernetes.io/","code":"k8s"}' | sed -n 's/.*"code":"\([^"]*\)".*/\1/p')
[ "$code" = "k8s" ] && ok "created a link through the port-forward" || fail "create: '$code'"
loc=$(curl -sS -o /dev/null -w '%{http_code} %{redirect_url}' localhost:18080/k8s)
[ "$loc" = "302 https://kubernetes.io/" ] && ok "resolved it: $loc" || fail "resolve: $loc"
base=$(curl -fsS localhost:18080/v1/links/k8s | sed -n 's/.*"short_url":"\([^"]*\)".*/\1/p')
[ "$base" = "http://localhost:8080/k8s" ] && ok "short_url uses SHORTY_BASE_URL from the ConfigMap" || fail "short_url '$base'"

victim=$(kubectl -n "$NS" get pods -l app=shorty -o jsonpath='{.items[0].metadata.name}')
kubectl -n "$NS" delete pod "$victim" --wait=false >/dev/null
gone_and_back() {
  ! kubectl -n "$NS" get pods -l app=shorty -o name | grep -q "$victim" && [ "$(ready_endpoints)" = "2" ]
}
wait_for 90 "a replacement pod" gone_and_back
ok "deleted $victim; the ReplicaSet replaced it and 2 endpoints are ready again"

pod=$(kubectl -n "$NS" get pods -l app=shorty -o jsonpath='{.items[0].metadata.name}')
forward "pod/$pod" 16060:6060 # the debug listener is bound to 127.0.0.1 inside the pod
curl -sS -o /dev/null -X POST 'localhost:16060/debug/ready?state=off'
one_ready() { [ "$(ready_endpoints)" = "1" ]; }
wait_for 30 "the pod to leave the endpoints" one_ready
restarts=$(kubectl -n "$NS" get pod "$pod" -o jsonpath='{.status.containerStatuses[0].restartCount}')
[ "$restarts" = "0" ] && ok "readiness off: $pod left the endpoints, restarts=$restarts" || fail "restarts=$restarts"
curl -fsS -o /dev/null -X POST 'localhost:16060/debug/ready?state=on'
two_ready() { [ "$(ready_endpoints)" = "2" ]; }
wait_for 30 "the pod to return" two_ready
ok "readiness on: back to 2 ready endpoints"

kubectl -n "$NS" delete pod sender --ignore-not-found >/dev/null
kubectl apply -f pods/sender.yaml >/dev/null
done_pod() { [ "$(kubectl -n "$NS" get pod sender -o jsonpath='{.status.phase}')" = "Succeeded" ]; }
wait_for 60 "the sender pod" done_pod
out=$(kubectl -n "$NS" logs sender)
[[ "$out" == 200* ]] && ok "sender reached receiver.ch20.svc.cluster.local: $out" || fail "sender: $out"
kubectl -n "$NS" delete pod sender >/dev/null

echo "PASS: $pass checks"
