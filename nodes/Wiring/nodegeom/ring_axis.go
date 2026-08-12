// ring_axis.go — ring-axis derivation, split out of port_geometry.go (which keeps the
// single edge segment / distance-and-direction concern). These functions compute the
// streamed RingAxisTheta/RingAxisPhi a node's own drawn ring uses, distinct from the
// navigation pole (MODEL.md "The drawn ring axis and the navigation pole are two different
// streamed values, on purpose").

package nodegeom

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
)

// PoleContainingEdge returns the ring axis closest to the given pole whose PLANE contains
// the edge from selfCenter to partnerCenter — the pole with its along-the-edge component
// projected out.
//
// A ring's plane is everything perpendicular to its axis, so "the edge lies in the plane"
// is exactly "the axis is perpendicular to the edge". Removing the component along the edge
// is the minimal rotation achieving that: it keeps as much of the original inward pole as
// the constraint allows, rather than picking an unrelated perpendicular.
//
// Not resolvable when the two centres coincide (no edge direction) or when the pole is
// PARALLEL to the edge — there the projection vanishes and every perpendicular is equally
// good, so the caller keeps the pole it had rather than this function inventing one.
func PoleContainingEdge(poleTheta, polePhi float64, selfCenter, partnerCenter vec3) (theta, phi float64, ok bool) {
	delta := partnerCenter.Sub(selfCenter)
	if delta.Length() < 1e-9 {
		return 0, 0, false
	}
	dir := delta.Normalize()
	pole := geom.AnglesToWorldOffset(1, poleTheta, polePhi)
	projected := pole.Sub(dir.Scale(pole.Dot(dir)))
	if projected.Length() < 1e-6 {
		return 0, 0, false
	}
	u := projected.Normalize()
	return math.Acos(geom.Clamp(u.Y, -1, 1)), math.Atan2(u.Z, u.X), true
}

// TorusDefaultAxisAngles is the torus geometry's OWN normal (+Z) as this codebase's angle
// pair. A ring streamed with this axis is drawn exactly as an unrotated one, which is what
// every scene looked like before ring orientation existed — so it is the default, and a
// scene opts IN to anything else.
func TorusDefaultAxisAngles() (theta, phi float64) {
	return math.Pi / 2, math.Pi / 2
}

// UprightRingAxis returns the ring axis whose PLANE contains BOTH the edge and world +y —
// a ring standing upright along the edge, with the node's own up-vector lying IN that plane
// rather than sticking out of it.
//
// A plane contains a direction exactly when its axis is perpendicular to that direction, so
// an axis perpendicular to both the edge and up is the one plane holding both. That is their
// cross product, which is why this is the ONLY axis satisfying the pair of constraints —
// there is nothing to choose and no free sign beyond which way the normal faces, and the
// ring is the same disc either way.
//
// Not resolvable when the edge runs straight up (parallel to +y): the cross product vanishes,
// every plane through the edge also contains up, and there is no unique answer to give.
func UprightRingAxis(selfCenter, partnerCenter vec3) (theta, phi float64, ok bool) {
	delta := partnerCenter.Sub(selfCenter)
	if delta.Length() < 1e-9 {
		return 0, 0, false
	}
	n := delta.Normalize().Cross(vec3{X: 0, Y: 1, Z: 0})
	if n.Length() < 1e-6 {
		return 0, 0, false
	}
	u := n.Normalize()
	return math.Acos(geom.Clamp(u.Y, -1, 1)), math.Atan2(u.Z, u.X), true
}

// coplanarNormalTowardPartner (the edge-derived coplanar normal) was REMOVED: the drawn
// coplanar normal is now streamed straight from PairNode's own normalThetaIdx
// (a +90° in θ from that node's own tilt index, decided on that node's
// own goroutine and mirrored by a direct PairNodeSelf.SetTiltIndex call — nodes/PairNode/node.go's
// coplanarNormal, nodes/Wiring/node_mover.go's writeStreamFrame), so it turns WITH the
// tilt instead of holding still toward the partner. See straighten_loop_test.go /
// coplanar_edges_test.go for what replaced the tests that exercised this function.
