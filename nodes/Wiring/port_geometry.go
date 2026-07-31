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

// edgeSurfaceGap returns the ACTUAL surface-to-surface distance between two nodes'
// tori, from their live cartesian world centers — the exact gap chain_beads.go's
// chainBeads spans (see its "the last bead's far edge lands exactly on the target's
// torus surface" placement). selfCenter/targetCenter must be nodeWorldPos of each
// node, the SAME function edgeSegment (above) and every emitGeometry call use, so
// this reads the identical value the renderer draws the node at — not the SOURCE
// node's stored, quantized LocalPolar (lp.QuantIR*StepR), which is an integer-step
// APPROXIMATION of this distance and is exactly the value whose rounding residue
// produced the half-bead gap this function exists to close (docs/bead-lattice.md;
// the residue is bounded by half a bead because QuantIR is round(distance/step)).
//
// This one Length() call is deliberately NOT in chain_beads.go: that file is
// guarded against math.Sqrt/Vec3.Length/Normalize
// (tools/check-no-sqrt-in-chain-beads.sh, "index arithmetic, trig only at the
// polar2cart boundary" — memory/feedback_abc_times_constant_not_rederive.md).
// chainBeads calls this helper and receives only the resulting scalar gap; the
// sqrt itself lives here, in the file that already computes edgeSegment the same
// way.
func edgeSurfaceGap(selfCenter, targetCenter vec3, selfTorusR, targetTorusR float64) float64 {
	gap := targetCenter.Sub(selfCenter).Length() - selfTorusR - targetTorusR
	if gap < 0 {
		gap = 0
	}
	return gap
}
