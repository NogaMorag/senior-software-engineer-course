#!/usr/bin/env bash
# Runs the steps of the CI workflows on your own machine, in the same order and
# with the same commands, so you know they work before the first push.
# It does not run GitHub Actions itself: the YAML is checked by tests/lint.sh.
#
# Usage, from the root of the repository (~/senior-course):
#   tests/run_ci_locally.sh              Go and Python jobs, PostgreSQL in Docker
#   tests/run_ci_locally.sh --image      ... and build the shorty image
#   tests/run_ci_locally.sh --kind       ... and the kind deploy and smoke test
# Needs: go, uv, docker; kind and kubectl for --kind.
set -euo pipefail

want_image=false
want_kind=false
for arg in "$@"; do
  case "$arg" in
    --image) want_image=true ;;
    --kind) want_image=true; want_kind=true ;;
    *) echo "unknown argument: $arg" >&2; exit 2 ;;
  esac
done

root="$(pwd)"
logs="$(mktemp -d)"
pg_name="ch23-ci-postgres-$$"
cluster="ch23-local"
summary=()

cleanup() {
  docker rm -f "$pg_name" >/dev/null 2>&1 || true
  if [ "$want_kind" = true ]; then
    kind delete cluster --name "$cluster" >/dev/null 2>&1 || true
  fi
  rm -rf "$logs"
}
trap cleanup EXIT

step() { printf '\n==> %s\n' "$*"; }

# The same four steps as .github/actions/go-checks/action.yml.
go_checks() {
  local dir="$1" fail_on_skip="$2" log="$logs/$1.log"
  cd "$root/$dir"
  step "$dir: gofmt"
  local unformatted
  unformatted="$(gofmt -l .)"
  if [ -n "$unformatted" ]; then
    echo "gofmt would change these files:"; echo "$unformatted"; exit 1
  fi
  step "$dir: go vet"
  go vet ./...
  step "$dir: staticcheck"
  go run honnef.co/go/tools/cmd/staticcheck@2026.2.1 ./...
  step "$dir: go test -race"
  go test -race -count=1 -v ./... >"$log" 2>&1 || { cat "$log"; exit 1; }
  local passed skipped
  passed="$(grep -c -- '--- PASS' "$log" || true)"
  skipped="$(grep -c -- '--- SKIP' "$log" || true)"
  echo "passed $passed, skipped $skipped"
  if [ "$fail_on_skip" = true ] && [ "$skipped" != 0 ]; then
    grep -- '--- SKIP' "$log"; echo "tests were skipped; is DATABASE_URL set?"; exit 1
  fi
  summary+=("$dir: $passed passed, $skipped skipped")
  cd "$root"
}

# ch06-shorty.yml, job "test".
go_checks ch06 false

# ch11-isolation-lab.yml: the service container becomes a local container
# with the same image, credentials and health check.
step "ch11: start postgres:18.6"
docker run -d --name "$pg_name" \
  -e POSTGRES_USER=course -e POSTGRES_PASSWORD=change-me -e POSTGRES_DB=lab \
  -p 127.0.0.1::5432 \
  --health-cmd "pg_isready -U course -d lab" --health-interval 2s \
  --health-timeout 3s --health-retries 30 \
  postgres:18.6 >/dev/null
for _ in $(seq 1 60); do
  status="$(docker inspect -f '{{.State.Health.Status}}' "$pg_name")"
  [ "$status" = healthy ] && break
  sleep 1
done
[ "$status" = healthy ] || { docker logs "$pg_name"; exit 1; }
port="$(docker port "$pg_name" 5432/tcp | head -n1 | sed 's/.*://')"
export DATABASE_URL="postgres://course:change-me@localhost:${port}/lab"
echo "postgres healthy on port $port"
go_checks ch11 true
unset DATABASE_URL

# ch08-receiver.yml.
cd "$root/ch08"
step "ch08: uv lock --check";   uv lock --check
step "ch08: uv sync --frozen";  uv sync --frozen
step "ch08: ruff check";        uv run --no-sync ruff check .
step "ch08: ruff format";       uv run --no-sync ruff format --check .
step "ch08: mypy";              uv run --no-sync mypy src
step "ch08: pytest"
pytest_out="$(uv run --no-sync pytest -q)" || { echo "$pytest_out"; exit 1; }
echo "$pytest_out" | tail -n 3
summary+=("ch08: $(echo "$pytest_out" | tail -n 1)")
cd "$root"

if [ "$want_image" = true ]; then
  step "ch06: docker build"
  docker build -t shorty:ci ch06
  summary+=("image: shorty:ci built")
fi

if [ "$want_kind" = true ]; then
  # ch06-shorty.yml, job "kind-smoke".
  step "kind: create cluster $cluster"
  # A private kubeconfig, so your own ~/.kube/config is left alone.
  export KUBECONFIG="$logs/kubeconfig"
  kind create cluster --name "$cluster" --wait 60s \
    --image kindest/node:v1.37.0@sha256:a1ed56cfb0e7b93589bdf97c8cd566405a265939e3620fc4f5de89adff580ae5
  kind load docker-image shorty:ci --name "$cluster"
  cd "$root/ch06"
  kubectl apply -f deploy/shorty.yaml
  kubectl rollout status deployment/shorty --timeout=120s
  ./deploy/smoke.sh
  summary+=("kind: smoke test passed")
  cd "$root"
fi

printf '\n==> summary\n'
printf '%s\n' "${summary[@]}"
