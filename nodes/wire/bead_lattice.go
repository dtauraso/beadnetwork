// bead_lattice.go — the bead lattice constants (docs/bead-lattice.md).
//
// An edge's length is ONE INTEGER, the bead-step count between two nodes' tori
// (docs/bead-lattice.md "The model"). Everything else here — the node lattice's
// own radial cell, and the uniform per-bead dwell that makes pulse speed structural
// instead of computed — falls out of the bead's AUTHORED size and these constants,
// not from independently chosen literals that could drift apart from it.
package wire

import "math"

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

// BeadStepCells is the number of node-lattice radial cells (LocalStepR each,
// layout_holder.go) a single bead step spans. Fixing this at 4 — rather than
// letting the node lattice's cell be authored independently of the bead step —
// is what keeps the bead lattice a commensurate SUBLATTICE of the node lattice:
// node separations keep their authored quantIR meaning, still counted in
// LocalStepR-sized cells, and exact double tangency only requires a separation
// to land on a multiple of 4 cells rather than on every cell. That property is
// UNCHANGED by the primitive/derived flip above — BeadStepCells never moved;
// only which of BeadStepR/LocalStepR is now the primitive and which is derived
// swapped (LocalStepR = BeadStepR / BeadStepCells, layout_holder.go).
// docs/bead-lattice.md "The lattice is commensurate with the node lattice"
// records this history in full.
const BeadStepCells = 4

// DwellTicksPerBead is the constant tick dwell every bead spends per lattice
// step, at the one uniform pulse speed (PulseSpeedWuPerTick, above in this
// package). Uniform pulse speed is structural under the bead lattice because
// dwell-per-bead is now a CONSTANT rather than a division that could vary per
// edge: ticksToCross for an N-step edge is simply N * DwellTicksPerBead, with
// no per-edge arc to divide by speed (docs/bead-lattice.md "Timing"). This is
// the value the ONE production PacedWire construction site (loader.go, guarded
// by tools/check-uniform-pulse-speed.sh) is expected to pass as dwellTicks.
const DwellTicksPerBead = BeadStepR / PulseSpeedWuPerTick

// SnapQuantIR rounds a node-separation quant index to the nearest multiple of
// BeadStepCells (4), the bead lattice's coarse-sublattice pitch (docs/bead-lattice.md
// "The lattice is commensurate with the node lattice"). QuantIR must be stored ONLY
// through this function — LayoutHolder.SetLocalPolar and LoadLocalPolars, the two
// write choke points for a LocalPolar entry, both call it — so an off-lattice
// separation can never be PERSISTED. That is what makes edgeStepCount's count a pure
// integer subtraction (nodeTorusSteps difference, no division) rather than exact only
// by luck of what happened to be stored: snapping at the read side instead would let
// an unsnapped value keep re-entering through every other writer that skipped this
// call. Nearest, not floor/ceil: this shifts an authored separation by at most half a
// bead step (2*LocalStepR world units — 4.0 when LocalStepR was 2.0, 4.48 now that it
// is derived as BeadStepR/BeadStepCells = 2.24) on first load — the accepted one-time cost of exact double
// tangency at both ends (docs/bead-lattice.md "The count").
func SnapQuantIR(quantIR int) int {
	return int(math.Round(float64(quantIR)/float64(BeadStepCells))) * BeadStepCells
}
