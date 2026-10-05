#!/usr/bin/env bash
# Static checks for the Hookline chart and the kind manifests (milestone M7, after chapter 21's
# tests/check_static.sh). No cluster needed; Docker and the network are (kubeconform runs from its image
# and downloads the Kubernetes schemas).
#   scripts/check-chart.sh
set -euo pipefail
cd "$(dirname "$0")/.."

KUBECONFORM_IMAGE="ghcr.io/yannh/kubeconform:v0.8.0"
K8S_VERSION="1.37.0"
CHART=deploy/helm/hookline
fails=0
pass() { echo "PASS  $1"; }
fail() { echo "FAIL  $1"; fails=$((fails + 1)); }
kubeconform() { docker run --rm -i -v "$PWD/deploy:/work:ro" "$KUBECONFORM_IMAGE" -strict -summary -kubernetes-version "$K8S_VERSION" "$@"; }

if kubeconform /work/kind/infra.yaml; then pass "kubeconform deploy/kind/infra.yaml"; else fail "kubeconform deploy/kind/infra.yaml"; fi
if helm lint --strict "$CHART" > /dev/null; then pass "helm lint --strict"; else fail "helm lint --strict"; fi

all="$(helm template hookline "$CHART" --set api.autoscaling.enabled=true --set worker.autoscaling.enabled=true)" || all=""
if [ -n "$all" ] && printf '%s\n' "$all" | kubeconform -; then pass "helm template | kubeconform"; else fail "helm template | kubeconform"; fi
for want in "name: hookline-api" "name: hookline-relay" "name: hookline-dispatcher" "name: hookline-worker" \
            "kind: HorizontalPodAutoscaler" "\"helm.sh/hook\": pre-install,pre-upgrade" "checksum/config:" \
            "readOnlyRootFilesystem: true" "runAsNonRoot: true" "enableServiceLinks: false" \
            "secretKeyRef:" "terminationGracePeriodSeconds: 60" "/usr/local/bin/hookline-admin"; do
  if printf '%s\n' "$all" | grep -qF "$want"; then pass "rendered chart has $want"; else fail "rendered chart has $want"; fi
done
if helm template hookline "$CHART" | grep -qF "kind: HorizontalPodAutoscaler"; then fail "HPAs off by default"; else pass "HPAs off by default"; fi
if helm template hookline "$CHART" | grep -qE "postgres://|change-me"; then fail "no credential in the rendered chart"; else pass "no credential in the rendered chart"; fi
if grep -qE "image: .*:latest|image: [^:]+$" deploy/kind/infra.yaml; then fail "infra images are pinned"; else pass "infra images are pinned"; fi

for bad in "api.replicas=two" "logLevel=chatty" "image.pullPolicy=Sometimes" "worker.lease=30" "kafkaBrokers=nope" "replicas=3"; do
  if helm template hookline "$CHART" --set "$bad" > /dev/null 2>&1; then fail "schema rejects $bad"; else pass "schema rejects $bad"; fi
done

echo
if [ "$fails" -eq 0 ]; then echo "all chart checks passed"; else echo "$fails check(s) failed"; exit 1; fi
