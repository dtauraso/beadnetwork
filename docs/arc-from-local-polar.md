# One integer per edge — the arc comes from the stored LocalPolar

## The problem

Three independent derivations of the same geometry, and nothing makes them agree.

| what | derived from | owner |
|---|---|---|
| a node's world position | its `ScenePolar` (r, θ, φ about the scene centre) | that node's mover |
| a chain's bead positions and count | the source node's `LocalPolar` to that neighbour: `QuantIR × StepR` | that node's mover |
| a bead's traversal time (`ticksToCross = arc/pulseSpeed`) | `edgeArcPolar` = `polarDist(both ScenePolars)` − both port offsets, requantised to `edgeLengthCellWu` (0.1) | the EDGE's mover |

Measured on node 1's two edges, as committed:

```
edge   QuantIR × StepR (beads)   polarDist(ScenePolars) (timing)   gap
1->2                   256.000                           256.973   -0.973
1->3                   258.000                           257.491   +0.509
```

`QuantIR` is quantised to `StepR` = 2 world units, so it can sit up to 1 unit either side of the
true scene-polar separation. The chain is laid out from `QuantIR`; the bead is timed against
`polarDist`. **The two disagree in opposite directions on the two edges**, so the two chains
cannot stay in step by construction — the bead is timed to cross a distance the chain is not
built to.

This is the root of the offsets chased on 2026-07-29. Each earlier fix was a real bug
(drain-time `placementTick`; fraction-vs-distance in the lit index; centre-distance vs arc; the
unreachable tail) but none of them was this, because none of them made the two lengths one length.

## Why it ended up this way

Neither derivation is wrong for its own job. `LocalPolar` exists for LAYOUT — it is how a node
repositions a neighbour at a stored bearing and distance when something moves
(`neighborSetCRequantize`). `edgeArcPolar` exists for TRAVERSAL — it feeds `ticksToCross`. They
were never the same number and nothing required them to be.

Putting the beads on the layout lattice (`task/beads-on-local-polar-lattice`) coupled them for the
first time, which is what exposed the gap. The gap was always there; nothing had depended on it.

## The decision

**`QuantIR` becomes authoritative for edge length.** The edge's arc is derived from the stored
`LocalPolar` the beads are laid out from:

    arc = QuantIR × StepR − srcPortOffset − dstPortOffset

instead of from `polarDist` of the two scene polars.

This INVERTS today's relationship. Currently the scene polars are the truth and `QuantIR` is a
quantised cache of them; afterwards an edge is exactly `QuantIR` steps long because that is what
is stored, and the rendered geometry follows. That is consistent with the polar model, which
already treats the quantised indices as the source of truth for movement
(`memory/feedback_abc_times_constant_not_rederive.md`), but it is a real model change and is the
part to disagree with if any part of this is wrong.

What it buys:

- Bead count, bead positions, and traversal timing all come from ONE integer per edge. Drift
  becomes unrepresentable rather than merely small.
- `arc` lands on the bead grid (`chainBeadSpacing / StepR` = 8/2 = 4 radial steps per bead), so
  `arc / chainBeadSpacing` is whole. The final bead stops dwelling longer than the others and the
  remainder clamp in `litBeadIndex` becomes dead code.
- One radial step of a drag = exactly one bead added or removed from that chain, visibly.

## Ownership: the arc must be published, not read

`edgeArcPolar` is called from `edgeMover.recomputeGeometry`, on the EDGE's goroutine, from
`srcGeom`/`dstGeom` that the endpoints sent it by message. The `LocalPolar` lives in the SOURCE
node's `LayoutHolder`. So the edge cannot read it — that would be a cross-goroutine read of
another goroutine's state, the same defect the `placementTick` fix removed a few commits ago
(decided by one goroutine, written by another).

The source node must SEND it. `moveMsgKindCenter` already carries a moved node's centre to its
incident edges, so the quantised distance rides that existing message: no new channel, no new
message kind, no new mechanism. The edge stores the last value it was told, exactly as it already
stores `srcGeom`.

**Rule to hold:** the number that decides an edge's length is written by the node that owns the
`LocalPolar` it comes from. The edge is a consumer.

## Steps

1. **Carry the quantised distance on `moveMsgKindCenter`.** Source node → its outgoing edges. The
   edge stores it beside `srcGeom`.
2. **`edgeArcPolar` takes the stored distance** instead of calling `polarDist`, keeping the port
   offset subtraction and the `CurveParamMinArcLength` floor. Where no value has arrived yet
   (startup, before the first move), fall back to today's `polarDist` path so the edge is never
   without an arc — and say so in the comment, because a silent fallback that never fires is
   indistinguishable from one that always does.
3. **Set `edgeLengthCellWu` to the bead spacing** so the requantise step lands on the bead grid
   rather than on 0.1. Re-measure and confirm `arc mod chainBeadSpacing == 0` for every edge.
4. **Delete the remainder clamp** in `litBeadIndex` once step 3 makes it unreachable — and delete
   the test that pins it, since it will assert something that can no longer happen. Do not leave a
   clamp "just in case": it hides exactly the drift this document removes.
5. **Re-measure the two chains** and confirm the numbers agree where the table above shows a gap.

## Risks worth naming

- **`ScenePolar` and `QuantIR` will now disagree by up to `StepR`/2, permanently and visibly.**
  Today that disagreement is absorbed by the arc; afterwards the drawn edge is the quantised
  length while the node centres sit wherever their scene polars put them. The wire may therefore
  not exactly touch both ports. Whether that is acceptable, or whether node placement must snap to
  the quantised distance too, is the open question this spec does not settle.
- **A feedback loop** if the arc ever writes back into node positions. It must not: `QuantIR` is
  written only by its owning node's requantise path, and the arc is a read of it.
- **`CurveParamMinArcLength` (1.0) is not on the bead grid.** An edge short enough to hit the floor
  gets an arc that is not a whole number of beads, so the remainder returns for that edge alone.
  Either raise the floor to one bead spacing or accept and document it.
