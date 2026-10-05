#!/usr/bin/env bash
# Validates the chapter 20 manifests against the running cluster's API server
# (a server-side dry run: real schema, real admission, nothing is stored) and
# checks the rules the chapter sets. Needs the kind cluster from step 1:
#   bash tests/check_manifests.sh
set -euo pipefail
cd "$(dirname "$0")/.."

pass=0
ok()   { echo "ok   $*"; pass=$((pass + 1)); }
fail() { echo "FAIL $*"; exit 1; }
dry()  { kubectl apply --dry-run=server -f "$1" -o jsonpath="$2"; }

kubectl apply -f k8s/00-namespace.yaml >/dev/null # namespaced dry runs need it to exist
for f in k8s/*.yaml pods/*.yaml; do
  kubectl apply --dry-run=server -f "$f" >/dev/null || fail "API server rejected $f"
done
ok "API server accepts every manifest (server-side dry run)"

images=$(grep -h 'image:' k8s/*.yaml pods/*.yaml | awk '{print $2}')
for img in $images; do
  if [[ "$img" != *:* ]]; then fail "image $img has no tag (that means latest)"; fi
  if [[ "$img" == *:latest ]]; then fail "image $img uses latest"; fi
done
ok "every image has an explicit, non-latest tag: $(echo $images | tr ' ' ',')"

if grep -q '^kind: Secret' k8s/*.yaml pods/*.yaml; then fail "a Secret is stored in a manifest"; fi
ok "no Secret object is stored in the manifests"

d=k8s/shorty-deployment.yaml
probes=$(dry "$d" '{.spec.template.spec.containers[0].startupProbe.httpGet.path} {.spec.template.spec.containers[0].livenessProbe.httpGet.path} {.spec.template.spec.containers[0].readinessProbe.httpGet.path}')
[ "$probes" = "/healthz /healthz /readyz" ] && ok "shorty probes: startup, liveness, readiness = $probes" || fail "shorty probes = '$probes'"

for d in k8s/shorty-deployment.yaml k8s/receiver-deployment.yaml; do
  sec=$(dry "$d" '{.spec.template.spec.securityContext.runAsNonRoot} {.spec.template.spec.containers[0].securityContext.readOnlyRootFilesystem} {.spec.template.spec.containers[0].securityContext.allowPrivilegeEscalation}')
  [ "$sec" = "true true false" ] && ok "$d: runAsNonRoot, read-only root, no privilege escalation" || fail "$d security = '$sec'"
done

for app in shorty receiver; do
  sel=$(dry "k8s/$app-service.yaml" '{.spec.selector.app}')
  lab=$(dry "k8s/$app-deployment.yaml" '{.spec.template.metadata.labels.app}')
  [ -n "$sel" ] && [ "$sel" = "$lab" ] && ok "Service $app selects the pods of Deployment $app (app=$sel)" || fail "$app selector '$sel' vs labels '$lab'"
done

cfg=$(dry k8s/shorty-deployment.yaml '{.spec.template.spec.containers[0].envFrom[0].configMapRef.name}')
[ "$cfg" = "$(dry k8s/shorty-configmap.yaml '{.metadata.name}')" ] && ok "shorty reads its environment from ConfigMap $cfg" || fail "envFrom '$cfg'"

echo "PASS: $pass checks"
