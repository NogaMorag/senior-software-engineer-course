#!/usr/bin/env bash
# test_doctor.sh: run doctor.sh against fake tools and check its exit codes.
#
#   bash test_doctor.sh                 # tests ./doctor.sh (the copy next to this file)
#   DOCTOR=path/to/doctor.sh bash test_doctor.sh
#
# Every fake tool is a tiny script in a temporary bin folder, and PATH holds only that folder,
# so the result does not depend on what is installed on this machine.
set -u

here=$(cd "$(dirname "$0")" && pwd)
DOCTOR=${DOCTOR:-$here/doctor.sh}
BASH_BIN=${BASH:-bash}
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
passed=0
failed=0

# fake NAME OUTPUT: a command that prints OUTPUT whatever its arguments are.
fake() {
  printf '#!/bin/sh\necho "%s"\n' "$2" > "$work/bin/$1"
  chmod +x "$work/bin/$1"
}

# setup: a full, healthy workbench (Day 0 and Part E tools at their pinned versions).
setup() {
  rm -rf "${work:?}/bin" "${work:?}/home" "${work:?}/course"
  mkdir -p "$work/bin" "$work/home/.ssh" "$work/home/senior-course/.git" "$work/course/exercises"
  touch "$work/home/.ssh/id_ed25519.pub"
  for tool in grep head uname timeout dirname; do
    if command -v "$tool" > /dev/null 2>&1; then ln -s "$(command -v "$tool")" "$work/bin/$tool"; fi
  done
  cat > "$work/bin/git" << 'EOF'
#!/bin/sh
if [ "$1" = "-C" ]; then [ -d "$2/.git" ] && echo true && exit 0; exit 128; fi
echo "git version 2.53.0"
EOF
  cat > "$work/bin/docker" << 'EOF'
#!/bin/sh
case "$1" in
  --version) echo "Docker version 29.8.1, build fake" ;;
  info) echo "29.8.1" ;;
  compose) echo "5.6.0" ;;
esac
EOF
  cat > "$work/bin/uv" << EOF
#!/bin/sh
if [ "\$1" = "python" ]; then echo "$work/bin/python3.14"; exit 0; fi
echo "uv 0.12.23 (fake)"
EOF
  fake python3.14 "Python 3.14.8"
  fake code "1.105.0"
  fake go "go version go1.27.1 linux/amd64"
  fake staticcheck "staticcheck 2026.2.1 (0.8.1)"
  fake goose "goose version: v3.28.0"
  fake buf "1.73.0"
  fake protoc-gen-go "protoc-gen-go v1.36.12"
  fake protoc-gen-go-grpc "protoc-gen-go-grpc 1.6.2"
  printf '#!/bin/sh\necho "Go: go1.27.1"\necho "Scanner: govulncheck@v1.8.0"\n' > "$work/bin/govulncheck"
  chmod +x "$work/bin/govulncheck"
  fake actionlint "v1.7.12"
  fake golangci-lint "golangci-lint has version 2.14.0 built with go1.27.0 from 114493f9 on 2026-09-24T11:07:15Z"
  fake kind "kind v0.33.0 go1.26.7 linux/amd64"
  fake kubectl "Client Version: v1.37.1"
  fake helm "v4.3.0+gfake"
  fake terraform "Terraform v1.16.5"
  fake k6 "k6 v2.3.0 (commit/fake, go1.26.8, linux/amd64)"
  chmod +x "$work/bin/git" "$work/bin/docker" "$work/bin/uv"
}

# expect NAME WANTED_EXIT [doctor args...]
expect() {
  local name=$1 want=$2
  shift 2
  local got
  env -i HOME="$work/home" PATH="$work/bin" ${COURSE_DIR:+COURSE="$COURSE_DIR"} \
    "$BASH_BIN" "$DOCTOR" "$@" > "$work/out.txt" 2>&1
  got=$?
  if [ "$got" -eq "$want" ]; then
    passed=$((passed + 1))
    echo "PASS $name"
  else
    failed=$((failed + 1))
    echo "FAIL $name: exit $got, want $want"
    sed 's/^/     | /' "$work/out.txt"
  fi
}

COURSE_DIR="$work/course"

setup
expect "healthy Day 0 workbench" 0

setup
expect "healthy workbench with Part E" 0 --part-e

setup
fake go "go version go1.26.8 linux/amd64"
expect "old Go is one failure" 1

setup
printf '#!/bin/sh\necho "Go: go1.27.1"\necho "Scanner: govulncheck@v1.7.0"\n' > "$work/bin/govulncheck"
expect "govulncheck is read from its Scanner line, not the Go line" 1

setup
rm "$work/bin/docker"
expect "no Docker fails client, engine and compose" 3

setup
rm "$work/bin/kind" "$work/bin/kubectl" "$work/bin/helm" "$work/bin/terraform" "$work/bin/k6"
expect "Day 0 check ignores missing Part E tools" 0
expect "missing Part E tools are five failures" 5 --part-e

setup
fake kubectl "Client Version: v1.37.0"
fake k6 "k6 v2.10.0 (commit/fake)"
expect "versions compare as numbers, not text" 1 --part-e

setup
COURSE_DIR="/mnt/c/Users/me/senior-software-engineer-course"
expect "a course folder on the Windows drive fails" 1
COURSE_DIR=""
expect "an unset COURSE fails" 1
COURSE_DIR="$work/course"

setup
expect "an unknown option is a usage error" 64 --bogus

echo "test_doctor: $passed passed, $failed failed"
exit "$failed"
