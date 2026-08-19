#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: src/Buffer/buffer_layout_gen_singletons.go,cmd/gen-node-defs/buflayout/buffer_layout_singletons.go | SetOverlayRow must take one named OverlayRow struct, never positional same-typed scalars

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"

GEN_FILE="$REPO_ROOT/src/Buffer/buffer_layout_gen_singletons.go"
GENERATOR="$REPO_ROOT/cmd/gen-node-defs/buflayout/buffer_layout_singletons.go"

for f in "$GEN_FILE" "$GENERATOR"; do
  if [[ ! -f "$f" ]]; then
    echo "check-overlay-row-struct: MISCONFIGURED — file not found: $f" >&2
    exit 1
  fi
done

if ! grep -q "func SetOverlayRow(" "$GEN_FILE"; then
  echo "check-overlay-row-struct: no Overlay row (moved to column channels); nothing to check"
  exit 0
fi

FAIL=0

if ! grep -qE '^func SetOverlayRow\(buf \[\]byte, row OverlayRow\) \{$' "$GEN_FILE"; then
  echo "check-overlay-row-struct: SetOverlayRow in $GEN_FILE is not the expected"
  echo "  'func SetOverlayRow(buf []byte, row OverlayRow)' named-struct signature."
  echo "  A positional-scalar signature reintroduces the transposition hazard."
  FAIL=1
fi

if ! grep -q 'type OverlayRow struct {' "$GEN_FILE"; then
  echo "check-overlay-row-struct: OverlayRow struct type not found in $GEN_FILE"
  FAIL=1
fi

if [[ $FAIL -ne 0 ]]; then
  exit 1
fi

echo "check-overlay-row-struct: clean"
exit 0
