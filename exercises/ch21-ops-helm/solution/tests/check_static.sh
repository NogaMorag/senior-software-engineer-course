#!/usr/bin/env bash
# Static checks for chapter 21; no cluster needed, Docker and network needed
# (kubeconform runs from its image and downloads the Kubernetes schemas).
#   bash tests/check_static.sh            # checks the directory above tests/
#   bash tests/check_static.sh ~/senior-course/ch21
set -euo pipefail

ROOT="${1:-$(cd "$(dirname "$0")/.." && pwd)}"
KUBECONFORM_IMAGE="ghcr.io/yannh/kubeconform:v0.8.0"
K8S_VERSION="1.37.0"
CHART="$ROOT/chart/shorty"
fails=0

pass() { echo "PASS  $1"; }
fail() { echo "FAIL  $1"; fails=$((fails + 1)); }
kubeconform() { docker run --rm -i -v "$ROOT:/work:ro" "$KUBECONFORM_IMAGE" -strict -summary -kubernetes-version "$K8S_VERSION" "$@"; }

# 1. Every manifest is valid Kubernetes for the cluster's version.
if kubeconform /work/k8s /work/drill; then pass "kubeconform k8s/ and drill/"; else fail "kubeconform k8s/ and drill/"; fi

# 2. The plain manifests carry what the chapter asks for.
dep="$ROOT/k8s/deployment.yaml"
grep -q "requests:" "$dep" && grep -q "limits:" "$dep" && pass "deployment has requests and limits" || fail "deployment has requests and limits (TODO 1)"
grep -q "maxUnavailable: 0" "$dep" && grep -q "progressDeadlineSeconds:" "$dep" && pass "rolling update strategy" || fail "rolling update strategy (TODO 2)"
grep -q "kind: HorizontalPodAutoscaler" "$ROOT/k8s/hpa.yaml" && grep -q "autoscaling/v2" "$ROOT/k8s/hpa.yaml" && pass "HPA autoscaling/v2" || fail "HPA autoscaling/v2 (TODO 3)"
grep -q "kind: Ingress" "$ROOT/k8s/ingress.yaml" && pass "ingress" || fail "ingress (TODO 4)"

# 3. The chart lints and renders valid manifests, with and without the HPA.
if helm lint --strict "$CHART" >/dev/null; then pass "helm lint --strict"; else fail "helm lint --strict"; fi
rendered="$(helm template ch21 "$CHART" --set autoscaling.enabled=true)" || rendered=""
if [ -n "$rendered" ] && printf '%s\n' "$rendered" | kubeconform -; then pass "helm template | kubeconform"; else fail "helm template | kubeconform"; fi
for want in "kind: ConfigMap" "kind: HorizontalPodAutoscaler" "readOnlyRootFilesystem: true" "runAsNonRoot: true" \
            "readinessProbe:" "livenessProbe:" "memory: 64Mi" "SHORTY_STRESS: \"off\"" "checksum/config:"; do
  if printf '%s\n' "$rendered" | grep -qF "$want"; then pass "rendered chart has $want"; else fail "rendered chart has $want (TODO 5 to 7)"; fi
done
if helm template ch21 "$CHART" | grep -qF "kind: HorizontalPodAutoscaler"; then fail "HPA must be off by default"; else pass "HPA off by default"; fi

# 4. The values schema rejects bad values before anything reaches the cluster.
for bad in "replicaCount=two" "config.logLevel=chatty" "image.pullPolicy=Sometimes" "service.port=70000" "replicas=3"; do
  if helm template ch21 "$CHART" --set "$bad" >/dev/null 2>&1; then fail "schema rejects $bad (TODO 8)"; else pass "schema rejects $bad"; fi
done

echo
if [ "$fails" -eq 0 ]; then echo "all static checks passed"; else echo "$fails check(s) failed"; exit 1; fi
