// bead_lattice.go — the bead lattice constants (docs/bead-lattice.md).
//
// An edge's length is ONE INTEGER, the bead-step count between two nodes' tori
// (docs/bead-lattice.md "The model"). Everything else here — the bead's own
// radius, and the uniform per-bead dwell that makes pulse speed structural
// instead of computed — falls out of that one integer and these constants, not
// from independently chosen literals that could drift apart from it.
package wire

import "math"

// BeadStepCells is the number of node-lattice radial cells (LocalStepR each)
// a single bead step spans. Fixing it at 4 (rather than re-deriving LocalStepR
// itself to match a chosen bead radius) is what keeps the bead lattice a
// commensurate SUBLATTICE of the node lattice: node separations keep their
// authored quantIR meaning, still counted in LocalStepR-sized cells, and exact
// double tangency only requires a separation to land on a multiple of 4 cells
// rather than on every cell. docs/bead-lattice.md "The lattice is commensurate
// with the node lattice" records the rejected alternative (an 8.96-unit
// LocalStepR) and why it was rejected: it would have re-interpreted every
// stored quantIR against a coarser step and shifted every authored position on
// load.
const BeadStepCells = 4

// BeadStepR is the centre-to-centre distance between two tangent beads, in the
// same world units as LocalStepR: BeadStepCells node-lattice cells per bead
// step (docs/bead-lattice.md "The lattice is commensurate with the node
// lattice"). = 4 * 2.0 = 8.0.
const BeadStepR = BeadStepCells * LocalStepR

// BeadTorusOuterR is a bead's true extent — its invisible sphere radius —
// derived from tangency, not chosen independently: two tangent beads' tori
// touch at the lattice step's midpoint, so each bead's own torus radius is
// exactly half the step (docs/bead-lattice.md "The bead radius is derived,
// not chosen"). = 8.0 / 2 = 4.0. The VISIBLE bead radius (a further division
// by 1 + the ring/tube ratio) is a TS/shading-side derivation from this value,
// not duplicated here.
const BeadTorusOuterR = BeadStepR / 2

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
// bead step (4 world units) on first load — the accepted one-time cost of exact double
// tangency at both ends (docs/bead-lattice.md "The count").
func SnapQuantIR(quantIR int) int {
	return int(math.Round(float64(quantIR)/float64(BeadStepCells))) * BeadStepCells
}
