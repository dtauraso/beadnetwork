package nodegeom

import (
	"math"

	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// EdgeStepCount is the bead-lattice length of an edge (docs/bead-model/bead-lattice.md "The count"): ONE
// INTEGER, the number of bead steps between the two nodes' tori. Computed from the LIVE
// measured center-to-center distance (dist), never a stored cache, plus both nodes' kinds:
//
//	K = round(dist / lattice.BeadStepR)
//	N = K - NodeTorusSteps(srcKind) - NodeTorusSteps(dstKind), minimum 1
//
// Under bead CRUD (MODEL.md "Moving a node is CRUD on the edge beads that touch it",
// bead_crud.go) a node's placement does NOT guarantee dist lands on an exact integer
// multiple of BeadStepR for every neighbour simultaneously (that guarantee belonged to the
// rejected global bead-cell solver) — round() here is the real discretizer, not a no-op:
// it is what turns a live, generally off-lattice distance into the whole bead-step count
// this edge actually renders.
//
// PURE INTEGER SUBTRACTION once K is known — no division anywhere else, not even by a
// fixed cell count. That used to divide the STORED QuantIR cache by a per-bead cell-count
// constant (4) before subtracting, assuming the node lattice's cell was a quarter of a
// bead step; before that, it read QuantIR straight off a cache that could go stale
// relative to a node's live position (an offset propagated to a neighbor before that
// neighbor's own commit landed, or a load-time value never re-measured after a drag).
// Reading the live distance instead means this can never disagree with where the node
// ACTUALLY is.
//
// NodeTorusOuterR is still snapped to a whole number of bead steps (NodeTorusSteps,
// port_geometry.go) rather than measured from width/height, so the subtraction's second and
// third terms are exact integers too — no term here can reintroduce a fraction.
//
// This is a pure function of state a node already owns, called from TWO places that must
// never disagree: chainBeads (chain_beads.go, this node's own goroutine, every cycle, for
// LAYOUT) and build.go's allocateWires (load time, for the wire's INITIAL published step
// count) — both call this same function on the same live center-to-center distance, so
// layout and timing can never read two different lengths (the exact divergence
// docs/bead-model/bead-lattice.md replaces the old arc-length model to close off).
func EdgeStepCount(dist float64, srcKind, dstKind string) int {
	k := int(math.Round(dist / lattice.BeadStepR))
	n := k - NodeTorusSteps(srcKind) - NodeTorusSteps(dstKind)
	if n < 1 {
		return 1
	}
	return n
}
