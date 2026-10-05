#!/usr/bin/env bash
# Static checks for the CI files, from the root of the repository:
# actionlint (with shellcheck when it is installed), shellcheck for the shell
# scripts, and "every workflow declares its permissions".
set -euo pipefail

if ! git rev-parse --git-dir >/dev/null 2>&1; then
  echo "run me from inside the Git repository (actionlint needs it)" >&2
  exit 2
fi

echo "==> actionlint"
actionlint
echo "actionlint: no findings"

echo "==> shellcheck"
shellcheck ch06/deploy/smoke.sh tests/*.sh
echo "shellcheck: no findings"

echo "==> permissions"
missing=0
for wf in .github/workflows/*.yml; do
  if ! grep -q '^permissions:' "$wf"; then
    echo "no top-level permissions: $wf"
    missing=1
  fi
done
[ "$missing" = 0 ] && echo "every workflow declares its permissions"
exit "$missing"
