#!/usr/bin/env bash
set -euo pipefail

# check-no-sqrt-in-chain-beads.sh — guard that chain-bead placement stays index arithmetic.
#
# docs/beads-are-the-edge.md's chain beads sit on the SAME local polar lattice a node's own
# LocalPolar triples describe (nodes/wire/layout_holder.go). A bead's distance from its
# source node comes from an integer index (QuantIR) times a step constant (StepR) —
# multiplication, per memory/feedback_abc_times_constant_not_rederive.md — never from a
# cartesian offset's square root. Both math.Sqrt directly and Vec3's .Length()/.Normalize()
# (each internally a sqrt, nodes/wire/geometry.go) are banned in this one file: a distance
# computed by sqrt means someone reintroduced a cartesian shortcut that can drift from the
# lattice the nodes actually sit on, silently reproducing the bug this design exists to rule
# out. Trig (sin/cos/atan2/acos) is NOT banned — it is expected at the one
# cartesian<->polar boundary conversion per bead (polar2cart).
#
# Exit 0 clean, exit 1 with a report.
#
# PLACEMENT: nodes/Wiring/chain_beads.go | no math.Sqrt/.Length()/.Normalize(); bead placement stays index arithmetic (QuantIR*StepR)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FILE="$REPO_ROOT/nodes/Wiring/chain_beads.go"

if [ ! -f "$FILE" ]; then
  echo "✗ no-sqrt-in-chain-beads: MISCONFIGURED — file not found: $FILE" >&2
  exit 1
fi

PATTERN='math\.Sqrt|\.Length\(\)|\.Normalize\(\)'

hits=$(grep -nE "$PATTERN" "$FILE" || true)

if [ -n "$hits" ]; then
  echo "✗ sqrt fingerprint(s) found in chain_beads.go — bead placement must stay index arithmetic on the node's own local polar lattice, never a cartesian distance:"
  echo "$hits"
  echo "  (math.Sqrt / Vec3.Length() / Vec3.Normalize() are each a sqrt — use QuantIR*StepR and fromAxisFrame instead. See docs/beads-are-the-edge.md and memory/feedback_abc_times_constant_not_rederive.md.)"
  exit 1
fi

echo "✓ no sqrt in chain_beads.go: bead placement stays index arithmetic on the local polar lattice."
