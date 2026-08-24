#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
gofmt -w .
go test ./...
lines=$(find . -name '*.go' ! -name '*_test.go' -print0 | xargs -0 cat | wc -l)
test "$lines" -ge 2000
echo "non-test Go source lines: $lines"
