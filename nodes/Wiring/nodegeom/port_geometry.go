// port_geometry.go — Go-OWNED port-to-port segment geometry.
//
// Go owns geometry in this architecture: it computes each node/port world position
// and streams the result into the content buffer; the TS webview renders from that
// buffer and computes no geometry of its own (guard: check-ts-computes-no-geometry.sh).
// This file is NOT a mirror of a TS counterpart — the former TS port-geometry helpers
// were removed when Go took over, and geometry-helpers.ts now holds only screen-coord
// conversion (ndcToPixel/pixelToNDC) for input picking, nothing this file reproduces.
//
// Go must compute a pulse's travel budget from the SAME segment the
// bead is drawn on: a straight line from the source OUTPUT port's sphere-surface
// point to the target INPUT port's sphere-surface point. NodeWorldPos, NodeRadius,
// portDir and portWorldPos here feed arcLengthBetweenPorts
// (loader.go / stdin_reader.go), which returns the chord length.
//
// Inputs the geometry needs, per node:
//   - kind        → width/height via KindDims (generated from SPEC.md View)
//   - center      → world center (from meta.json x/y/z or origin fallback)
//   - port lists  → inputs/outputs with optional side + slot (from the spec node;
//                   falls back to registry ports with default sides when absent)
//
// Every magic number is pulled from CurveParam* constants in curve_params.go.
//
// A node's OWN geometry (identity, polar position, sphere radius) lives in
// node_geom.go; the interior-bead 2x2 slot grid lives in interior_slot_geometry.go.
// This file keeps the seam it is named for: the geometry BETWEEN two nodes — edge
// segments, the mutual-pair parallel-chain offset, and ring-axis derivation.

package nodegeom

import (
	"math"
	"strconv"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// EdgeSegment is the straight world segment the renderer draws for an edge: NODE SURFACE
// TO NODE SURFACE along the centre-to-centre line (docs/bead-model/channels-not-ports.md — a port is
// a load-time channel-binding ROLE now, never a place, so it contributes no geometry to
// this segment at all). start = the source node's center, moved out to its own
// NodeTorusOuterR toward the target; end = the target's center, moved out to ITS
// NodeTorusOuterR toward the source. These are the SAME two surface points
// chain_beads.go anchors bead 0 and the last bead to (docs/bead-model/bead-lattice.md "Placement":
// "Bead 0's torus is tangent to the source node's torus... bead N-1's torus is tangent to
// the target node's torus, EXACTLY") — this is deliberate, not incidental: the edge
// segment and the bead chain must measure between the identical two points, which is
// exactly the invariant the old port-radius offset broke (the chain measured node-torus
// to node-torus while the port sat proud of/inside that surface, so the first and last
// bead were off by the port's own radius while interior spacing stayed correct).
func EdgeSegment(src, tgt NodeGeom) wireSegment {
	srcCenter := NodeWorldPos(src)
	tgtCenter := NodeWorldPos(tgt)
	dir := tgtCenter.Sub(srcCenter)
	if dir.Length() < 1e-9 {
		// Degenerate (coincident centers, e.g. a not-yet-positioned node): fall back to the
		// bare centers rather than dividing by a near-zero length.
		return wireSegment{Start: srcCenter, End: tgtCenter}
	}
	unit := dir.Normalize()
	start := srcCenter.Add(unit.Scale(NodeTorusOuterR(src.Kind)))
	end := tgtCenter.Sub(unit.Scale(NodeTorusOuterR(tgt.Kind)))
	return wireSegment{Start: start, End: end}
}

// EdgeCenterDistAndDir returns the LIVE center-to-center distance BETWEEN two nodes AND
// the live unit direction from selfCenter toward targetCenter, from their live cartesian
// world centers — ONE measurement of the edge, not two. selfCenter/targetCenter must be
// NodeWorldPos of each node, the SAME function EdgeSegment (above) and every emitGeometry
// call use, so this reads the identical value the renderer draws the node at — not the
// SOURCE node's stored, quantized LocalPolar (lp.QuantIR*StepR and its QuantITheta/
// QuantIPhi bearing), which is an integer-step APPROXIMATION of both this distance and
// this direction (1-degree angular cells), which can drift from the live geometry between
// drags. chain_beads.go reads the LIVE value (EdgeStepCount's `dist`) rather than the
// stored one, so a bearing residue (a chain aimed up to half a degree off, from the
// 1-degree stored cell) can never reappear independent of the length: distance and
// direction are returned from the SAME Length()/Sub() call, so a caller cannot read a
// length from one measurement and a bearing from another.
//
// ok is false only when the centers are degenerate (coincident, e.g. a
// not-yet-positioned node with HasPos false) — direction is undefined at zero
// separation, and the caller falls back to the stored quantized bearing/distance rather
// than dividing by a near-zero length.
//
// This one Length()/Normalize() pair is deliberately NOT in chain_beads.go: that file is
// guarded against math.Sqrt/Vec3.Length/Normalize
// (tools/network/beads/check-no-sqrt-in-chain-beads.sh, "index arithmetic, trig only at the
// polar2cart boundary" — memory/feedback_abc_times_constant_not_rederive.md).
// chainBeads calls this helper and receives only the resulting scalar distance and unit
// vector; the sqrt itself lives here, in the file that already computes EdgeSegment the
// same way.
func EdgeCenterDistAndDir(selfCenter, targetCenter vec3) (dist float64, unitDir vec3, ok bool) {
	delta := targetCenter.Sub(selfCenter)
	length := delta.Length()
	if length < 1e-9 {
		return 0, vec3{}, false
	}
	return length, delta.Normalize(), true
}

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
