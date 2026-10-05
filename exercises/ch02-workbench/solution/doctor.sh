#!/usr/bin/env bash
# doctor.sh: check that the course workbench is ready (chapter 2).
#
#   bash doctor.sh            # the Day 0 tools and the course folders
#   bash doctor.sh --part-e   # the same, plus kind, kubectl, Helm, Terraform and k6 (Day 57)
#
# Exit code: the number of failed checks (0 means ready); 64 for an unknown option.
# Works with bash 3.2 (macOS) and bash 5 (Ubuntu); outside bash it uses only grep, head and uname.
set -u
export CHECKPOINT_DISABLE=1 # stop Terraform from calling home during the version check

PART_E=0
OK=0
FAILED=0
LINE_FILTER='.'

usage() {
  echo "usage: bash doctor.sh [--part-e]"
}

case "${1:-}" in
  "") ;;
  --part-e) PART_E=1 ;;
  -h | --help) usage; exit 0 ;;
  *) echo "doctor: unknown option: $1" >&2; usage >&2; exit 64 ;;
esac

# first_version: print the first x.y or x.y.z found on stdin.
first_version() {
  grep -Eo '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -n 1
}

# version_ge A B: succeed when version A >= version B, comparing three numeric fields.
version_ge() {
  local IFS=.
  local -a a b
  local i x y
  read -r -a a <<< "$1"
  read -r -a b <<< "$2"
  for i in 0 1 2; do
    x=${a[i]:-0}
    y=${b[i]:-0}
    if ((10#$x > 10#$y)); then return 0; fi
    if ((10#$x < 10#$y)); then return 1; fi
  done
  return 0
}

# limited: run a command with a 20-second limit when `timeout` exists (it does not on stock macOS).
limited() {
  if command -v timeout > /dev/null 2>&1; then
    timeout 20 "$@"
  else
    "$@"
  fi
}

pass() { # name found need
  OK=$((OK + 1))
  printf '  ok    %-18s %-12s (%s)\n' "$1" "$2" "$3"
}

fail() { # name found need hint
  FAILED=$((FAILED + 1))
  printf '  FAIL  %-18s %-12s (%s)  see %s\n' "$1" "$2" "$3" "$4"
}

# check_tool NAME MINIMUM HINT COMMAND...: the command must exist and print a version >= MINIMUM.
# A MINIMUM of "-" means "present, any version".
check_tool() {
  local name=$1 min=$2 hint=$3
  shift 3
  local need="need >= $min"
  [ "$min" = "-" ] && need="any version"
  if ! command -v "$1" > /dev/null 2>&1; then
    fail "$name" "missing" "$need" "$hint"
    return
  fi
  local version
  version=$(limited "$@" 2>&1 | grep -E -- "$LINE_FILTER" | first_version)
  if [ -z "$version" ]; then
    fail "$name" "no-version" "$need" "$hint"
  elif [ "$min" = "-" ] || version_ge "$version" "$min"; then
    pass "$name" "$version" "$need"
  else
    fail "$name" "$version" "$need" "$hint"
  fi
}

# check_tool_line PATTERN NAME MINIMUM HINT COMMAND...: like check_tool, but reads the version
# only from the output lines that match PATTERN (govulncheck prints the Go version first).
check_tool_line() {
  LINE_FILTER=$1
  shift
  check_tool "$@"
  LINE_FILTER='.'
}

# check_python: uv must find a Python 3.14 (without downloading one).
check_python() {
  local path version
  if ! command -v uv > /dev/null 2>&1; then
    fail "python" "missing" "need >= 3.14.0" "chapter 2, section 2.8"
    return
  fi
  path=$(uv python find 3.14 2> /dev/null)
  if [ -z "$path" ]; then
    fail "python" "missing" "need >= 3.14.0" "chapter 2, section 2.8"
    return
  fi
  version=$("$path" --version 2>&1 | first_version)
  if [ -n "$version" ] && version_ge "$version" "3.14.0"; then
    pass "python" "$version" "need >= 3.14.0"
  else
    fail "python" "${version:-no-version}" "need >= 3.14.0" "chapter 2, section 2.8"
  fi
}

# check_dir NAME PATH TEST: a course folder must exist and must not live on a Windows drive (/mnt/c).
check_dir() {
  local name=$1 dir=$2 kind=$3
  case "$dir" in
    /mnt/*)
      fail "$name" "on-/mnt" "inside the Linux file system" "chapter 2, section 2.9"
      return
      ;;
  esac
  if [ "$kind" = "repo" ] && [ -d "$dir" ] && git -C "$dir" rev-parse --is-inside-work-tree > /dev/null 2>&1; then
    pass "$name" "git-repo" "$dir"
  elif [ "$kind" = "book" ] && [ -d "$dir/exercises" ]; then
    pass "$name" "found" "$dir"
  else
    fail "$name" "missing" "${dir:-COURSE is not set}" "chapter 2, section 2.9"
  fi
}

platform="$(uname -s) $(uname -m)"
if grep -qi microsoft /proc/version 2> /dev/null; then
  platform="$platform (WSL2)"
fi
echo "doctor: checking the course workbench on $platform"

echo "Day 0 tools:"
check_tool "git" "2.28" "chapter 2, section 2.5" git --version
has_key=0
for key in "$HOME"/.ssh/id_*.pub; do
  [ -e "$key" ] && has_key=1
done
if [ "$has_key" -eq 1 ]; then
  pass "ssh-key" "found" "a public key in ~/.ssh"
else
  fail "ssh-key" "missing" "a public key in ~/.ssh" "chapter 2, section 2.5"
fi
check_tool "docker" "29.0" "chapter 2, section 2.4" docker --version
check_tool "docker-engine" "29.0" "chapter 2, section 2.4 (is Docker Desktop running?)" \
  docker info --format '{{.ServerVersion}}'
check_tool "compose" "2.20" "chapter 2, section 2.4" docker compose version --short
check_tool "code" "-" "chapter 2, section 2.6" code --version
check_tool "go" "1.27.1" "chapter 2, section 2.7" go version
check_tool "uv" "0.12.23" "chapter 2, section 2.8" uv --version
check_python

echo "Go tools (go install, Day 0):"
check_tool "staticcheck" "2026.2.1" "chapter 2, section 2.7" staticcheck -version
check_tool "goose" "3.28.0" "chapter 2, section 2.7" goose -version
check_tool "buf" "1.73.0" "chapter 2, section 2.7" buf --version
check_tool "protoc-gen-go" "1.36.12" "chapter 2, section 2.7" protoc-gen-go --version
check_tool "protoc-gen-go-grpc" "1.6.2" "chapter 2, section 2.7" protoc-gen-go-grpc --version
check_tool_line "Scanner" "govulncheck" "1.8.0" "chapter 2, section 2.7" govulncheck -version
check_tool "actionlint" "1.7.12" "chapter 2, section 2.7" actionlint -version

echo "Course folders:"
check_dir "COURSE" "${COURSE:-}" "book"
check_dir "senior-course" "$HOME/senior-course" "repo"

if [ "$PART_E" -eq 1 ]; then
  echo "Part E tools:"
  check_tool "kind" "0.33.0" "chapter 2, section 2.10" kind version
  check_tool "kubectl" "1.37.1" "chapter 2, section 2.10" kubectl version --client
  check_tool "helm" "4.3.0" "chapter 2, section 2.10" helm version --short
  check_tool "terraform" "1.16.5" "chapter 2, section 2.10" terraform version
  check_tool "k6" "2.3.0" "chapter 2, section 2.10" k6 version
fi

echo "doctor: $OK ok, $FAILED failed"
if [ "$FAILED" -eq 0 ]; then
  echo "doctor: the workbench is ready"
else
  echo "doctor: fix the FAIL lines from the top; each names the section that installs the tool"
fi
exit "$FAILED"
