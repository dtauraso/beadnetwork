#!/usr/bin/env bash

# PLACEMENT: none | universal Go formatting, not a placement decision (gofmt every .go file)
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

unformatted="$(gofmt -l . 2>/dev/null | grep -vE '(^|/)(vendor|node_modules)/' || true)"

if [ -n "$unformatted" ]; then
  echo "gofmt: the following Go files are not formatted (run 'gofmt -w .'):" >&2
  echo "$unformatted" >&2
  exit 1
fi

echo "check-gofmt: clean"
