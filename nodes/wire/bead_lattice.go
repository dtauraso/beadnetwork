// bead_lattice.go — the bead lattice constants (docs/bead-lattice.md).
//
// An edge's length is ONE INTEGER, the bead-step count between two nodes' tori
// (docs/bead-lattice.md "The model"). Everything else here — the node lattice's
// own radial cell, and the uniform per-bead dwell that makes pulse speed structural
// instead of computed — falls out of the bead's AUTHORED size and these constants,
// not from independently chosen literals that could drift apart from it.
package wire

// BeadRadius is the bead's own visible sphere radius — the AUTHORED primitive
// this whole file derives from. It used to be the other way around: the node
// lattice's LocalStepR was the primitive and the bead radius fell out of it by
// tangency, landing at 3.5714285714285716 — visibly ~11% smaller than the
// hand-picked 4.0 David had actually chosen. That direction is now REJECTED
// (docs/bead-lattice.md "The bead radius is derived, not chosen" — renamed to
// "the lattice is derived, not the bead"): the bead's size is what a person
// looks at, so it wins, and the node lattice's own cell absorbs the change
// instead (LocalStepR grows from 2.0 to 2.24 in layout_holder.go). Stored
// quantIR values are NOT rewritten for this — each cell simply now measures
// 12% more, so the whole graph expands by that fraction on load. That
// expansion is the accepted, agreed cost; nothing here compensates for it.
const BeadRadius = 4.0

// BeadRingTubeRatio is a bead ring's torus tube radius as a fraction of
// BeadRadius, same role as ShadingParamBeadRingTubeRatio used to play alone —
// moved into this package (from nodes/Wiring/shading_params.go) because it now
// feeds BeadTorusOuterR, which nodes/wire's own lattice constants need, and
// nodes/wire cannot import nodes/Wiring (Wiring imports wire; that direction
// would be a cycle). nodes/Wiring's ShadingParamBeadRingTubeRatio now just
// references this value.
const BeadRingTubeRatio = 0.12

// BeadTorusOuterR is a bead's true extent — its invisible sphere radius — now
// derived from the AUTHORED sphere radius by the same outer = r*(1+ratio)
// formula every other ring in this codebase uses: 4.0 * 1.12 = 4.48. Tangency
// is unchanged by the flip (docs/bead-lattice.md "The bead radius is derived,
// not chosen"/its rewrite): two tangent beads' tori still touch at the lattice
// step's midpoint, so BeadStepR below is still exactly twice this value — only
// which of {bead radius, lattice step} is the free variable and which is
// computed has flipped.
const BeadTorusOuterR = BeadRadius * (1 + BeadRingTubeRatio)

// BeadStepR is the centre-to-centre distance between two tangent beads, fixed
// by tangency at twice a bead's own torus outer radius: 2 * 4.48 = 8.96.
const BeadStepR = 2 * BeadTorusOuterR

// BeadStepCells does not exist. It used to be the ratio between TWO lattices —
// the node lattice's own small LocalStepR cell, and the coarser bead lattice a
// fixed number of those cells (4) was said to span. That ratio was a
// mismeasurement: BeadStepCells was fixed at 4 while the STORED per-node
// stepR (topology/nodes/*/local-polars.json, "stepR": 2) disagreed with what
// LocalStepR actually computed to (2.24) — placement read the stored 2, the
// edge-length count divided by the assumed 4, and the two lattices did not
// nest, so the count over-budgeted and surplus chain beads ran into the
// target node (docs/bead-lattice.md). The fix collapses the two lattices into
// ONE: LocalStepR is now simply BeadStepR (below) — a node moves exactly one
// bead distance per lattice tick — so there is no second, coarser lattice
// left to express a cell-count between, and the constant that named that
// relationship has nothing to name. There is no replacement constant here —
// deleting it is the point: the bug was two lattices that could disagree, and
// the fix is having only one, not renaming the ratio.

// DwellTicksPerBead is the constant tick dwell every bead spends per lattice
// step, at the one uniform pulse speed (PulseSpeedWuPerTick, above in this
// package). Uniform pulse speed is structural under the bead lattice because
// dwell-per-bead is now a CONSTANT rather than a division that could vary per
// edge: ticksToCross for an N-step edge is simply N * DwellTicksPerBead, with
// no per-edge arc to divide by speed (docs/bead-lattice.md "Timing"). This is
// the value the ONE production PacedWire construction site (loader.go, guarded
// by tools/check-uniform-pulse-speed.sh) is expected to pass as dwellTicks.
const DwellTicksPerBead = BeadStepR / PulseSpeedWuPerTick

// SnapQuantIR does not exist. It used to round a stored quantIR to the nearest
// multiple of BeadStepCells so the node lattice stayed a commensurate SUBLATTICE
// of the bead lattice. With BeadStepCells gone (comment above) there is exactly
// one lattice, so the only multiple left to snap to is 1 — an identity
// operation, which is not a snap. Its two write choke points
// (LayoutHolder.SetLocalPolar/LoadLocalPolars) now store quantIR verbatim.

// BeadVector is a bead-lattice displacement: a quantized bearing (Dir — same shape as
// Pole, colatitude+azimuth about some pole already in force) and a magnitude that is a
// WHOLE COUNT of bead steps (N), never a float distance. The two fields are kept
// deliberately SEPARATE rather than folded into one cartesian vector: a bead's legal
// motion is always "this many whole bead-steps along this stored bearing", and a single
// vec3 would let direction and magnitude be read back out independently via
// Length()/Normalize() — exactly the cartesian shortcut
// (nodes/Wiring/quantized_move.go's now-deleted walkBeadPath, and the reverted
// single-neighbour cell attempt before it) that let a drag silently become a
// normalize-and-scale cartesian stride instead of index arithmetic on the lattice. With
// Dir and N as separate fields there is no field to normalize or measure a length from —
// the cartesian shortcut is structurally unavailable, not merely avoided by convention.
type BeadVector struct {
	Dir Pole
	N   int
}

// Length is this vector's magnitude in world units: N whole BeadStepR hops — pure
// multiplication (index × constant, memory/feedback_abc_times_constant_not_rederive.md),
// never a measured-then-rounded cartesian distance. The one caller-needed method: nothing
// else in nodes/Wiring's drag-candidate math reads a BeadVector's components any other way.
func (bv BeadVector) Length() float64 {
	return float64(bv.N) * BeadStepR
}
