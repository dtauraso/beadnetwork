---
branch: task/cascade-double-links
---

# Restore the 5-8 and 7-9 cascade double links

## Goal

`topology/nodes/*/cascade-edges.json` currently stores a spanning TREE. The domain
graph (from each node's `localPolars`) has two more edges — `5-8` and `7-9` — and
those two are exactly the cycle-closers. Restore them so cascade adjacency equals
domain adjacency, then point the re-quantize fan at `cascadeEdges` so there is one
stored notion of "who is my neighbor" instead of two that can drift.

## Measured before starting

With both links restored and NO terminus rules in place, a single drag of node 5
produced a self-sustaining delta-forward flood: 130 messages in 400ms, 920 in
3000ms, a steady ~300/sec continuing long after the drag ended. `forwardDelta` has
no visit-tracking — it relied on the tree for termination.

Two cycles exist:

- Cycle A: `5-8-3-1-2-5`, closed by `5-8`
- Cycle B: `7-9-5-2-4-7`, closed by `7-9`

## Status: DONE

Both cycles are now cut by terminus rules already on main:

- Cycle A dies at node 3 — PulseLeft attends only to Input/SelectRight and never
  relays (merged `c4d9a9f1`).
- Cycle B dies at node 7 — PulseRight attends only to Time/SelectLeft and never
  relays (merged `c257f186`).

**Re-measure the flood before trusting this.** The ~300/sec number was taken before
node 7 had its rule.

## Result

Both links are in. Cascade adjacency now equals domain adjacency.

- `5-8` restored with the Pulse cascade rules (merged `9308f13a`) — it was required
  there, because without the entry a Pulse could not classify node 8 as SelectRight and
  the gate-routing rule silently fell through to a flood.
- `7-9` restored here.

Re-measured after both, with the terminus rules in place — every single-node drag
settles, none run away:

| drag | forwards | drag | forwards | drag | forwards |
|---|---|---|---|---|---|
| 1 | 4 | 4 | 1 | 7 | 6 |
| 2 | 3 | 5 | 5 | 8 | 2 |
| 3 | 4 | 6 | 3 | 9 | 2 |

Compare the uncut graph: a steady ~300 forwards/sec continuing indefinitely.

Termination is now a property of the PER-KIND RULES, not the edge set — PulseLeft (3)
and PulseRight (7) are termini; TimeStart (2) and Pulse (5) route by sender kind. That
is recorded in `delta_forward_test.go`'s header, which previously claimed loop-freedom
"by construction" from the tree shape.

## Not done: the requantize fan

Step 3 below is still open — `requantizeLocalPolars` still fans `neighborSetC` over the
domain-derived `updatesX` rather than `nm.cascadeEdges`. The two sets are now identical,
so there is no behavioral difference; it remains worth doing so there is a single stored
source of truth for "who is my neighbor". That mismatch is exactly what produced the
node-8-drag bug: 5 was told about a drag across a link its own cascade graph did not
have.

## Steps

1. Add `5-8` and `7-9` back to `cascade-edges.json` for nodes 5, 8, 7, 9 (both
   directions, with `cascadeKinds`).
2. Re-run the flood probe: drag node 5, count `moveMsgKindDeltaForward` via
   `SetMsgTap`, confirm it terminates instead of running at a steady rate.
3. In `requantizeLocalPolars` (`nodes/Wiring/quantized_move.go`), fan the
   `neighborSetC` send over `nm.cascadeEdges` instead of the domain-derived
   `updatesX`. Leave `requantizePoleTraced(lhX, updatesX)` alone — X still
   re-quantizes its own triples.

Step 3 is a refactor, not a behavior change: once cascade == domain the two sets are
identical. The win is a single source of truth.

## Side effect

`emitGeometry` draws one overlay tube per cascade pair, so `5-8` and `7-9` will start
rendering as cascade links.
