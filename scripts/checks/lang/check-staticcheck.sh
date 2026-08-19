#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: none | universal Go vet/staticcheck hygiene, not a placement decision

cd "$(git rev-parse --show-toplevel)"

if ! vet_out=$(go vet ./... 2>&1); then
  echo "go vet failed:" >&2
  echo "$vet_out" >&2
  exit 1
fi

staticcheck_bin=""
if command -v staticcheck >/dev/null 2>&1; then
  staticcheck_bin="$(command -v staticcheck)"
elif [ -x "$(go env GOPATH)/bin/staticcheck" ]; then
  staticcheck_bin="$(go env GOPATH)/bin/staticcheck"
fi

if [ -z "$staticcheck_bin" ]; then
  echo "staticcheck not installed; skipping (install with:" >&2
  echo "  go install honnef.co/go/tools/cmd/staticcheck@latest)" >&2
  exit 0
fi

if ! sc_out=$("$staticcheck_bin" ./... 2>&1); then
  echo "staticcheck reported findings:" >&2
  echo "$sc_out" >&2
  exit 1
fi

exit 0
