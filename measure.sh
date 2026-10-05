#!/usr/bin/env bash
# Reproduce the before/after numbers in CASE-STUDY.md.
# Requires: docker, and (optional) trivy for vuln counts.
# Usage: ./measure.sh
set -euo pipefail

cd "$(dirname "$0")"
BEFORE_IMG="orders-api:before"
AFTER_IMG="orders-api:after"

echo "==> Building BEFORE image (naive, single-stage)…"
t0=$(date +%s)
docker build -q -f before/Dockerfile -t "$BEFORE_IMG" . >/dev/null
echo "    cold build time: $(($(date +%s) - t0))s"

echo "==> Building AFTER image (multi-stage, distroless)…"
t0=$(date +%s)
docker build -q -t "$AFTER_IMG" . >/dev/null
echo "    cold build time: $(($(date +%s) - t0))s"

size() { docker image inspect "$1" --format '{{.Size}}' | awk '{printf "%.0f MB", $1/1024/1024}'; }

echo
echo "================  IMAGE SIZE  ================"
printf "  before : %s\n" "$(size "$BEFORE_IMG")"
printf "  after  : %s\n" "$(size "$AFTER_IMG")"

if command -v trivy >/dev/null 2>&1; then
  echo
  echo "============  VULNERABILITIES  =============="
  count() { trivy image --quiet --severity HIGH,CRITICAL --format json "$1" \
    | grep -o '"Severity"' | wc -l | tr -d ' '; }
  printf "  before HIGH+CRITICAL : %s\n" "$(count "$BEFORE_IMG")"
  printf "  after  HIGH+CRITICAL : %s\n" "$(count "$AFTER_IMG")"
else
  echo
  echo "(trivy not installed — skipping vuln counts. Install: https://trivy.dev)"
fi

echo
echo "Done. Drop these numbers into CASE-STUDY.md."
