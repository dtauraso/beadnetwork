// node_geom.go — a node's own geometry: identity, polar position, and sphere radius.
//
// Split out of port_geometry.go (god-object decomposition, pure move — no logic changes):
// this file owns what ONE node's own geometry IS (nodeIdentity/nodeGeom, its polar
// position and world-center derivation, and its sphere/torus radius). port_geometry.go
// keeps the seam it is named for — the geometry BETWEEN two nodes (edge segments,
// parallel-chain offsets, ring-axis derivation).

package Wiring

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
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
	ScenePolar geom.Polar
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
// a placement calculation is exactly the half-bead-step drift docs/bead-model/bead-lattice.md
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
// making the tangency exact (docs/bead-model/bead-lattice.md "Placement").
func nodeRadius(kind string) float64 {
	return nodeTorusOuterR(kind) / (1 + ShadingParamNodeRingTubeRatio)
}

// effectiveRadius returns the node's REACH radius (max distance to a surface child),
// falling back to nodeR for childless nodes (ReachR == 0) so the value stays sane. Used
// by nodeMover.writeStreamFrame (sphereR) and the load-time node-seed build (move_dispatch_construct.go).
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
	return g.SceneCenter.Add(geom.Polar2cart(g.ScenePolar))
}

// setNodeWorld updates a node's polar source of truth from a world center at an INPUT
// boundary (a pointer-derived world point, or a re-propagated solve center). The held
// representation stays polar: ScenePolar = cart2polar(world − SceneCenter). Cartesian
// enters only here and at nodeWorldPos — never as a stored cartesian center.
func setNodeWorld(g *nodeGeom, world vec3) {
	g.ScenePolar = geom.Cart2polar(world.Sub(g.SceneCenter))
	g.HasPos = true
}

// nodeTorusSteps is a node's torus-outer extent expressed as a whole number of bead
// steps — the integer edgeStepCount subtracts from an edge's separation
// (docs/bead-model/bead-lattice.md "The count"). ROUND, not ceil: this used to ceil, on the
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
	return int(math.Round(unsnapped / lattice.BeadStepR))
}

// nodeTorusOuterR is a node's TORUS OUTER radius, SNAPPED to a whole number of bead
// steps (nodeTorusSteps) — its true visual/geometric extent, not the unsnapped
// width/height formula (docs/bead-model/bead-lattice.md "Placement"). nodeRadius (above) is
// DERIVED from this value, so the node's drawn sphere/ring and the bead-tangent
// point at nodeTorusOuterR(kind) can never disagree: there is one snapped number,
// not a snapped one for beads and an independently-rounded one for the renderer.
func nodeTorusOuterR(kind string) float64 {
	return float64(nodeTorusSteps(kind)) * lattice.BeadStepR
}
