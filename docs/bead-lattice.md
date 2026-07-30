# The bead lattice — an edge is one integer

Supersedes the arc-length model for edges. Companion to
[beads-are-the-edge.md](beads-are-the-edge.md), which describes the chain; this file
describes the LENGTH.

## The model

A bead is a polar entity of the same kind as a node. It has its own local polar setup —
the same mechanism nodes use — and an invisible sphere whose radius is its TORUS OUTER
radius, not its bare sphere radius. Placement is tangency: the torus of a bead touches the
torus of whatever comes before it, and the target node is just another item on the last
bead's surface, its torus tangent in exactly the same way.

There is no arc. An edge's length is ONE INTEGER: the number of bead steps between the two
nodes' tori. Everything derives from it.

## The lattice is commensurate with the node lattice

A bead step is FOUR node-lattice radial cells:

	BeadStepCells = 4
	BeadStepR     = BeadStepCells * LocalStepR = 4 * 2.0 = 8.0

So the bead lattice is a coarse SUBLATTICE of the node lattice. This is what keeps the
cost of double tangency small: node separations keep their authored `quantIR` meaning
(still counted in 2.0-unit cells), so exact tangency needs the separation snapped only to
every 4th cell — a shift of at most 4 world units, not a re-interpretation. An earlier draft
made `LocalStepR` itself 8.96 to match a torus-touching bead of radius 4.0; that would
have re-interpreted every stored `quantIR` against a coarser step and shifted every
authored position on load. Rejected for that reason.

## The bead radius is derived, not chosen

Tangency fixes the bead's extent from the step:

	BeadTorusOuterR        = BeadStepR / 2 = 4.0
	ShadingParamBeadRadius = BeadTorusOuterR / (1 + ShadingParamBeadRingTubeRatio)
	                       = 4.0 / 1.12 = 3.5714...

The visible bead therefore SHRINKS ~11% from the old hand-picked 4.0. That is the point:
a bead's size is a consequence of the lattice it sits on, not an independent knob that can
drift away from the spacing.

## Placement

Along the source node's own stored bearing to the target, bead `i` sits at

	srcTorusOuterR + BeadTorusOuterR + i * BeadStepR

where `srcTorusOuterR = nodeTorusOuterR(kind)` — see "The count" below for why this is a
SNAPPED value, not the raw `nodeRadius(kind) * (1 + ShadingParamNodeRingTubeRatio)`. Bead
0's torus is tangent to the source node's torus; bead `i`'s torus is tangent to bead
`i-1`'s; bead `N-1`'s torus is tangent to the target node's torus, EXACTLY (see below, not
approximately as an earlier draft of this doc had it). "Beads are never inside a node"
falls out of the tangency, with no clamp.

## The count

Computed by the SOURCE NODE, from state it already owns — its stored `LocalPolar` to the
target and the target's kind from `cascadeKinds`:

	N = QuantIR/BeadStepCells - nodeTorusSteps(srcKind) - nodeTorusSteps(dstKind)   minimum 1

PURE INTEGER SUBTRACTION — no division by a float step, no `round()`. An earlier version of
this formula read

	separation = QuantIR * stepR
	gap        = separation - nodeTorusOuterR(srcKind) - nodeTorusOuterR(dstKind)
	N          = round(gap / BeadStepR)                 minimum 1

which promised tangency at BOTH ends but only delivered it at the source end. The defect
was not that `QuantIR` sat on the fine node lattice — snapping `QuantIR` alone would not
have made the division exact. The defect was `nodeTorusOuterR`: `nodeRadius(kind) * (1 +
ShadingParamNodeRingTubeRatio)` is an arbitrary float derived from a kind's width/height,
not a point on the bead lattice, so `gap` was essentially never an exact multiple of
`BeadStepR` and `round()` silently absorbed up to half a bead step at the TARGET end — the
double-tangency promise was a rounding coincidence, not a guarantee.

The fix makes every term an integer count of bead steps, so there is nothing left to round:

- `QuantIR` is snapped to a multiple of `BeadStepCells` (4) at every WRITE — the two write
  choke points, `wire.LayoutHolder.SetLocalPolar` (drag/requantize) and `LoadLocalPolars`
  (initial load) — never at the read side, because a value that CAN be stored unsnapped is
  the bug re-entering by another door. This shifts an authored separation by at most half a
  bead step (4 world units) on first load — the accepted one-time cost of exact tangency at
  both ends.
- A node's torus extent is snapped to a whole number of bead steps too:
  `nodeTorusSteps(kind) = ceil(bareRadius(kind) * (1+ShadingParamNodeRingTubeRatio) /
  BeadStepR)`, and `nodeTorusOuterR(kind) = nodeTorusSteps(kind) * BeadStepR`. CEIL, not
  round, so the snapped extent never cuts inside the node's own unsnapped body. `nodeRadius`
  (the node's SPHERE radius — the streamed/drawn radius, and the basis for ring-anchor
  placement) is in turn DERIVED from this snapped extent
  (`nodeTorusOuterR(kind)/(1+ShadingParamNodeRingTubeRatio)`), so the drawn ring the TS
  renderer scales by `nodeRadius` reaches exactly `nodeTorusOuterR(kind)` — the same integer
  a bead's tangent point is computed against. Nodes change size by up to one bead step
  versus the pre-snap width/height formula; that is the accepted cost of making both ends
  exact.

With both terms on the lattice, `QuantIR/BeadStepCells` is always an exact integer bead-step
count and `N` is plain subtraction. The off-by-a-fraction bug class is UNREPRESENTABLE, not
tuned away (`memory/feedback_make_bug_class_unrepresentable.md`) — pinned by
`TestChainBeadsExactDoubleTangency` (`nodes/Wiring/chain_beads_test.go`).

`edgeArcPolar` — `polarDist`'s sqrt over two scene polars, minus port radii, rounded to
`edgeLengthCellWu` — is DELETED. It was a second, independently-measured length that could
disagree with the lattice the beads were laid out on; that disagreement is the bug class
this model removes.

## Timing

Uniform pulse speed becomes structural rather than computed. Dwell per bead is a constant,
so a longer edge is simply more beads:

	ticksToCross = N * DwellTicksPerBead

which is numerically what `arcLength / pulseSpeed` used to give, with no per-edge division
and no length to divide. Lighting reads the same integer:

	litBeadIndex(t, N) = floor(t * N)

No length is multiplied anywhere, so layout and lighting cannot read two different values —
they read the same `N`.

## Ownership

- The SOURCE NODE computes `N` and publishes it onto its own `*wire.Out`. It owns the
  `LocalPolar` the beads are laid out from, it owns the chain, and it owns the edge file
  (`topology/nodes/<source>/edges/<label>.json`).
- The `edgeMover` publishes the SEGMENT (start/end) only. It no longer computes any length.

## Consequences to keep in mind

- `PacedWire` stores `steps int`, not an arc float. Its per-instance dwell stays a TEST
  affordance exactly as the per-instance `pulseSpeed` was: production passes the one
  constant (guard: `tools/check-uniform-pulse-speed.sh`), lean tests pass
  `NewPacedWire(latMs, 1.0)` so `ticksToCross == latMs` and their tick expectations are
  unchanged.
- The buffer's Event block carries `BeadSteps` where it carried `ArcLength`. Fingerprint
  bump in `Buffer/buffer_layout_gen.go` and its TS mirror.
- `ShadingParamNodeRingTubeRatio` is a new Go mirror of the TS-side `NODE_RING_TUBE_RATIO`
  (0.08), needed because the node's torus outer radius is now load-bearing geometry rather
  than decoration.
