#!/usr/bin/env bash

# PLACEMENT: nodes/**/*.go,Buffer/**/*.go,tools/topology-vscode/src/**/*.ts | no node-node polar record (LocalPolar et al.); a bead streams its world position, never a node-local offset to be summed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

fail=0
SELF="$(basename "$0")"

BANNED_SYMBOLS='\bLocalPolar\b|\bLayoutHolder\b|\bSetLocalPolar\(|\bLocalPolarsSnapshot\(|\bLoadLocalPolars\(|\brequantizeLocalPolars\(|\brequantizePoleTraced\(|\bneighborSetCRequantize\('

hits=$(
  find nodes Buffer -name '*.go' -print0 2>/dev/null \
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
  echo "✗ node-node polar record reappeared — MODEL.md \"the polar model\": a node has ONE"
  echo "  polar coordinate (about the scene centre), never a stored coordinate for a"
  echo "  neighbour. Delete the record, do not re-add it:"
  echo "$hits"
  fail=1
fi

TS_DIR="tools/topology-vscode/src"

# There is no bead-centre summation left to police. A bead's offset used to be
# stored NODE-LOCAL, so somewhere had to add it to that node's world centre,
# and the whole risk was that "somewhere" becoming two places. Beads are now
# placed by the edge they travel, on the segment that edge already holds, and
# what streams is the world position itself — no offset, no origin, no sum.
#
# What the guard still watches is the offset ever coming back: a node-local
# bead column would bring the summation with it.
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
