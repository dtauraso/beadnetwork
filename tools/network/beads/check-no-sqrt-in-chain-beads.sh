#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: nodes/wire/live_beads.go | no math.Sqrt/.Length()/.Normalize(); bead placement stays a fraction along the edge's own segment

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
  echo "  (math.Sqrt / Vec3.Length() / Vec3.Normalize() are each a sqrt — use BeadFraction and Lerp on the stored segment instead."
  echo "   An edge has ONE length: the integer bead-step count its source node computed. A sqrt here measures a second,"
  echo "   independently-derived length that will disagree with the lattice the beads are laid out on, and the bead ends up"
  echo "   timed to cross a distance the chain is not built to. See memory/feedback/architecture/geometry/feedback_abc_times_constant_not_rederive.md.)"
  exit 1
fi

echo "✓ no sqrt in live_beads.go: a bead's position stays a fraction along its edge's own segment."
