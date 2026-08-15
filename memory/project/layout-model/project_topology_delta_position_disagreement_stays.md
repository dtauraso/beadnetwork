---
name: project-topology-delta-position-disagreement-stays
description: topology/'s stored edge deltas disagree with stored node positions for most edges; this is KNOWN and deliberately left alone — do not "reconcile" it
metadata:
  type: project
---

In `topology/`, the saved edge delta index (`deltaIndex*`, formerly `deltaPolar*`) and the
saved node position index (`indexPhi`/`indexTheta`/`indexR`, formerly duplicated in
`scenePolar*`) are two representations of the same geometry, and for most edges they
DISAGREE — by tens of steps, not by a rounding unit (measured 2026-08-14, pre-index-only
migration; only `1To2` and `1To3` agreed). MODEL.md's polar model says `A + D = B`; the
saved tree does not satisfy it for those edges. The `scenePolar*` fields themselves are
gone (task/geometry-one-name-one-place): the index is now the sole authored position, so
this note's ORIGINAL comparison (continuous `scenePolar*` vs `indexPhi/Theta/R`) no longer
exists as a live pair of fields — the disagreement now lives entirely between an edge's
`deltaIndex*` and the two endpoint nodes' own `indexPhi/Theta/R`.

Nothing reconciles them at load, by design: `ResolveEdgeDeltas` keeps a stored delta whenever
the edge file has one, and `PlaceFromDeltas` derives a node point only when that node has no
stored point. Every node here has one, so both sides load as-is and the disagreement is inert.

**David's call, asked and answered: let it stay.** Do not regenerate the files, do not pick a
winner, do not add a load-time reconciliation or a guard that flags it. Committing either side
would bake in a choice about his layout.

Related: the OLD edgeMover's `updateDeltaFromEndpoints` recomputed the delta from mirrored
absolute endpoint positions and persisted it, so a run silently OVERWROTE the stored delta
with the positional difference. That is gone — the source node now persists the vector it
actually maintains (see [[project-edge-has-no-goroutine]]), so a run round-trips the stored
value instead. A residual ~5e-15 drift in `deltaPolarTheta` per run remains unexplained and
is NOT the same phenomenon as this disagreement.

Trap this cost once: "the edge JSONs change on every run" reads like float noise from a
`git diff` of one or two files. Check several edges before calling it noise —
[[feedback-debug-data-before-theory]].
