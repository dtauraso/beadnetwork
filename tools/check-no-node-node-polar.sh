#!/usr/bin/env bash
# check-no-node-node-polar.sh — guard for the polar model (MODEL.md "the polar model"):
# a node has ONE polar coordinate, about the scene sphere's centre only. It carries NO
# stored coordinate for a NEIGHBOUR node, and a bead's absolute centre is never written
# as an independent position — it is derived by summation at exactly one site (node
# world centre + node-local bead offset).
#
# This guard checks two things a behavioural test cannot distinguish from a correct
# implementation:
#
#   1. No node-node polar record. wire.LocalPolar (a node's quantised record of a
#      NEIGHBOUR's bearing/distance) and its machinery (the LocalPolar type, its
#      LayoutHolder container, SetLocalPolar/LocalPolarsSnapshot/LoadLocalPolars, and
#      the requantize call sites that kept it in sync — requantizeLocalPolars/
#      requantizePoleTraced/neighborSetCRequantize) must never reappear as real CODE
#      (a type/field/function reference) anywhere in nodes/ or Buffer/. Historical prose
#      naming the deleted symbols (comments explaining what was removed and why) is not
#      flagged — only comment LINES are stripped first, so this cannot be satisfied by
#      just talking about the deletion.
#   2. One bead-centre summation site. Exactly one production call site may add a
#      node's world centre to a chain-bead's node-local offset to produce an absolute
#      bead position — node-stream-blocks.ts's getChainBeads (Go streams the node
#      centre and each bead's NODE-LOCAL offset as separate columns on purpose; TS
#      composes them at decode time, matching every other node-local buffer column). A
#      SECOND site doing that same addition would mean a bead centre is being computed
#      (and potentially cached/stored) somewhere else — the independent-absolute-
#      position shape this guard exists to catch. Scoped to the actual arithmetic (a
#      value summed with a readChainBeadOX/OY/OZ call), not merely importing/defining
#      the reader (the schema/decode files that declare or forward those readers are not
#      themselves a second summation site).
#
# Exit 0 clean, exit 1 with a report.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

fail=0
SELF="$(basename "$0")"

# --- 1. No node-node polar record ---------------------------------------------------
# Strip full-line and trailing `//` comments before matching, so historical prose
# ("wire.LocalPolar ... was deleted") cannot trip this — only a real code reference to
# the type/container/functions can.
BANNED_SYMBOLS='\bLocalPolar\b|\bLayoutHolder\b|\bSetLocalPolar\(|\bLocalPolarsSnapshot\(|\bLoadLocalPolars\(|\brequantizeLocalPolars\(|\brequantizePoleTraced\(|\bneighborSetCRequantize\('

hits=$(
  grep -rn --include='*.go' -E '.' nodes/ Buffer/ 2>/dev/null \
    | grep -v "/${SELF}:" \
    | sed -E 's#^([^:]+:[0-9]+:).*$#\1 __SPLIT__ &#' \
    | while IFS= read -r line; do
        path_line="${line%% __SPLIT__*}"
        rest="${line#*__SPLIT__ }"
        code="${rest#*:}"
        code="${code#*:}"
        # Drop a trailing // comment and skip lines that are ENTIRELY comment.
        code_nocomment="${code%%//*}"
        printf '%s:%s\n' "$path_line" "$code_nocomment"
      done \
    | grep -E "$BANNED_SYMBOLS" || true
)

if [ -n "$hits" ]; then
  echo "✗ node-node polar record reappeared — MODEL.md \"the polar model\": a node has ONE"
  echo "  polar coordinate (about the scene centre), never a stored coordinate for a"
  echo "  neighbour. Delete the record, do not re-add it:"
  echo "$hits"
  fail=1
fi

# --- 2. One bead-centre summation site ----------------------------------------------
TS_DIR="tools/topology-vscode/src"
# The actual danger shape: a variable summed with a decoded chain-bead offset, e.g.
# `cx + readChainBeadOX(...)`. Importing/declaring the reader alone (schema/decode
# files) is not a summation.
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
