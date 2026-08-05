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
// point to the target INPUT port's sphere-surface point. nodeWorldPos, nodeRadius,
// portDir and portWorldPos here feed arcLengthBetweenPorts
// (loader.go / stdin_reader.go), which returns the chord length.
//
// Inputs the geometry needs, per node:
//   - kind        → width/height via kindDims (generated from SPEC.md View)
//   - center      → world center (from meta.json x/y/z or origin fallback)
//   - port lists  → inputs/outputs with optional side + slot (from the spec node;
//                   falls back to registry ports with default sides when absent)
//
// Every magic number is pulled from CurveParam* constants in curve_params.go.

package Wiring

import (
	"math"
	"strconv"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// nodeIdentity is the WRITE-ONCE-AT-CONSTRUCTION part of a node's geometry: set by the
// loader (loader.go) when the nodeGeom is built and never written again by any handler
// (applyCenter, setPortAnchorId, emitGeometry — grepped clean of writes to these fields).
// It is split out from nodeGeom specifically so a reader that only wants IDENTITY (e.g.
// MoveDispatch.NodeKind, called from the gesture/stdin-reader goroutine, NOT the mover's
// own goroutine) can read it safely: the memory it touches is never in a writer's
// footprint, by construction of the type, not by coincidence of which fields a particular
// access happens to touch. See node_mover.go's geomMu doc-comment history for why a
// "the two byte ranges just don't happen to overlap today" argument is the bug class this
// split exists to make unrepresentable (memory/feedback_make_bug_class_unrepresentable.md).
type nodeIdentity struct {
	Kind  string
	Label string   // human label for this node (data.label, else node id); streamed on node-geometry events for the new-system label sidecar
	R     *float64 // optional per-node sphere radius for this node's edges; nil → defaultNodeR (see nodeR)
	// SceneCenter is the scene sphere's center — the ONLY cartesian value carried here.
	// Set once at construction (loader.go) alongside the rest of identity; never
	// reassigned afterward (grepped clean — no `.SceneCenter =` outside the literal).
	SceneCenter vec3
}

// nodeGeom carries everything the port-curve math needs for one node: the write-once
// nodeIdentity (embedded, so its fields read/promote as g.Kind/g.Label/g.R/g.SceneCenter)
// plus the MUTABLE per-node state that applyCenter/setPortAnchorId update on every move.
//
// Position is POLAR (polar-frame-rewrite.md): ScenePolar (r,θ,φ) about SceneCenter is the
// source of truth; the node's world center is DERIVED only at the display/geometry boundary
// as SceneCenter + polar2cart(ScenePolar) (nodeWorldPos). HasPos is false for
// hand-written/partial specs that carry no position (nodeWorldPos then falls back to the
// world origin).
type nodeGeom struct {
	nodeIdentity
	// ScenePolar is the node's position as (r,θ,φ) about SceneCenter — the polar source of
	// truth. World is derived: SceneCenter + polar2cart(ScenePolar). Valid only when HasPos.
	// Mutated only by setNodeWorld (applyCenter's sole write path), on the node's own
	// mover goroutine.
	ScenePolar polar
	HasPos     bool // false for hand-written/partial specs with no position → nodeWorldPos returns origin
	// ReachR is the sphere REACH radius: the max distance from this node's center to
	// any node it outputs to (its surface children), under the resolved centers. It is
	// streamed in the NodeGeometry sphereR field and consumed by the TS SphereRing so the
	// "show the sphere" ring reaches every surface node even when a child was placed by a
	// different parent. 0 when the node has no outgoing edges (childless).
	ReachR float64
}

// defaultNodeR is the default starting sphere radius (world units) used for a
// node that omits an explicit r. Tunable — chosen as a sensible starting size
// for the polar layout.
const defaultNodeR = 200.0

// nodeR returns the node's sphere radius: *g.R when set, else defaultNodeR.
func nodeR(g nodeGeom) float64 {
	if g.R != nil {
		return *g.R
	}
	return defaultNodeR
}

// Interior bead render dimensions — mirror scene-content.tsx INTERIOR_BEAD_R +
// torus tube fraction; keep in sync. Each interior bead draws a sphere of radius
// interiorBeadR PLUS a torus ring whose OUTER radius is
// interiorBeadR*(1+interiorTorusTubeFrac). The slot pitch must space by the torus
// outer radius (the larger extent), not the sphere, so adjacent rings don't touch.
const (
	interiorBeadR         = 5.0  // sphere radius (INTERIOR_BEAD_R)
	interiorTorusTubeFrac = 0.12 // torus tube fraction; outer radius = r*(1+frac)
	interiorBeadGap       = 0.2  // small gap BETWEEN adjacent toruses
)

// interiorTorusOuterR is the torus outer radius — the bead's true visual extent.
const interiorTorusOuterR = interiorBeadR * (1 + interiorTorusTubeFrac) // 5.6

// interiorSlot is the 2x2 grid half-pitch, computed TORUS-AWARE from the bead's
// torus outer radius plus half the desired inter-torus gap. Adjacent same-row
// beads are 2*interiorSlot apart, so torus-to-torus gap = 2*interiorSlot -
// 2*rt = interiorBeadGap. Pitch follows bead size (beads are a fixed visual
// size), NOT the node radius — nodeRadius is used only for the wall-fit guarantee.
const interiorSlot = interiorTorusOuterR + interiorBeadGap/2 // 5.9

// interiorSlotOffset returns the NODE-LOCAL OFFSET of the 2x2 interior grid slot
// at (row, col), relative to the node center (NOT a world position): row 0 =
// top/backup, row 1 = bottom/working; col 0 = left, col 1 = right. The grid is
// sized by the bead's TORUS OUTER RADIUS so adjacent rings keep a small gap and
// never overlap:
//
//	slot   = interiorTorusOuterR + interiorBeadGap/2
//	dx = (col - 0.5) * 2*slot
//	dy = (0.5 - row) * 2*slot
//	dz = 0
//
// The grid is centered on the node, so offsets are symmetric about (0,0). TS
// renders the bead as a child of the node group, so its world position =
// node center + offset is composed by the scene graph (no node center added on
// the Go side). Discrete — beads snap to these slot centers. The corner bead's
// torus reach (|offset| + rt) must stay inside the node sphere radius r (see
// TestInteriorBeadsInsideSphere). The Z offset is always 0 (grid is planar).
func interiorSlotOffset(row, col int) vec3 {
	slot := interiorSlot
	pitch := 2 * slot
	return vec3{
		X: (float64(col) - 0.5) * pitch,
		Y: (0.5 - float64(row)) * pitch,
		Z: 0,
	}
}

// kindWidthHeight returns the render width/height for a kind, mirroring the
// TS defaults (width ?? 110, height ?? 60) when the kind is unknown.
func kindWidthHeight(kind string) (float64, float64) {
	if d, ok := kindDims[kind]; ok {
		return d.Width, d.Height
	}
	return 110, 60
}

// bareNodeRadius is the UNSNAPPED sphere radius from a kind's width/height —
// min(width, height) / CurveParamNodeRadiusDivisor, mirroring nodeRadius() in
// geometry-helpers.ts. It exists ONLY as the basis nodeTorusSteps snaps to the bead
// lattice below; nothing else may call it. Every other reader of "this kind's
// radius" must go through nodeRadius (which is the SNAPPED value, derived from
// nodeTorusOuterR) — a second, unsnapped copy of the radius reaching a renderer or
// a placement calculation is exactly the half-bead-step drift docs/bead-lattice.md
// exists to remove, so this helper is deliberately unexported and single-purpose.
func bareNodeRadius(kind string) float64 {
	w, h := kindWidthHeight(kind)
	return min(w, h) / float64(CurveParamNodeRadiusDivisor)
}

// nodeRadius is a node's SPHERE radius — the streamed/drawn radius, and the basis
// for ring-anchor placement (ringAnchorCount, portRingPolar, snapToRingAnchorIndex).
// It is DERIVED from the snapped torus extent (nodeTorusOuterR), by inverting the
// ring's tube-fraction scale, rather than computed independently from
// width/height: the TS renderer draws the border ring as a unit torus scaled by
// this exact value with tube thickness ShadingParamNodeRingTubeRatio
// (NodeInstances.tsx), so ring-outer-radius = nodeRadius(kind) *
// (1+ShadingParamNodeRingTubeRatio) = nodeTorusOuterR(kind) by construction — the
// drawn ring and the bead-tangent point can never disagree, because both trace back
// to the one snapped integer nodeTorusSteps. Nodes change size by up to one bead
// step versus the pre-snap width/height formula; that is the intended cost of
// making the tangency exact (docs/bead-lattice.md "Placement").
func nodeRadius(kind string) float64 {
	return nodeTorusOuterR(kind) / (1 + ShadingParamNodeRingTubeRatio)
}

// effectiveRadius returns the node's REACH radius (max distance to a surface child),
// falling back to nodeR for childless nodes (ReachR == 0) so the value stays sane. Used
// by nodeMover.writeStreamFrame (sphereR) and the load-time node-seed build (node_move.go).
func effectiveRadius(g nodeGeom) float64 {
	if g.ReachR > 0 {
		return g.ReachR
	}
	return nodeR(g)
}

// nodeWorldPos derives a node's world center from its polar source of truth:
// SceneCenter + polar2cart(ScenePolar). This is the ONE polar→cartesian conversion for a
// node center; it happens only here, at the geometry/display boundary. A node with no
// position (HasPos false — hand-written/partial specs) falls back to the world origin.
func nodeWorldPos(g nodeGeom) vec3 {
	if !g.HasPos {
		return vec3{}
	}
	return g.SceneCenter.Add(polar2cart(g.ScenePolar))
}

// setNodeWorld updates a node's polar source of truth from a world center at an INPUT
// boundary (a pointer-derived world point, or a re-propagated solve center). The held
// representation stays polar: ScenePolar = cart2polar(world − SceneCenter). Cartesian
// enters only here and at nodeWorldPos — never as a stored cartesian center.
func setNodeWorld(g *nodeGeom, world vec3) {
	g.ScenePolar = cart2polar(world.Sub(g.SceneCenter))
	g.HasPos = true
}

// edgeSegment is the straight world segment the renderer draws for an edge: NODE SURFACE
// TO NODE SURFACE along the centre-to-centre line (docs/channels-not-ports.md — a port is
// a load-time channel-binding ROLE now, never a place, so it contributes no geometry to
// this segment at all). start = the source node's center, moved out to its own
// nodeTorusOuterR toward the target; end = the target's center, moved out to ITS
// nodeTorusOuterR toward the source. These are the SAME two surface points
// chain_beads.go anchors bead 0 and the last bead to (docs/bead-lattice.md "Placement":
// "Bead 0's torus is tangent to the source node's torus... bead N-1's torus is tangent to
// the target node's torus, EXACTLY") — this is deliberate, not incidental: the edge
// segment and the bead chain must measure between the identical two points, which is
// exactly the invariant the old port-radius offset broke (the chain measured node-torus
// to node-torus while the port sat proud of/inside that surface, so the first and last
// bead were off by the port's own radius while interior spacing stayed correct).
func edgeSegment(src, tgt nodeGeom) wireSegment {
	srcCenter := nodeWorldPos(src)
	tgtCenter := nodeWorldPos(tgt)
	dir := tgtCenter.Sub(srcCenter)
	if dir.Length() < 1e-9 {
		// Degenerate (coincident centers, e.g. a not-yet-positioned node): fall back to the
		// bare centers rather than dividing by a near-zero length.
		return wireSegment{Start: srcCenter, End: tgtCenter}
	}
	unit := dir.Normalize()
	start := srcCenter.Add(unit.Scale(nodeTorusOuterR(src.Kind)))
	end := tgtCenter.Sub(unit.Scale(nodeTorusOuterR(tgt.Kind)))
	return wireSegment{Start: start, End: end}
}

// nodeTorusSteps is a node's torus-outer extent expressed as a whole number of bead
// steps — the integer edgeStepCount subtracts from an edge's separation
// (docs/bead-lattice.md "The count"). ROUND, not ceil: this used to ceil, on the
// reasoning that rounding down could snap the extent smaller than the node's true
// unsnapped body and let a bead's tangent point land inside it. That reasoning was
// wrong — there IS no unsnapped body left to protect. nodeRadius (below) is DERIVED
// from nodeTorusOuterR, which is derived from THIS value, so the node's drawn
// sphere/ring always follows the snapped extent, never the raw width/height number;
// nothing is ever measured against the unsnapped `unsnapped` local below except this
// one snap. Ceil therefore did nothing but inflate every node (a ~46.5-world-unit
// unsnapped radius snapped to 6 bead distances, 53.8, instead of the nearer 5,
// 44.8) with no corresponding safety, which fails the "node sizes must not grow"
// requirement. Round keeps the node closest to its authored width/height size.
func nodeTorusSteps(kind string) int {
	unsnapped := bareNodeRadius(kind) * (1 + ShadingParamNodeRingTubeRatio)
	return int(math.Round(unsnapped / wire.BeadStepR))
}

// nodeTorusOuterR is a node's TORUS OUTER radius, SNAPPED to a whole number of bead
// steps (nodeTorusSteps) — its true visual/geometric extent, not the unsnapped
// width/height formula (docs/bead-lattice.md "Placement"). nodeRadius (above) is
// DERIVED from this value, so the node's drawn sphere/ring and the bead-tangent
// point at nodeTorusOuterR(kind) can never disagree: there is one snapped number,
// not a snapped one for beads and an independently-rounded one for the renderer.
func nodeTorusOuterR(kind string) float64 {
	return float64(nodeTorusSteps(kind)) * wire.BeadStepR
}

// edgeCenterDistAndDir returns the LIVE center-to-center distance BETWEEN two nodes AND
// the live unit direction from selfCenter toward targetCenter, from their live cartesian
// world centers — ONE measurement of the edge, not two. selfCenter/targetCenter must be
// nodeWorldPos of each node, the SAME function edgeSegment (above) and every emitGeometry
// call use, so this reads the identical value the renderer draws the node at — not the
// SOURCE node's stored, quantized LocalPolar (lp.QuantIR*StepR and its QuantITheta/
// QuantIPhi bearing), which is an integer-step APPROXIMATION of both this distance and
// this direction (1-degree angular cells), which can drift from the live geometry between
// drags. chain_beads.go reads the LIVE value (edgeStepCount's `dist`) rather than the
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
// (tools/check-no-sqrt-in-chain-beads.sh, "index arithmetic, trig only at the
// polar2cart boundary" — memory/feedback_abc_times_constant_not_rederive.md).
// chainBeads calls this helper and receives only the resulting scalar distance and unit
// vector; the sqrt itself lives here, in the file that already computes edgeSegment the
// same way.
func edgeCenterDistAndDir(selfCenter, targetCenter vec3) (dist float64, unitDir vec3, ok bool) {
	delta := targetCenter.Sub(selfCenter)
	length := delta.Length()
	if length < 1e-9 {
		return 0, vec3{}, false
	}
	return length, delta.Normalize(), true
}

// parallelChainOffset is the perpendicular displacement a node applies to its OWN chain
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
// The magnitude is one bead STEP each way, so the two chains end up exactly 2*wire.BeadStepR
// apart — still on the lattice, not a tuned pixel gap. It was half that (one bead radius
// each way, the chains exactly touching), which separated them in principle but read as one
// thick wire; a full step each way leaves a clear bead-sized gap between the two chains.
//
// The cost, stated rather than hidden: an offset chain no longer starts exactly tangent to
// its node's torus, since it is displaced off the centre line that tangency is measured on.
// That is the trade the separation buys, and it applies ONLY to a mutual pair.
func parallelChainOffset(selfID, targetID string, selfCenter, targetCenter, sceneCenter vec3) (vec3, bool) {
	lowCenter, highCenter := selfCenter, targetCenter
	if !nodeIDLess(selfID, targetID) {
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
	if !nodeIDLess(selfID, targetID) {
		sign = -1.0
	}
	return perp.Normalize().Scale(sign * wire.BeadStepR), true
}

// nodeIDLess orders two node ids NUMERICALLY, because node ids are numbers that are strings
// only because they are directory names (.claude/rules/persistence-ownership.md). A plain
// string compare would order "10" before "2" and hand both ends of that pair the same sign,
// which is the one thing parallelChainOffset must never do. A non-numeric id (impossible
// today — loadTree rejects one) falls back to the string compare rather than panicking in
// geometry code.
func nodeIDLess(a, b string) bool {
	ai, aerr := strconv.Atoi(a)
	bi, berr := strconv.Atoi(b)
	if aerr == nil && berr == nil {
		return ai < bi
	}
	return a < b
}

// poleContainingEdge returns the ring axis closest to the given pole whose PLANE contains
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
func poleContainingEdge(poleTheta, polePhi float64, selfCenter, partnerCenter vec3) (theta, phi float64, ok bool) {
	delta := partnerCenter.Sub(selfCenter)
	if delta.Length() < 1e-9 {
		return 0, 0, false
	}
	dir := delta.Normalize()
	pole := anglesToWorldOffset(1, poleTheta, polePhi)
	projected := pole.Sub(dir.Scale(pole.Dot(dir)))
	if projected.Length() < 1e-6 {
		return 0, 0, false
	}
	u := projected.Normalize()
	return math.Acos(clamp(u.Y, -1, 1)), math.Atan2(u.Z, u.X), true
}

// torusDefaultAxisAngles is the torus geometry's OWN normal (+Z) as this codebase's angle
// pair. A ring streamed with this axis is drawn exactly as an unrotated one, which is what
// every scene looked like before ring orientation existed — so it is the default, and a
// scene opts IN to anything else.
func torusDefaultAxisAngles() (theta, phi float64) {
	return math.Pi / 2, math.Pi / 2
}

// uprightRingAxis returns the ring axis whose PLANE contains BOTH the edge and world +y —
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
func uprightRingAxis(selfCenter, partnerCenter vec3) (theta, phi float64, ok bool) {
	delta := partnerCenter.Sub(selfCenter)
	if delta.Length() < 1e-9 {
		return 0, 0, false
	}
	n := delta.Normalize().Cross(vec3{X: 0, Y: 1, Z: 0})
	if n.Length() < 1e-6 {
		return 0, 0, false
	}
	u := n.Normalize()
	return math.Acos(clamp(u.Y, -1, 1)), math.Atan2(u.Z, u.X), true
}

// coplanarNormalTowardPartner (the edge-derived coplanar normal) was REMOVED: the drawn
// coplanar normal is now streamed straight from Node1/Node2's own normalThetaIdx/
// normalPhiIdx (a fixed ±90° in θ from that node's own tilt index, decided on that node's
// own goroutine and mirrored via moveMsgKindTiltIndexSync — nodes/Node1/node.go's
// coplanarNormal, nodes/Wiring/node_mover.go's writeStreamFrame), so it turns WITH the
// tilt instead of holding still toward the partner. See straighten_loop_test.go /
// coplanar_edges_test.go for what replaced the tests that exercised this function.
