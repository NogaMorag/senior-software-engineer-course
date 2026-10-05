#!/usr/bin/env bash
# The whole chapter 20 run in one go, from nothing back to nothing: create the
# kind cluster, build and load the images, create the Secret, check and apply
# the manifests, run the smoke test, delete the cluster.
#   bash tests/run_all.sh
set -euo pipefail
cd "$(dirname "$0")/.."

CLUSTER=ch20
DATA=data; [ -d "$DATA" ] || DATA=../data   # the build contexts for shorty and receiver
finish() {
  local code=$?
  if [ "$code" -ne 0 ]; then # show why before the evidence is deleted
    kubectl -n ch20 get pods -o wide || true
    kubectl -n ch20 get events --sort-by=.lastTimestamp | tail -n 20 || true
  fi
  kind delete cluster --name "$CLUSTER"
}
trap finish EXIT

kind create cluster --name "$CLUSTER" --config kind-config.yaml --wait 120s
docker build -q -t shorty:ch20 "$DATA/shorty"
docker build -q -t receiver:ch20 "$DATA/receiver"
kind load docker-image shorty:ch20 receiver:ch20 --name "$CLUSTER"

bash tests/check_manifests.sh
kubectl -n ch20 create secret generic receiver-secret --from-literal=signing-secret=change-me
kubectl apply -f k8s/
bash tests/smoke.sh
