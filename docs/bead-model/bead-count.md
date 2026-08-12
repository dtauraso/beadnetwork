# The bead lattice — the count

[← bead-lattice.md](bead-lattice.md)

Computed by the SOURCE NODE, from state it already owns — its stored `LocalPolar` to the
target and the target's kind from `neighborKinds`:

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
tuned away (`memory/feedback/architecture/geometry/feedback_make_bug_class_unrepresentable.md`).

`edgeArcPolar` — `polarDist`'s sqrt over two scene polars, minus port radii, rounded to
`edgeLengthCellWu` — is DELETED. It was a second, independently-measured length that could
disagree with the lattice the beads were laid out on; that disagreement is the bug class
this model removes.
