#!/usr/bin/env bash

# PLACEMENT: Categories/Node/**/*.go,**/*.ts | no DOUBLE-LINK node-node polar record (LocalPolar et al. — the edge's own triple D is the model, not this); a bead streams its world position, never a node-local offset to be summed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

fail=0
SELF="$(basename "$0")"

BANNED_SYMBOLS='\bLocalPolar\b|\bLayoutHolder\b|\bSetLocalPolar\(|\bLocalPolarsSnapshot\(|\bLoadLocalPolars\(|\brequantizeLocalPolars\(|\brequantizePoleTraced\(|\bneighborSetCRequantize\('

hits=$(
  find Node NodeKinds Categories/Ring/Bead -name '*.go' -print0 2>/dev/null \
    | xargs -0 awk '
        {
          line = $0
          idx = index(line, "//")
          if (idx > 0) line = substr(line, 1, idx - 1)
          print FILENAME ":" FNR ":" line
        }
      ' \
    | grep -v "/${SELF}:" \
    | grep -E "$BANNED_SYMBOLS" || true
)

if [ -n "$hits" ]; then
  echo "✗ the REJECTED node-node polar record reappeared — MODEL.md \"the polar model\"."
  echo
  echo "  What is banned is the DOUBLE-LINK form: each endpoint of an edge holding its OWN"
  echo "  quantised bearing/distance to the other node, re-quantized node-to-node on every"
  echo "  drag. That was two half-authoritative records for one position, and was measured"
  echo "  disagreeing with the node's own scene polar (node 3 by +3.24 world units, node 4"
  echo "  by -3.08, against a step of 8.96)."
  echo
  echo "  What is NOT banned, and is now the model, is the edge's own triple D: A + D = B,"
  echo "  component by component, stored ONCE in the edge file under its source and owned by"
  echo "  that edge's edgeMover (owners.Deltas, loadspec/edge_delta.go). A target's copy of"
  echo "  it is what the source TOLD it, is never persisted, and never answers \"where is that"
  echo "  node\" — the loader asserts the triangle closes on load, per component."
  echo
  echo "  Delete the record below, do not re-add it:"
  echo "$hits"
  fail=1
fi

TS_DIR="."

sum_hits=$(grep -rlnE 'readChainBeadO[XYZ]\(|readEdgeBeadO[XYZ]\(' \
  --include='*.ts' --include='*.tsx' \
  "$TS_DIR" 2>/dev/null || true)

if [ -n "$sum_hits" ]; then
  echo "✗ a node-local bead offset is being read again, which reintroduces the"
  echo "  summation (some centre + bead offset -> absolute bead centre) this guard"
  echo "  exists to keep at one site. A bead streams its WORLD position, placed by"
  echo "  the edge it travels:"
  printf '%s\n' "$sum_hits"
  fail=1
fi

if [ "$fail" -eq 0 ]; then
  echo "✓ no node-node polar record; exactly one bead-centre summation site."
fi

exit "$fail"
