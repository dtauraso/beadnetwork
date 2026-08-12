#!/usr/bin/env bash

# PLACEMENT: nodes/*/node.go,nodes/*/*.go | a node-kind package may import only the shared spine (Wiring/gatecommon/wire), never a sibling kind























set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

NODES_DIR="$REPO_ROOT/nodes"
if [ ! -d "$NODES_DIR" ]; then
  echo "check-dep-rules: MISCONFIGURED — $NODES_DIR not found; refusing a vacuous pass." >&2
  echo "  nodes/ holds every node-kind package this guard exists to police; if it is gone," >&2
  echo "  the invariant it enforces no longer has a home. Update the guard deliberately." >&2
  exit 1
fi

MODULE="github.com/dtauraso/wirefold"


is_spine() { [ "$1" = "Wiring" ] || [ "$1" = "gatecommon" ] || [ "$1" = "wire" ]; }

fail=0
for dir in "$NODES_DIR"/*/; do
  kind="$(basename "$dir")"
  is_spine "$kind" && continue







  imported="$(grep -rhoE "\"$MODULE/nodes/[A-Za-z0-9_]+(/[A-Za-z0-9_/]+)?\"" "$dir" --include="*.go" 2>/dev/null \
      | sed -E "s#\"$MODULE/nodes/([A-Za-z0-9_]+)(/[A-Za-z0-9_/]+)?\"#\1#" | sort -u || true)"

  while IFS= read -r dep; do
    [ -z "$dep" ] && continue
    [ "$dep" = "$kind" ] && continue
    is_spine "$dep" && continue
    echo "ILLEGAL DEP: nodes/$kind imports sibling nodes/$dep — kinds must couple only through the shared spine (Wiring/gatecommon)"
    fail=1
  done <<< "$imported"
done

exit $fail
