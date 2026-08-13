#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: nodes/wire/live_beads.go | no math.Sqrt/.Length()/.Normalize(); bead placement stays a fraction along the edge's own segment

# The chain a NODE laid toward a neighbour is gone, and so is the file this
# guard used to read. Placement now happens where the beads actually are — on
# the edge's own segment, in nodes/wire/live_beads.go — and the invariant is
# unchanged: a bead's position is arithmetic on values already held, never a
# cartesian distance measured on the spot.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
FILE="$REPO_ROOT/nodes/wire/live_beads.go"

if [ ! -f "$FILE" ]; then
  echo "✗ no-sqrt-in-chain-beads: MISCONFIGURED — file not found: $FILE" >&2
  exit 1
fi

PATTERN='math\.Sqrt|\.Length\(\)|\.Normalize\(\)'

hits=$(grep -nE "$PATTERN" "$FILE" || true)

if [ -n "$hits" ]; then
  echo "✗ sqrt fingerprint(s) found in live_beads.go — a bead's position must stay a fraction along the segment its edge already holds, never a cartesian distance:"
  echo "$hits"
  echo "  (math.Sqrt / Vec3.Length() / Vec3.Normalize() are each a sqrt — use BeadFraction and Lerp on the stored segment instead. See docs/bead-model/beads-are-the-edge.md and memory/feedback/architecture/geometry/feedback_abc_times_constant_not_rederive.md.)"
  exit 1
fi

echo "✓ no sqrt in live_beads.go: a bead's position stays a fraction along its edge's own segment."
