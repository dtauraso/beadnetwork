#!/usr/bin/env bash
set -euo pipefail

# check-staticcheck.sh — Go static-analysis guard (the Go equivalent of


# PLACEMENT: none | universal Go vet/staticcheck hygiene, not a placement decision




#   2. `staticcheck ./...` — a dev tool that must be installed separately. Policy:


#               the dev tool; staticcheck is enforced whenever it is present


cd "$(git rev-parse --show-toplevel)"


if ! vet_out=$(go vet ./... 2>&1); then
  echo "go vet failed:" >&2
  echo "$vet_out" >&2
  exit 1
fi

# --- Layer 2: staticcheck (found → enforce, absent → warn+skip) ---

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
