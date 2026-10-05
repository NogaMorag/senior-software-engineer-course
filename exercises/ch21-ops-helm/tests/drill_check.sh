#!/usr/bin/env bash
# Calls the four drill services from a short-lived pod inside the cluster and
# reports which ones answer. Before the drill all four fail; when you are done,
# all four pass.
#   bash tests/drill_check.sh
set -uo pipefail

kubectl -n drill delete pod drill-probe --ignore-not-found >/dev/null
kubectl -n drill run drill-probe --restart=Never --image=busybox:1.37.0 --command -- sh -c '
  for app in app-a app-b app-c app-d; do
    ok=no
    for try in 1 2 3; do  # a just-fixed Service can take a few seconds to reach every node
      if wget -q -T 3 -O - "http://$app/healthz" 2>/dev/null | grep -q ok; then ok=yes; break; fi
      sleep 2
    done
    if [ "$ok" = yes ]; then
      echo "PASS  $app answers /healthz through its Service"
    else
      echo "FAIL  $app"
    fi
  done' >/dev/null
kubectl -n drill wait pod/drill-probe --for=jsonpath='{.status.phase}'=Succeeded --timeout=90s >/dev/null
out="$(kubectl -n drill logs drill-probe)"
kubectl -n drill delete pod drill-probe --wait=false >/dev/null
echo "$out"
fails="$(printf '%s\n' "$out" | grep -c '^FAIL' || true)"
if [ "$fails" -eq 0 ] && [ -n "$out" ]; then echo "drill solved"; else echo "$fails service(s) still broken"; exit 1; fi
