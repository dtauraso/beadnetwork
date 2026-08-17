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

Do not "fix" this by snapping the stored index to a multiple of the bead step at write
time. That was tried: it makes the separation exact at the cost of shifting every authored
position on load, and it does not remove the rounding so much as move it earlier.

**Never measure an edge's length a second way.** One length, read from one place. A second,
independently-derived length — a sqrt over two scene polars, say — will disagree with the
lattice the beads are laid out on, and the beads will be timed to cross a distance the
chain is not built to. That is the bug class this model exists to remove, and it is removed
by having ONE length, not by making that length exact.
