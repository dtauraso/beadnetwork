# The bead lattice — an edge is one integer

This file describes an edge's LENGTH. The chain it measures — the node-owned chain of
placeholder beads that renders a traversal — is described in MODEL.md and
[docs/model/entities.md](../model/entities.md).

## The model

A bead is a polar entity of the same kind as a node. It has its own local polar setup —
the same mechanism nodes use — and an invisible sphere whose radius is its TORUS OUTER
radius, not its bare sphere radius. Placement is tangency: the torus of a bead touches the
torus of whatever comes before it, and the target node is just another item on the last
bead's surface, its torus tangent in exactly the same way.

There is no arc. An edge's length is ONE INTEGER: the number of bead steps between the two
nodes' tori. Everything derives from it.

The bead's radius is the AUTHORED primitive and the lattice derives from it:

	BeadRadius        = 4.0                                   (lattice.BeadRadius)
	BeadRingTubeRatio = 0.12                                  (lattice.BeadRingTubeRatio)
	BeadTorusOuterR   = BeadRadius * (1 + BeadRingTubeRatio) = 4.48
	BeadStepR         = 2 * BeadTorusOuterR                  = 8.96

The scene's radial grid IS that step: `constants.json`'s `constantR` is 8.96, and
`loadspec.LoadSceneConstants` fails the load by path if it disagrees with
`lattice.BeadStepR`, so a stored `indexR` counts bead steps directly.

## Placement

Along the source node's own stored bearing to the target, bead `i` sits at

	srcTorusOuterR + BeadTorusOuterR + i * BeadStepR

where `srcTorusOuterR = nodegeom.NodeTorusOuterR(kind)` — see "The count" below for why this
is a SNAPPED value, not the raw `nodegeom.NodeRadius(kind) * (1 +
ShadingParamNodeRingTubeRatio)`. Bead
0's torus is tangent to the source node's torus; bead `i`'s torus is tangent to bead
`i-1`'s; bead `N-1`'s torus is tangent to the target node's torus, EXACTLY (see below, not
approximately as an earlier draft of this doc had it). "Beads are never inside a node"
falls out of the tangency, with no clamp.

## The count

See [bead-count.md](bead-count.md) for how the source node computes `N`, why it is pure
integer subtraction, and what is snapped at write time to make that exact.

## Timing

Uniform pulse speed becomes structural rather than computed. Dwell per bead is a constant,
so a longer edge is simply more beads:

	ticksToCross = N * DwellTicksPerBead

with no per-edge division and no length to divide. Uniform pulse speed is structural rather
than computed: a longer edge is more beads, not a bigger number to divide (guard:
`tools/network/beads/check-uniform-pulse-speed.sh`). Lighting does not read a length either:
each bead carries its own
`Lit` flag on its actor (`nodes/wire/beadchain/bead_actor.go`) rather than being selected by
a fraction of the edge. No length is multiplied anywhere, so layout and lighting cannot read
two different values.

## Ownership

- The SOURCE NODE computes `N` and publishes it onto its own `*wire.Out`. It owns the
  bearing the beads are laid out from — its own `owners.Deltas`, read through `DeltaTo` —
  it owns the chain, and it owns the edge file
  (`topology/nodes/<source>/edges/<label>.json`).
- The `edgeMover` publishes the SEGMENT (start/end) only. It no longer computes any length.

## Consequences to keep in mind

- `PacedWire` stores `steps int`, not an arc float. `NewPacedWire(steps int, dwellTicks
  float64)` takes the dwell as an argument, and every call site passes the one constant
  `lattice.DwellTicksPerBead` (guard:
  `tools/network/beads/check-uniform-pulse-speed.sh`) — there is no per-instance speed.
- The buffer's Event block carries `BeadSteps` where it carried `ArcLength`. Fingerprint
  bump in `Buffer/buffer_layout_gen.go` and its TS mirror.
- `ShadingParamNodeRingTubeRatio` is a new Go mirror of the TS-side `NODE_RING_TUBE_RATIO`
  (0.08), needed because the node's torus outer radius is now load-bearing geometry rather
  than decoration.
