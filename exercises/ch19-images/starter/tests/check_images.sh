#!/usr/bin/env bash
# Builds the chapter 19 images and checks what the chapter promises: they run,
# they serve, they are not root, they survive a read-only root filesystem, and
# the compose stack comes up healthy. Run from anywhere:
#   bash tests/check_images.sh
set -euo pipefail
cd "$(dirname "$0")/.."

SHA=$(git rev-parse --short=12 HEAD 2>/dev/null || echo unknown)
pass=0
ok()   { echo "ok   $*"; pass=$((pass + 1)); }
fail() { echo "FAIL $*"; exit 1; }
cleanup() {
  docker rm -f ch19-check-shorty ch19-check-receiver >/dev/null 2>&1 || true
  docker compose -p ch19 down -v >/dev/null 2>&1 || true
}
trap cleanup EXIT

wait_http() { # url: retry for up to 10 seconds
  for _ in $(seq 1 20); do
    curl -fsS "$1" >/dev/null 2>&1 && return 0
    sleep 0.5
  done
  return 1
}

echo "== build"
docker build -q -t shorty:dev --build-arg GIT_SHA="$SHA" shorty >/dev/null || fail "build shorty"
docker build -q -t shorty:naive -f shorty/Dockerfile.naive shorty >/dev/null || fail "build shorty naive"
docker build -q -t receiver:dev --build-arg GIT_SHA="$SHA" receiver >/dev/null || fail "build receiver"
ok "three images built"

echo "== users"
user=$(docker image inspect --format '{{.Config.User}}' shorty:dev)
[ "$user" = "65532:65532" ] && ok "shorty:dev User is $user" || fail "shorty:dev User is '$user'"
user=$(docker image inspect --format '{{.Config.User}}' receiver:dev)
[ "$user" = "10001:10001" ] && ok "receiver:dev User is $user" || fail "receiver:dev User is '$user'"
user=$(docker image inspect --format '{{.Config.User}}' shorty:naive)
[ -z "$user" ] && ok "shorty:naive has no User, so it runs as root (expected)" || fail "naive User '$user'"
id_out=$(docker run --rm --entrypoint id receiver:dev)
[[ "$id_out" == uid=10001* ]] && ok "receiver process: $id_out" || fail "receiver id: $id_out"
if docker run --rm --entrypoint sh shorty:dev -c true >/dev/null 2>&1; then
  fail "shorty:dev has a shell"
else
  ok "shorty:dev has no shell (distroless)"
fi

echo "== run with a read-only root filesystem"
docker run -d --name ch19-check-shorty --read-only -p 127.0.0.1:18080:8080 \
  -e SHORTY_BASE_URL=http://localhost:18080 shorty:dev >/dev/null
wait_http http://127.0.0.1:18080/healthz || fail "shorty healthz"
uid=$(docker top ch19-check-shorty -o pid,uid | awk 'NR == 2 {print $2}')
[ "$uid" = "65532" ] && ok "shorty process runs as uid $uid" || fail "shorty process uid '$uid'"
code=$(curl -fsS -X POST localhost:18080/v1/links -H 'Content-Type: application/json' \
  -d '{"url":"https://go.dev/","code":"gohome"}' | sed -n 's/.*"code":"\([^"]*\)".*/\1/p')
[ "$code" = "gohome" ] && ok "shorty created a link" || fail "shorty create: '$code'"
loc=$(curl -sS -o /dev/null -w '%{http_code} %{redirect_url}' localhost:18080/gohome)
[ "$loc" = "302 https://go.dev/" ] && ok "shorty redirects: $loc" || fail "shorty resolve: $loc"
[ "$(curl -fsS localhost:18080/readyz)" = "ready" ] && ok "shorty readyz" || fail "shorty readyz"
docker run --rm --network container:ch19-check-shorty --entrypoint /shorty shorty:dev healthcheck \
  && ok "shorty healthcheck subcommand exits 0" || fail "shorty healthcheck"

docker run -d --name ch19-check-receiver --read-only -p 127.0.0.1:18081:8081 \
  -e RECEIVER_DEFAULT_SECRET=change-me receiver:dev >/dev/null
wait_http http://127.0.0.1:18081/stats || fail "receiver stats"
ok "receiver serves /stats on a read-only filesystem"

echo "== compose"
docker rm -f ch19-check-shorty ch19-check-receiver >/dev/null
docker compose -p ch19 up --build --detach --wait receiver shorty >/dev/null 2>&1 || fail "compose up --wait"
for svc in receiver shorty; do
  state=$(docker inspect --format '{{.State.Health.Status}}' "$(docker compose -p ch19 ps -q "$svc")")
  [ "$state" = "healthy" ] && ok "compose $svc is $state" || fail "compose $svc is $state"
done
smoke=$(docker compose -p ch19 run --rm smoke 2>/dev/null | tail -n 1)
[[ "$smoke" == 200* ]] && ok "smoke delivery through compose DNS: $smoke" || fail "smoke: $smoke"

echo "== sizes (record them in your notes)"
docker image ls --format '{{.Repository}}:{{.Tag}}\t{{.Size}}' | grep -E '^(shorty|receiver):' | sort
echo "PASS: $pass checks"
