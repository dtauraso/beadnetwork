# The bead lattice — the count

[← bead-lattice.md](bead-lattice.md)

Computed by the SOURCE NODE, from state it already owns — the centre-to-centre distance to
the target and the target's kind (`nodes/Wiring/nodeactor/owners/out_edges.go`, which calls
`edgegeom.EdgeCenterDistAndDir` and then):

	EdgeStepCount(dist, srcKind, dstKind) =
	    round(dist / BeadStepR) - NodeTorusSteps(srcKind) - NodeTorusSteps(dstKind)
	    minimum 1

Both subtracted terms ARE integers. A node's torus extent is snapped to a whole number of
bead steps — `NodeTorusSteps(kind) = ceil(BareNodeRadius(kind) * (1 +
ShadingParamNodeRingTubeRatio) / BeadStepR)`, CEIL so the snapped extent never cuts inside
the node's own unsnapped body, and `NodeTorusOuterR(kind) = NodeTorusSteps(kind) *
BeadStepR`. `NodeRadius` (the node's SPHERE radius — the streamed/drawn radius, and the
basis for ring-anchor placement) is DERIVED back from that snapped extent, so the drawn ring
the TS renderer scales by `NodeRadius` reaches exactly `NodeTorusOuterR(kind)` — the same
integer a bead's tangent point is computed against.

## Where the rounding still is

The separation itself is NOT an integer count. `dist` is a float world distance between two
node centres, and `round(dist / BeadStepR)` is the count it becomes. So the last rounding
step is still in the formula, and tangency at the target end is exact only to within half a
bead step.

An earlier version of this file said otherwise — that `N` was "PURE INTEGER SUBTRACTION —
no division by a float step, no `round()`", because the separation was stored as a snapped
`QuantIR` in bead-step multiples, snapped at both write choke points (`SetLocalPolar` and
`LoadLocalPolars` on a `wire.LayoutHolder`). None of that exists any more: there is no
`LayoutHolder`, no `SetLocalPolar`, no `BeadStepCells`, and no snapping of the stored index
to a multiple of the bead step. A node's position is stored as `indexPhi`/`indexTheta`/
`indexR` on the NODE lattice (`.claude/rules/persistence-ownership.md`), which is a finer
grid than the bead lattice and is not a whole number of bead steps.

What survives from that reasoning is the half worth keeping: the two NODE-extent terms are
snapped, so they contribute no fraction. The remaining rounding is in the separation only.

`edgeArcPolar` — a sqrt over two scene polars, minus port radii, rounded to a cell width —
is DELETED. It was a second, independently-measured length that could disagree with the
lattice the beads were laid out on; that disagreement is the bug class this model removes,
and it is removed by having ONE length, not by making that length exact.
