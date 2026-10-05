#!/usr/bin/env bash
# Scans the chapter 19 images without installing anything: Syft, Grype and
# Trivy run as containers. Each image is exported with docker save first, so no
# scanner needs the Docker socket (a container holding the socket is root on
# the host). Reports land in reports/. Run after check_images.sh:
#   bash tests/scan_images.sh
set -euo pipefail
cd "$(dirname "$0")/.."

SYFT=anchore/syft:v1.54.0
GRYPE=anchore/grype:v0.120.0
TRIVY=aquasec/trivy:0.75.0
mkdir -p reports

for img in shorty:dev receiver:dev; do
  name=${img%%:*}
  echo "== $img"
  docker save "$img" -o "reports/$name.tar"

  # SBOM: the list of every package in the image, in SPDX JSON.
  docker run --rm -v "$PWD/reports:/work" "$SYFT" \
    "docker-archive:/work/$name.tar" -o "spdx-json=/work/$name.spdx.json" -q
  echo "sbom   reports/$name.spdx.json ($(grep -o '"SPDXID"' "reports/$name.spdx.json" | wc -l | tr -d ' ') SPDX elements)"

  # Grype reads the SBOM, not the image: a scan you can repeat later without the image.
  docker run --rm -v "$PWD/reports:/work" -v ch19-grype-db:/root/.cache/grype "$GRYPE" \
    "sbom:/work/$name.spdx.json" -o table --file "/work/$name.grype.txt" -q || true
  echo "grype  reports/$name.grype.txt ($(($(wc -l < "reports/$name.grype.txt") - 1)) findings, all severities)"

  # Trivy scans the image archive itself; only HIGH and CRITICAL, fixed or not.
  docker run --rm -v "$PWD/reports:/work" -v ch19-trivy-cache:/root/.cache/trivy "$TRIVY" \
    image --quiet --input "/work/$name.tar" --severity HIGH,CRITICAL \
    --format table --output "/work/$name.trivy.txt"
  echo "trivy  reports/$name.trivy.txt"

  rm "reports/$name.tar"
done
