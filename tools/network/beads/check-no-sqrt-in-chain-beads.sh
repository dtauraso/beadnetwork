#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: nodes/Wiring/nodeactor/chain_beads.go | no math.Sqrt/.Length()/.Normalize(); bead placement stays index arithmetic (QuantIR*StepR)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
FILE="$REPO_ROOT/nodes/Wiring/nodeactor/chain_beads.go"

if [ ! -f "$FILE" ]; then
  echo "✗ no-sqrt-in-chain-beads: MISCONFIGURED — file not found: $FILE" >&2
  exit 1
fi

PATTERN='math\.Sqrt|\.Length\(\)|\.Normalize\(\)'

hits=$(grep -nE "$PATTERN" "$FILE" || true)

if [ -n "$hits" ]; then
  echo "✗ sqrt fingerprint(s) found in chain_beads.go — bead placement must stay index arithmetic on the node's own local polar lattice, never a cartesian distance:"
  echo "$hits"
  echo "  (math.Sqrt / Vec3.Length() / Vec3.Normalize() are each a sqrt — use QuantIR*StepR and fromAxisFrame instead. See docs/bead-model/beads-are-the-edge.md and memory/feedback_abc_times_constant_not_rederive.md.)"
  exit 1
fi

echo "✓ no sqrt in chain_beads.go: bead placement stays index arithmetic on the local polar lattice."
