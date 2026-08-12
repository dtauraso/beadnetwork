#!/usr/bin/env bash



# PLACEMENT: CLAUDE.md,MODEL.md | must not reintroduce tokens from the DEAD_TOKENS list (retired React Flow terms)
set -euo pipefail






SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

DOCS=(CLAUDE.md MODEL.md)


DEAD_TOKENS=(
  "rf/nodes"
  "GenericNode"
  "PUMP_SLOT_HANDLER"
  "webview/schema/"
  "webview/rf/"
)

fail=0
for doc in "${DOCS[@]}"; do
  if [ ! -f "$doc" ]; then
    echo "dead-doc-tokens: MISCONFIGURED — doc not found: $doc (renamed? a missing doc would vacuously pass)" >&2
    exit 1
  fi
  for token in "${DEAD_TOKENS[@]}"; do
    if grep -aqF "$token" "$doc" 2>/dev/null; then
      echo "DEAD TOKEN: '$token' found in $doc — remove or update the reference"
      fail=1
    fi
  done
done

exit $fail
