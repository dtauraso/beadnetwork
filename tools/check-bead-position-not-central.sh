#!/usr/bin/env bash
set -euo pipefail

# check-bead-position-not-central.sh — guard that a chain bead's POSITION is derived by
# the bead's OWN goroutine (nodes/wire/bead_actor.go), never recomputed centrally in
# nodes/Wiring/chain_beads.go. PLAN.md "two clocks per bead, three channel sets" / "THE
# REPLACEMENT IS NOT DONE UNTIL THE OLD PATH IS DELETED" names four hard requirements;
# this asserts the three that are source-shape (the fourth, this guard, asserts itself by
# existing and running):
#
#   1. offsetAt does not exist in chain_beads.go, and no per-index offset function is
#      passed into the bead layer.
#   2. No fallback branch ("use the actor's position if valid, else compute it here").
#   3. chain_beads.go does not itself turn (index, aim, distance) into a cartesian point
#      via Scale/polar2cart — that step is now inside nodes/wire/bead_actor.go
#      (Bead.applyTransform), reached only through ensureBeadEdgeActors/broadcastAndRead
#      (nodes/Wiring/bead_actor_bridge.go).
#
# A previous attempt (commit 100955e0, reverted) satisfied every behavioural test while
# keeping the central computation as a "fallback" the actor's output merely mirrored — no
# behavioural test can tell a replacement from a mirror, so only source shape can.
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
    echo "✗ bead-position-not-central: MISCONFIGURED — file not found: $f" >&2
    exit 1
  fi
done

# Requirement 1: offsetAt gone, chain_beads.go computes no per-index offset function.
if grep -nE '\boffsetAt\b' "$CHAIN_FILE"; then
  echo "✗ offsetAt reintroduced in chain_beads.go — a bead must derive its own offset/position from the broadcast transform, never be seeded with a caller-computed value (PLAN.md requirement 1)." >&2
  fail=1
fi

# Requirement 2/3: no "fallback to the old computation" shape — chain_beads.go must not
# turn a PER-BEAD-INDEX offset into a cartesian point itself. The textual fingerprint of
# the reverted mirror (and of the pre-actor original) is an R that scales with the bead
# index (float64(i)*wire.BeadStepR, the exact "d" formula the old inline math used) fed
# into .Scale(...) or polar2cart(...). polar2cart is still legitimately called in this
# file for the edge's fixed UNIT aim direction (R: 1, no per-bead index) — that is input
# derivation, not position; only an index-scaled R is banned.
if grep -nE 'float64\(i\)\s*\*\s*wire\.BeadStepR' "$CHAIN_FILE"; then
  echo "✗ chain_beads.go still computes a per-bead-index offset (float64(i)*wire.BeadStepR) — that arithmetic belongs ONLY inside nodes/wire/bead_actor.go's Bead.applyTransform (offsetR, fixed at construction in ensureBeadEdgeActors), never recomputed per index in the frame-packing loop (PLAN.md requirement 2/3)." >&2
  grep -nE 'float64\(i\)\s*\*\s*wire\.BeadStepR' "$CHAIN_FILE" >&2
  fail=1
fi
if grep -nE '\.Scale\(' "$CHAIN_FILE"; then
  echo "✗ chain_beads.go still scales a vector by a distance to produce a bead position (.Scale(...)) — that step belongs ONLY inside nodes/wire/bead_actor.go's Bead.applyTransform, reached via ensureBeadEdgeActors/broadcastAndRead (PLAN.md requirement 2/3)." >&2
  grep -nE '\.Scale\(' "$CHAIN_FILE" >&2
  fail=1
fi

# Requirement 3: chainBeads must call into the bead-actor bridge for position, not
# assemble ox/oy/oz from a locally-computed vec3 loop variable. Positive assertion: the
# broadcast+read call site must be present.
if ! grep -q 'broadcastAndRead(' "$CHAIN_FILE"; then
  echo "✗ chain_beads.go no longer reads bead positions through beadEdgeActors.broadcastAndRead — either it regressed to computing positions itself, or it was rewritten without wiring the bead-actor path back in (PLAN.md requirement 3)." >&2
  fail=1
fi

# Bead.applyTransform (nodes/wire/bead_actor.go) must remain the ONLY place a bead's
# position field is written — this is what requirement 1's "a bead derives its own
# position" means structurally: exactly one assignment to geomState.position, one method.
apply_writes=$(grep -c 'g\.position = ' "$ACTOR_FILE" || true)
if [ "$apply_writes" != "1" ]; then
  echo "✗ nodes/wire/bead_actor.go: expected exactly one write site for a bead's position (beadGeometryState.applyTransform); found $apply_writes. Position must have exactly one writer (PLAN.md 'disjoint writers')." >&2
  fail=1
fi

if [ "$fail" != "0" ]; then
  exit 1
fi

echo "✓ bead position stays actor-owned: no offsetAt, no central Scale/polar2cart in chain_beads.go, chainBeads reads through broadcastAndRead, and Bead.applyTransform is the sole position writer."
