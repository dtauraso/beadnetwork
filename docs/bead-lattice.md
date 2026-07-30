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

## The lattice is derived, not the bead

This flipped once, and the flip is worth recording in full because both directions look
reasonable in isolation.

**First draft (rejected then, tried, then also rejected):** derive `LocalStepR` itself
(8.96) to match a torus-touching bead of radius 4.0, so the node lattice's own cell is
sized by the bead. Rejected before it was built: that would have re-interpreted every
stored `quantIR` against a coarser step and shifted every authored position on load.

**Second draft (built, shipped, then reverted — this is the one that actually ran):** make
`LocalStepR = 2.0` (a small, hand-picked node-lattice cell) the primitive, fix
`BeadStepCells = 4` as a constant, and let the bead step — and therefore the bead's own
visible size — fall out of that:

	BeadStepCells   = 4
	BeadStepR       = BeadStepCells * LocalStepR = 4 * 2.0 = 8.0
	BeadTorusOuterR = BeadStepR / 2 = 4.0
	ShadingParamBeadRadius = BeadTorusOuterR / (1 + ShadingParamBeadRingTubeRatio)
	                       = 4.0 / 1.12 = 3.5714...

This kept `quantIR`'s meaning intact (still counted in 2.0-unit cells; exact tangency only
needed a separation snapped to every 4th cell, a shift of at most 4 world units on first
load, not a re-interpretation) — solving the problem the first draft was rejected for. But
it made the bead's SIZE a dependent variable: the visible bead SHRANK to 3.5714..., about
11% smaller than the 4.0 David had actually picked. That was visibly wrong once it shipped
and was looked at, not a theoretical concern — a bead is the one thing a person looks
directly at, and "the lattice happens to imply this exact rendered size" is not a good
enough reason for that size to be someone else's decision.

**Current (this file, as shipped):** the bead's own radius is the AUTHORED primitive, and
the node lattice's cell is what derives from it — `BeadStepCells` stays fixed at 4 (the
property the second draft got right — the bead lattice is still a coarse SUBLATTICE of the
node lattice, so `quantIR`'s cell-counted meaning is preserved, only the cell's SIZE
changes), but now the derivation runs the other way:

	BeadRadius              = 4.0                                  (authored — wire.BeadRadius)
	BeadRingTubeRatio       = 0.12                                  (authored — wire.BeadRingTubeRatio)
	BeadTorusOuterR         = BeadRadius * (1 + BeadRingTubeRatio) = 4.48
	BeadStepR               = 2 * BeadTorusOuterR                  = 8.96
	BeadStepCells           = 4                                    (unchanged)
	LocalStepR              = BeadStepR / BeadStepCells            = 2.24

Tangency is UNCHANGED by this flip — two beads' tori still touch at exactly half
`BeadStepR`, and a bead's torus is still tangent to a node's torus the same way; only which
of {bead radius, lattice step} is free and which is computed has swapped. What DOES change:
`LocalStepR` grows from 2.0 to 2.24, a node-lattice cell now measures 12% more, and stored
`quantIR` values are NOT rewritten to compensate — every existing separation, still counted
in the same integer cells, now spans more world units, so the whole graph expands ~12% on
load. That expansion is the accepted, agreed cost of getting the bead's own size right; it
is not compensated for anywhere in this file or its callers.

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
  bead step (2 node-lattice cells, `2*LocalStepR` world units — 4.0 when `LocalStepR` was
  2.0, 4.48 now that it is 2.24) on first load — the accepted one-time cost of exact
  tangency at both ends.
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
