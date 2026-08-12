#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: Buffer/buffer_layout_gen.go,tools/gen-node-defs/buflayout/buffer_layout.go | SetOverlayRow must take one named OverlayRow struct, never positional same-typed scalars

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

GEN_FILE="$REPO_ROOT/Buffer/buffer_layout_gen.go"
GENERATOR="$REPO_ROOT/tools/gen-node-defs/buflayout/buffer_layout.go"

for f in "$GEN_FILE" "$GENERATOR"; do
  if [[ ! -f "$f" ]]; then
    echo "check-overlay-row-struct: MISCONFIGURED — file not found: $f" >&2
    exit 1
  fi
done

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
