#!/usr/bin/env bash
# Build a tiny test image, run dive --json, and assert dive-mcp can parse the output.
set -euo pipefail

DIVE_BIN="${1:?usage: verify-dive-compat.sh <path-to-dive>}"
IMAGE="${DIVE_MCP_TEST_IMAGE:-dive-mcp-compat:ci}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$(mktemp)"
trap 'rm -f "$OUT"' EXIT

echo "Building test image ${IMAGE}..."
docker build -t "$IMAGE" -f "${ROOT}/testdata/Dockerfile.ci" "${ROOT}/testdata"

echo "Running dive against ${IMAGE}..."
"$DIVE_BIN" "$IMAGE" --json "$OUT"

echo "Verifying JSON parses with dive-mcp types..."
cd "$ROOT"
DIVE_JSON_OUTPUT="$OUT" go test -tags=integration -count=1 -run TestDiveJSONCompatibility ./internal/dive/
