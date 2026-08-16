---
branch: task/self-drag-rule
---

# A separate rule for dragging a single node

## Why

David, 2026-08-15: *"the rules for node 1 are for dragging node 1"*, then
*"add a separate thing for dragging a single node."*

Today's drag rule does not constrain the node it is attached to. It constrains
that node's OFFSET FROM EACH IN-NEIGHBOUR (`nodedrag.TrimToDragRule` loops
`NeighborKinds()`, skips `IsOutTarget`, and trims `DeltaFrom(neighbour)`). So:

- Node 1 has no in-edges at all (nothing in the tree targets it), so its rule is
  inert — setting phi on it changes nothing about dragging it.
- The node frame still streams `dragRLocked`/`dragPhiLocked` off `rule != nil`
  alone (`node_geometry_stream.go:90`), so node 1 would READ as locked while
  dragging freely. That mismatch is the visible bug.

The existing rule is not wrong, it is a different thing. So this adds a second
one rather than changing the first.

## The model

A node now carries TWO independent rules:

- **the neighbour rule** (existing, unchanged) — constrains this node's offset
  from each in-neighbour. Per-edge, gated by `edgeActive`.
- **the self rule** (new) — constrains this node's OWN composed index, i.e. its
  polar position about the scene centre. No neighbours involved, so no per-edge
  dimension and `edgeActive` does not gate it.

Same `polar.DragRule` shape and the same `TrimDelta` for both, so there is one
vocabulary and not two: `Phi != nil` pins phi, `MaxTheta != nil` clamps theta,
and r is pinned whenever the rule exists at all (the behaviour
`memory/project/layout-model/project_rule_off_frees_radius_too.md` records for
the neighbour rule — an absent rule frees r too).

Under the self rule: phi pinned = the node stays on its cone about the scene
centre; r pinned = it stays on its sphere; maxTheta clamps its own swing.

**Order:** the neighbour trim runs first, then the self trim, so the node's own
constraint has the final say on where the node may end up. Both only ever
remove freedom, so composing them cannot invent movement.

## Pieces

1. `nodes/Wiring/nodedrag/node_drag.go` — `TrimToSelfRule(delta, of)` trimming
   `of.ComposedIndex()`; `SelfRule()`/`SelfRuleActive()` on the `Node`
   interface; called from `Apply` after the existing trim.
2. `nodes/Wiring/rulenode/` — `selfRule`/`selfActive`, three new `EditKind`s,
   both on `State`.
3. Persistence — `selfDrag` in `base.json`, `selfActive` in the gitignored
   `rule-active.json`; loaded by `loadspec`.
4. Buffer — `SelfPhiLocked`, `SelfThetaMax`, `SelfActive` columns on the node
   frame, so the panel can show the self rule separately from the neighbour one.
5. Bridge — `selfDragPhi`, `selfDragMaxTheta`, `selfDragActive` as node
   ATTRIBUTES (a new attribute, not a new op), plus fingerprint and parity.
6. `NodeRulesPanel`/`NodeRuleBlocks` — a self-rule section per node, separate
   from the per-edge sections.

## Verification

- `bash scripts/stop-checks.sh` clean (empty stdout), incl. the bridge parity
  guards and `check-input-attr-dispatched` (which fails loudly on an attribute
  that decodes but never dispatches — the exact failure mode for a half-wired
  attribute).
- Drive it: set the self phi rule on node 1 and drag. Before: no change at all.
  After: node 1 holds its cone and swings only in theta.
- Node 1 is the only pure source in this tree, so it is the case that proves the
  self rule is independent of in-edges. Nodes 2 and 3 carry neighbour rules and
  must keep behaving exactly as they do now when no self rule is set.

## Risks

- Two rules on one node is two things to read in the panel; if the panel does
  not make "about its neighbour" vs "about itself" obvious, this trades one
  confusion for another.
- The lock indicator bug above is NOT fixed by this on its own: the neighbour
  rule's `dragRLocked`/`dragPhiLocked` still ignore whether the node has any
  in-edge to bite on. Worth fixing in the same pass so node 1 stops lying.
