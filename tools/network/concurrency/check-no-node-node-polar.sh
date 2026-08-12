#!/usr/bin/env bash
































# PLACEMENT: nodes/**/*.go,Buffer/**/*.go,tools/topology-vscode/src/**/*.ts | no node-node polar record (LocalPolar et al.); exactly one bead-centre summation site (node-stream-blocks.ts)

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



SUM_PATTERN='[A-Za-z0-9_.]+[[:space:]]*\+[[:space:]]*readChainBeadO[XYZ]\('

sum_hits=$(grep -rlnE "$SUM_PATTERN" \
  --include='*.ts' --include='*.tsx' \
  "$TS_DIR" 2>/dev/null || true)

sum_count=$(printf '%s\n' "$sum_hits" | grep -c . || true)

if [ "$sum_count" -gt 1 ]; then
  echo "✗ more than one file sums a node centre with a chain-bead node-local offset — the"
  echo "  summation (node world centre + node-local bead offset -> absolute bead centre)"
  echo "  must happen at exactly ONE site (node-stream-blocks.ts's getChainBeads):"
  printf '%s\n' "$sum_hits"
  fail=1
elif [ "$sum_count" -eq 1 ] && ! printf '%s\n' "$sum_hits" | grep -q 'node-stream-blocks\.ts$'; then
  echo "✗ the bead-centre summation site moved out of node-stream-blocks.ts to:"
  printf '%s\n' "$sum_hits"
  echo "  (update this guard's expected path if the move is deliberate and still"
  echo "   exactly one site — do not let a second site appear silently)"
  fail=1
fi

if [ "$fail" -eq 0 ]; then
  echo "✓ no node-node polar record; exactly one bead-centre summation site."
fi

exit "$fail"
