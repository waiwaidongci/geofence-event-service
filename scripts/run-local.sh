#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
HTTP_ADDR="${HTTP_ADDR:-:8087}" go run ./cmd/geofenced
