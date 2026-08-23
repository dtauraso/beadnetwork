#!/usr/bin/env bash

# PLACEMENT: Categories/NodeKinds/*/node.go,Categories/NodeKinds/*/*.go | a node-kind package may import only bead/clock/portwiring, never a sibling kind — the gate, the held driver and the build args all live in the kinds that use them

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

NODES_DIR="$REPO_ROOT/Categories/NodeKinds"
if [ ! -d "$NODES_DIR" ]; then
  echo "check-dep-rules: MISCONFIGURED — $NODES_DIR not found; refusing a vacuous pass." >&2
  echo "  Categories/NodeKinds/ holds every node-kind package this guard exists to police; if it is gone," >&2
  echo "  the invariant it enforces no longer has a home. Update the guard deliberately." >&2
  exit 1
fi

MODULE="github.com/dtauraso/wirefold"

is_spine() { [ "$1" = "bead" ] || [ "$1" = "spatial" ] || [ "$1" = "rowevent" ] || [ "$1" = "clock" ] || [ "$1" = "portwiring" ]; }

is_kind() { grep -rq 'BuilderFor("' "$1" --include="*.go" 2>/dev/null; }

KINDS_SEEN=0
fail=0
for dir in "$NODES_DIR"/*/; do
  kind="$(basename "$dir")"
  is_spine "$kind" && continue
  is_kind "$dir" || continue
  KINDS_SEEN=$((KINDS_SEEN + 1))

  imported="$(grep -rhoE "\"$MODULE/Categories/NodeKinds/[A-Za-z0-9_]+(/[A-Za-z0-9_/]+)?\"" "$dir" --include="*.go" 2>/dev/null \
      | sed -E "s#\"$MODULE/Categories/NodeKinds/([A-Za-z0-9_]+)(/[A-Za-z0-9_/]+)?\"#\1#" | sort -u || true)"

  while IFS= read -r dep; do
    [ -z "$dep" ] && continue
    [ "$dep" = "$kind" ] && continue
    is_spine "$dep" && continue
    echo "ILLEGAL DEP: Categories/NodeKinds/$kind imports sibling Categories/NodeKinds/$dep — kinds must not import each other"
    fail=1
  done <<< "$imported"
done

if [ "$KINDS_SEEN" -eq 0 ]; then
  echo "check-dep-rules: MISCONFIGURED — found no node-kind packages under $NODES_DIR." >&2
  echo "  A kind is a package calling BuilderFor(...); if that call site changed shape, this" >&2
  echo "  guard now polices nothing while still exiting 0. Update is_kind() deliberately." >&2
  exit 1
fi

exit $fail
