#!/usr/bin/env bash
set -euo pipefail

# check-bead-idle-costs-zero.sh — source guard for GAP 3/4 of task/beads-own-positions-fixed
# (PLAN.md "idle CPU at zero, machine-speed burst on drag"): chainBeads must gate its
# geometry broadcast on the node's OWN per-drag flag (nodeMover.dragging), not issue a
# broadcast unconditionally on every call and not hardcode the dragging argument to a
# constant (which would make the two clock modes behaviourally identical again — the exact
# defect this fix closes). The change-detection itself (aim/count/lit-set comparison) lives
# inside broadcastAndRead (bead_actor_bridge.go) and is exercised behaviourally by
# TestIdleIssuesNoBroadcastWhenNothingChanged / TestDraggingBroadcastsEveryCall
# (bead_chain_test.go); this guard is the SOURCE half — a call site that silently drops the
# real flag renders identically to a correct one on a quick read.
#
# Exit 0 clean, exit 1 with a report.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHAIN_FILE="$REPO_ROOT/nodes/Wiring/chain_beads.go"
BRIDGE_FILE="$REPO_ROOT/nodes/Wiring/bead_actor_bridge.go"

for f in "$CHAIN_FILE" "$BRIDGE_FILE"; do
  if [ ! -f "$f" ]; then
    echo "✗ bead-idle-costs-zero: MISCONFIGURED — file not found: $f" >&2
    exit 1
  fi
done

fail=0

# chain_beads.go's call site must pass m.dragging (the real per-drag flag), never a
# hardcoded true/false.
call_line=$(grep -nE '\.broadcastAndRead\(' "$CHAIN_FILE" || true)
if [ -z "$call_line" ]; then
  echo "✗ chain_beads.go no longer calls broadcastAndRead — either the bead-actor read path regressed, or this guard is stale." >&2
  exit 1
fi
if ! printf '%s\n' "$call_line" | grep -q 'm\.dragging'; then
  echo "✗ chain_beads.go's broadcastAndRead call site does not pass m.dragging — GAP 3/4 requires the geometry broadcast to be gated on the node's own per-drag flag, not a hardcoded constant (which makes dragging vs. idle behaviourally identical again)." >&2
  printf '%s\n' "$call_line" >&2
  fail=1
fi

# broadcastAndRead itself must actually branch on its dragging parameter (not silently
# drop it) — the fingerprint of the real gate is `dragging ||` immediately guarding the
# BroadcastGeometry call.
body=$(awk '/^func \(ea \*beadEdgeActors\) broadcastAndRead\(/{flag=1} flag{print} /^func /{if (flag && !/broadcastAndRead/) exit}' "$BRIDGE_FILE")
if ! printf '%s\n' "$body" | grep -qE '\bdragging\b.*\|\|.*haveGeom'; then
  echo "✗ broadcastAndRead no longer gates BroadcastGeometry on its dragging parameter alongside the change-detection cache (expected a \`dragging || !ea.haveGeom || ...\` shape) — idle would cost a broadcast on every call again (PLAN.md 'idle CPU at zero')." >&2
  fail=1
fi

if [ "$fail" != "0" ]; then
  exit 1
fi

echo "✓ idle costs zero: chain_beads.go's broadcastAndRead call site passes the real m.dragging flag, and broadcastAndRead gates its geometry broadcast on dragging alongside the change-detection cache."
