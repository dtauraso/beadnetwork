#!/usr/bin/env bash
set -euo pipefail

# check-bead-colour-not-central.sh — sibling of check-bead-position-not-central.sh, for
# COLOUR instead of position. PLAN.md: a woken bead does exactly two things, move (position)
# and send its own colour (lit/carried value) — GAP 2 of task/beads-own-positions-fixed
# closed this: chain_beads.go used to compute litIdx and fill the lit/litVal output slices
# itself directly from that map; now it only computes the INPUT (litIdx, edge-owned
# traversal progress) and hands it to the beads via one BroadcastAnim hop, then reads
# Lit/LitVal back from each bead's own snapshot, same as Position already worked.
#
# This guard is a source-shape check for the same reason bead-position-not-central.sh is:
# a replacement and a mirror render identically, so only source tells them apart.
#
#   1. chain_beads.go's output-row loop reads `s.Lit`/`s.LitVal` off a bead snapshot — never
#      builds `l`/`v` straight from the litIdx map it computed (e.g. `v, isLit := litIdx[i]`
#      immediately followed by using isLit/v for the output row).
#   2. chain_beads.go calls broadcastAndRead with a lit-set argument (the bridge, not
#      chain_beads.go itself, is what turns that into a BroadcastAnim call) — i.e.
#      chain_beads.go must not itself call BroadcastAnim or construct a BeadAnimIn.
#   3. bead_actor_bridge.go's broadcastAndRead calls ea.group.BroadcastAnim (one hop, not a
#      per-bead loop) when the lit set changed.
#   4. beadAnimationState (nodes/wire/bead_actor.go) has exactly one write site for lit/litVal
#      (beadAnimationState.tick) — colour has exactly one writer, same "disjoint writers"
#      property PLAN.md requires of position.
#
# Exit 0 clean, exit 1 with a report.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHAIN_FILE="$REPO_ROOT/nodes/Wiring/chain_beads.go"
BRIDGE_FILE="$REPO_ROOT/nodes/Wiring/bead_actor_bridge.go"
ACTOR_FILE="$REPO_ROOT/nodes/wire/bead_actor.go"

fail=0

for f in "$CHAIN_FILE" "$BRIDGE_FILE" "$ACTOR_FILE"; do
  if [ ! -f "$f" ]; then
    echo "✗ bead-colour-not-central: MISCONFIGURED — file not found: $f" >&2
    exit 1
  fi
done

# Requirement 1: chain_beads.go must not derive the output lit/litVal from litIdx directly
# — the reintroduced-regression fingerprint is a `litIdx[` index expression assigned or
# type-asserted anywhere in the output-append loop. The sound shape reads s.Lit/s.LitVal off
# a snapshot instead.
if grep -nE ':?=\s*litIdx\[' "$CHAIN_FILE"; then
  echo "✗ chain_beads.go still reads litIdx[i] directly to decide an output row's lit/litVal — that decision belongs ONLY to each bead's own goroutine (nodes/wire/bead_actor.go's animIn case), read back via a BeadSnapshot, never taken straight from the map chainBeads itself built (PLAN.md GAP 2)." >&2
  grep -nE ':?=\s*litIdx\[' "$CHAIN_FILE" >&2
  fail=1
fi
if ! grep -q '\.Lit\b' "$CHAIN_FILE" || ! grep -q '\.LitVal\b' "$CHAIN_FILE"; then
  echo "✗ chain_beads.go no longer reads .Lit/.LitVal off a bead snapshot — either it regressed to computing colour itself, or the snapshot read path was removed (PLAN.md GAP 2)." >&2
  fail=1
fi

# Requirement 2: chain_beads.go must not itself call BroadcastAnim or construct a
# BeadAnimIn — that stays inside the bridge (bead_actor_bridge.go), same separation
# bead-position-not-central.sh enforces for BroadcastGeometry/BeadGeometryIn.
if grep -nE '\bBroadcastAnim\b|\bBeadAnimIn\{' "$CHAIN_FILE"; then
  echo "✗ chain_beads.go calls BroadcastAnim or constructs BeadAnimIn directly — that belongs ONLY inside nodes/Wiring/bead_actor_bridge.go's broadcastAndRead (PLAN.md GAP 2)." >&2
  fail=1
fi

# Requirement 3: broadcastAndRead must call BroadcastAnim exactly once (one hop, not a
# per-bead loop).
anim_calls=$(grep -c 'group\.BroadcastAnim(' "$BRIDGE_FILE" || true)
if [ "$anim_calls" != "1" ]; then
  echo "✗ bead_actor_bridge.go: expected exactly one call site for BroadcastAnim (one hop to every bead on this edge); found $anim_calls." >&2
  fail=1
fi

# Requirement 4: beadAnimationState has exactly one write site for lit/litVal — colour has
# exactly one writer, structurally, same as position's "disjoint writers" guard.
lit_writes=$(grep -c 'a\.lit = ' "$ACTOR_FILE" || true)
if [ "$lit_writes" != "1" ]; then
  echo "✗ nodes/wire/bead_actor.go: expected exactly one write site for a bead's lit field (beadAnimationState.tick); found $lit_writes. Colour must have exactly one writer (PLAN.md 'disjoint writers')." >&2
  fail=1
fi

if [ "$fail" != "0" ]; then
  exit 1
fi

echo "✓ bead colour stays actor-owned: no central litIdx->output derivation in chain_beads.go, no BroadcastAnim/BeadAnimIn construction outside the bridge, one BroadcastAnim call site, and beadAnimationState.tick is the sole lit/litVal writer."
