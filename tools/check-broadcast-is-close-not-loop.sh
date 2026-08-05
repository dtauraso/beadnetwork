#!/usr/bin/env bash
set -euo pipefail

# check-broadcast-is-close-not-loop.sh — guard that BeadWakeGroup wakes/settles/broadcasts
# its beads with a single close (BroadcastChain.Advance), never a send-loop over N beads.
#
# PLAN.md "the wake is one operation, not N": a send-loop iterating a group's beads is
# behaviourally identical to a single close (every bead still gets woken) but SCALES WITH N
# — exactly the cost the close-based BroadcastChain exists to avoid. Invisible to a
# behavioural test with a small bead count, so this is a source guard: BeadWakeGroup's
# StartDrag/EndDrag/BroadcastGeometry methods (nodes/wire/bead_wake_group.go) must each call
# Advance/AdvanceWithValue and must not contain a `for`/`range` loop.
#
# PLACEMENT: nodes/wire/bead_wake_group.go | StartDrag/EndDrag/BroadcastGeometry must call Advance, never loop over beads
#
# Exit 0 clean, exit 1 with a report.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FILE="$REPO_ROOT/nodes/wire/bead_wake_group.go"

if [ ! -f "$FILE" ]; then
  echo "✗ broadcast-is-close-not-loop: MISCONFIGURED — file not found: $FILE" >&2
  exit 1
fi

fail=0
for fn in StartDrag EndDrag BroadcastGeometry; do
  body=$(awk -v fn="$fn" '
    index($0, "func (g *BeadWakeGroup) " fn "(") == 1 {inFn=1}
    inFn {print; if (/^}/) exit}
  ' "$FILE")
  if [ -z "$body" ]; then
    echo "✗ broadcast-is-close-not-loop: MISCONFIGURED — could not find func $fn on BeadWakeGroup in $FILE" >&2
    fail=1
    continue
  fi
  if echo "$body" | grep -qE '\bfor\b'; then
    echo "✗ BeadWakeGroup.$fn contains a for/range loop — a wake/settle/geometry broadcast must"
    echo "  be a single BroadcastChain.Advance/AdvanceWithValue call (one close wakes every"
    echo "  waiting bead at once), never a loop sending to each bead — that scales with N."
    fail=1
  fi
  if ! echo "$body" | grep -qE '\.Advance(WithValue)?\('; then
    echo "✗ BeadWakeGroup.$fn does not call Advance/AdvanceWithValue — it must broadcast via"
    echo "  BroadcastChain's single-close primitive."
    fail=1
  fi
done

if [ "$fail" -ne 0 ]; then
  exit 1
fi

echo "✓ BeadWakeGroup wake/settle/geometry broadcasts are single closes, not send-loops (nodes/wire/bead_wake_group.go)."
