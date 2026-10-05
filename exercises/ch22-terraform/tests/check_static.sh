#!/usr/bin/env bash
# Static checks for chapter 22; no cluster needed. terraform init downloads the
# providers the first time (network), then works from the local cache.
#   bash tests/check_static.sh            # checks the directory above tests/
set -euo pipefail

ROOT="${1:-$(cd "$(dirname "$0")/.." && pwd)}"
fails=0
pass() { echo "PASS  $1"; }
fail() { echo "FAIL  $1"; fails=$((fails + 1)); }

for dir in part1 part2 optional-aws; do
  [ -d "$ROOT/$dir" ] || continue
  # Validate a scratch copy, so .terraform/ never lands in your tree.
  work="$(mktemp -d)"
  cp -R "$ROOT/$dir/." "$work/"
  if terraform -chdir="$work" fmt -check -recursive >/dev/null; then pass "$dir: terraform fmt -check"; else fail "$dir: terraform fmt -check (run terraform fmt -recursive)"; fi
  if terraform -chdir="$work" init -backend=false -input=false >/dev/null && terraform -chdir="$work" validate -no-color >/dev/null; then
    pass "$dir: terraform validate"
  else
    fail "$dir: terraform validate"
  fi
  rm -rf "$work"
done

# A match counts only outside comments, so the TODO text in the starter does not.
has() { grep -rhE "$2" "$ROOT/$1" --include='*.tf' | grep -qvE '^[[:space:]]*#'; }
has part1 'resource "kubernetes_namespace_v1"' && pass "part1: namespace resource" || fail "part1: namespace resource (TODO 1)"
has part1 'resource "kubernetes_config_map_v1"' && pass "part1: ConfigMap resource" || fail "part1: ConfigMap resource (TODO 2)"
has part1 'output "config_map"' && pass "part1: outputs" || fail "part1: outputs (TODO 3)"
has part2 'provider "helm"' && pass "part2: helm provider" || fail "part2: helm provider (TODO 4)"
has part2/modules/shorty 'resource "helm_release"' && pass "part2: helm_release in the module" || fail "part2: helm_release (TODO 6)"
has part2 'module "shorty"' && pass "part2: module call" || fail "part2: module call (TODO 7)"
[ -f "$ROOT/part2/modules/shorty/chart/Chart.yaml" ] && pass "part2: chart copied into the module" || fail "part2: copy your chapter 21 chart to modules/shorty/chart"

echo
if [ "$fails" -eq 0 ]; then echo "all static checks passed"; else echo "$fails check(s) failed"; exit 1; fi
