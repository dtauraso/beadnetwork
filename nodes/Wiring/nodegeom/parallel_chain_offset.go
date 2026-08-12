// parallel_chain_offset.go — the mutual-pair parallel-chain-offset concern, split out of
// port_geometry.go (which keeps the single edge segment / distance-and-direction concern).
// See MODEL.md's "Node positions & movement locks (the polar model)" section, on a mutual
// pair offsetting its two chains, for the model this implements.

package nodegeom

import (
	"math"
	"strconv"

	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// ParallelChainOffset is the perpendicular displacement a node applies to its OWN chain
// toward `targetID` so that two nodes pointing AT EACH OTHER do not draw their chains on
// top of one another.
//
// Two edges between the same pair of nodes (1 -> 2 and 2 -> 1) run along the SAME centre
// line, so without this every bead of one chain sits exactly on a bead of the other and the
// pair renders as a single wire — the two edges are there, but unrepresentable on screen.
//
// CANONICAL ORDER IS THE WHOLE TRICK. Each node computes this alone, on its own goroutine,
// with no message to the other — so the two must arrive at OPPOSITE answers from their own
// local view. Deriving the perpendicular from each node's own outgoing direction fails:
// node 2's direction is the negation of node 1's, so its perpendicular negates too, and the
// two offsets cancel back onto the same line. The direction is therefore always measured
// from the LOWER node id to the HIGHER, giving both ends the identical perpendicular, and
// the sign is then taken from which end this node is. Local decision, no coordination, and
// the pair cannot disagree because neither is asking the other.
//
// The magnitude is one bead STEP each way, so the two chains end up exactly 2*lattice.BeadStepR
// apart — still on the lattice, not a tuned pixel gap. It was half that (one bead radius
// each way, the chains exactly touching), which separated them in principle but read as one
// thick wire; a full step each way leaves a clear bead-sized gap between the two chains.
//
// The cost, stated rather than hidden: an offset chain no longer starts exactly tangent to
// its node's torus, since it is displaced off the centre line that tangency is measured on.
// That is the trade the separation buys, and it applies ONLY to a mutual pair.
func ParallelChainOffset(selfID, targetID string, selfCenter, targetCenter, sceneCenter vec3) (vec3, bool) {
	lowCenter, highCenter := selfCenter, targetCenter
	if !NodeIDLess(selfID, targetID) {
		lowCenter, highCenter = targetCenter, selfCenter
	}
	delta := highCenter.Sub(lowCenter)
	if delta.Length() < 1e-9 {
		return vec3{}, false
	}
	dir := delta.Normalize()
	// COPLANAR WITH THE TORI. A node's ring lies in the plane whose normal is its own
	// INWARD pole — the direction from the node to the scene centre (node_mover.go's
	// inwardPole, the same derivation the streamed poleTheta/polePhi come from). Offsetting
	// along an arbitrary world axis would lift the chain out of that plane, so the offset is
	// taken perpendicular to the POLE: cross(pole, dir) is perpendicular to the pole, hence
	// IN the ring's plane, and perpendicular to the edge, hence still a clean displacement.
	//
	// The pole is taken from the canonically-lower node, not from whichever node is asking,
	// for the same reason the direction is: both ends must compute one identical vector.
	// Each end can derive it alone — an inward pole is a function of a centre and the scene
	// centre, and a node holds its partner's centre already (partnerCenters).
	poleAxis := sceneCenter.Sub(lowCenter)
	if poleAxis.Length() < 1e-9 {
		// A node sitting ON the scene centre has no inward direction and therefore no
		// ring plane to be coplanar with.
		return vec3{}, false
	}
	pole := poleAxis.Normalize()
	perp := pole.Cross(dir)
	if perp.Length() < 1e-6 {
		// The edge runs along the pole (radially, straight at the scene centre): every
		// perpendicular lies in the ring plane, so any one will do — but it must still be
		// the SAME one at both ends, so it is derived from the pole rather than picked.
		alt := vec3{X: 0, Y: 1, Z: 0}
		if math.Abs(pole.Dot(alt)) > 0.9 {
			alt = vec3{X: 1, Y: 0, Z: 0}
		}
		perp = pole.Cross(alt)
	}
	if perp.Length() < 1e-9 {
		return vec3{}, false
	}
	sign := 1.0
	if !NodeIDLess(selfID, targetID) {
		sign = -1.0
	}
	return perp.Normalize().Scale(sign * lattice.BeadStepR), true
}

// NodeIDLess orders two node ids NUMERICALLY, because node ids are numbers that are strings
// only because they are directory names (.claude/rules/persistence-ownership.md). A plain
// string compare would order "10" before "2" and hand both ends of that pair the same sign,
// which is the one thing ParallelChainOffset must never do. A non-numeric id (impossible
// today — loadTree rejects one) falls back to the string compare rather than panicking in
// geometry code.
func NodeIDLess(a, b string) bool {
	ai, aerr := strconv.Atoi(a)
	bi, berr := strconv.Atoi(b)
	if aerr == nil && berr == nil {
		return ai < bi
	}
	return a < b
}
