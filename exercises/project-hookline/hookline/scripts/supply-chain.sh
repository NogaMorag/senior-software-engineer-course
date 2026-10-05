#!/usr/bin/env bash
# Release hygiene of milestone M9 (chapter 28, and chapter 19's scan_images.sh): govulncheck on the Go
# module, then for each image an SBOM with Syft and a scan of that SBOM with Grype. The scanners run as
# containers, and each image is exported with docker save first, so none of them needs the Docker
# socket. Reports land in $OUT.
#   make images && scripts/supply-chain.sh
# Exits non-zero when govulncheck finds a vulnerability the code calls, or Grype finds one rated
# FAIL_ON (default high) or worse that has a fix. Findings without a fix are listed but do not fail the
# gate (chapter 19: a gate that fails on what nobody can fix gets bypassed). A fixed finding in a base
# image's package means a newer base image, not a lower threshold. Needs: docker, go.
set -euo pipefail
cd "$(dirname "$0")/.."

OUT=${OUT:-/tmp/hookline-supply-chain}
FAIL_ON=${FAIL_ON:-high}
SYFT=anchore/syft:v1.54.0
GRYPE=anchore/grype:v0.120.0
mkdir -p "$OUT"
status=0

echo "== govulncheck (the module's code, only what it calls)"
go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./... | tee "$OUT/govulncheck.txt" || status=1

for img in hookline:dev hookline-sink:dev; do
  name=${img%%:*}
  echo "== $img"
  docker save "$img" -o "$OUT/$name.tar"

  # SBOM: every package in the image, in SPDX JSON. Keep it with the release: a later scan reads it.
  docker run --rm -v "$OUT:/work" "$SYFT" "docker-archive:/work/$name.tar" -o "spdx-json=/work/$name.spdx.json" -q
  echo "sbom   $OUT/$name.spdx.json ($(grep -o '"SPDXID"' "$OUT/$name.spdx.json" | wc -l | tr -d ' ') SPDX elements)"

  # Grype reads the SBOM, not the image, and fails on a fixed FAIL_ON or worse; the table lists them all.
  if ! docker run --rm -v "$OUT:/work" -v hookline-grype-db:/root/.cache/grype "$GRYPE" \
    "sbom:/work/$name.spdx.json" -o table --file "/work/$name.grype.txt" --fail-on "$FAIL_ON" --only-fixed -q; then
    echo "grype  findings rated $FAIL_ON or worse, with a fix, in $img" >&2
    status=1
  fi
  echo "grype  $OUT/$name.grype.txt"
  rm "$OUT/$name.tar"
done
exit "$status"
