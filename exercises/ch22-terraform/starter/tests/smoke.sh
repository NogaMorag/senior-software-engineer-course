#!/usr/bin/env bash
# End-to-end run against the kind cluster "ch22" (image shorty:ch21 loaded):
# part 1 plan, apply, one change, drift, destroy; part 2 apply through the
# module, call the service, destroy. Works on scratch copies, so no state file
# is left in your tree.
#   bash tests/smoke.sh
set -euo pipefail

ROOT="${1:-$(cd "$(dirname "$0")/.." && pwd)}"
tf() { terraform -chdir="$work" "$@" -no-color -input=false; }
step() { echo "PASS  $1"; }
# On any failure, destroy what the current part created and remove the copy.
cleanup() { [ -n "${work:-}" ] && [ -d "$work" ] && { tf destroy -auto-approve >/dev/null 2>&1 || true; rm -rf "$work"; }; }
trap cleanup EXIT

# ---- part 1 ----
work="$(mktemp -d)"; cp -R "$ROOT/part1/." "$work/"
tf init >/dev/null
tf apply -auto-approve >/dev/null
step "part1 apply"
kubectl get namespace tf-demo >/dev/null
[ "$(kubectl -n tf-demo get configmap shorty-settings -o jsonpath='{.data.SHORTY_LOG_LEVEL}')" = info ]
step "namespace and ConfigMap exist"

set +e; tf plan -detailed-exitcode >/dev/null; code=$?; set -e
[ "$code" = 0 ] || { echo "FAIL a second plan should be empty (exit $code)"; exit 1; }
step "second plan is empty"

out="$(tf plan -var log_level=debug)"
grep -q "Plan: 0 to add, 1 to change, 0 to destroy." <<<"$out" || { echo "FAIL one change expected"; echo "$out"; exit 1; }
step "changing one variable plans exactly one change"

kubectl -n tf-demo patch configmap shorty-settings --type merge -p '{"data":{"SHORTY_LOG_LEVEL":"error"}}' >/dev/null
out="$(tf plan)"
grep -q '"error" -> "info"' <<<"$out" || { echo "FAIL drift not detected"; echo "$out"; exit 1; }
step "drift made with kubectl shows up in the plan"

tf destroy -auto-approve >/dev/null
kubectl wait --for=delete namespace/tf-demo --timeout=120s >/dev/null 2>&1 || true
! kubectl get namespace tf-demo >/dev/null 2>&1 || { echo "FAIL namespace survived destroy"; exit 1; }
step "part1 destroy removes the namespace"
rm -rf "$work"; work=""

# ---- part 2 ----
work="$(mktemp -d)"; cp -R "$ROOT/part2/." "$work/"
tf init >/dev/null
tf apply -auto-approve >/dev/null
[ "$(helm status shorty -n shorty-tf -o json | grep -o '"status":"deployed"')" = '"status":"deployed"' ]
step "part2 apply installs the release through the module"

kubectl -n shorty-tf port-forward svc/shorty 18090:80 >/dev/null 2>&1 &
pf=$!
for _ in $(seq 1 30); do curl -sf localhost:18090/healthz >/dev/null && break; sleep 1; done
curl -sf localhost:18090/readyz | grep -q ready
kill "$pf"
step "the service answers"

terraform -chdir="$work" output -raw port_forward | grep -q "kubectl -n shorty-tf port-forward svc/shorty"
step "outputs"

tf destroy -auto-approve >/dev/null
kubectl wait --for=delete namespace/shorty-tf --timeout=120s >/dev/null 2>&1 || true
! helm status shorty -n shorty-tf >/dev/null 2>&1 || { echo "FAIL release survived destroy"; exit 1; }
! kubectl get namespace shorty-tf >/dev/null 2>&1 || { echo "FAIL namespace survived destroy"; exit 1; }
step "part2 destroy removes the release and the namespace"
rm -rf "$work"; work=""
echo "smoke test passed"
