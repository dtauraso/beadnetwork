package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// chain_beads_helpers_test.go — shared fixtures used by more than one of this package's
// chain-beads test files.
//
// chainBeads is a pure function of ONE node's own state — its own kind, its own
// neighborKinds, and its own live copy of the neighbour's world center
// (m.partnerCenters) — all written only by that node's goroutine — so these are plain
// tables, no second goroutine (docs/process/testing-shape.md).
//
// Every case sets BOTH kinds explicitly. An unset kind is NOT neutral: kindWidthHeight falls
// back to (110, 60), i.e. radius 15. The original tests left kinds unset and asserted on
// distance from the node CENTER, which is exactly why they passed while beads rendered inside
// the nodes.
//
// A neighbour's position is supplied the same way production reads it now: a live cartesian
// center in m.partnerCenters, placed straight along +Y at the given center-to-center
// distance (an arbitrary but fixed direction; magnitude-only assertions below don't care
// which one). MODEL.md "the polar model": there is no stored node-node bearing any more
// (wire.LocalPolar deleted) — a node's aim to an edge's first bead is measured live.

// singleNeighborCenter builds a partnerCenters map with exactly one entry — `to` placed
// centerGap world units along +Y from the origin (this node's own center, since a bare
// nodeMover literal has HasPos false).
func singleNeighborCenter(to string, centerGap float64) map[string]vec3 {
	return map[string]vec3{to: {X: 0, Y: centerGap, Z: 0}}
}

// THE REGRESSION GUARD for this commit: exact double tangency, no float tolerance wider
// than round-trip noise. Before this commit, edgeStepCount measured
// `round((QuantIR*stepR - nodegeom.NodeTorusOuterR(src) - nodegeom.NodeTorusOuterR(dst)) / BeadStepR)` against
// an nodegeom.NodeTorusOuterR that was an arbitrary float NOT on the bead lattice
// (nodegeom.NodeRadius(kind)*(1+ShadingParamNodeRingTubeRatio)), so the division was essentially never
// exact and the rounding silently absorbed up to half a bead step at the TARGET end — bead 0
// was always exactly tangent to the source (offset by construction), but the last bead's far
// edge only APPROXIMATELY met the target's torus. This test pins the far edge to the target's
// torus to float-round-off tolerance (1e-3, chainBeads' streamed float32 offsets — see
// tangencyEps's own comment), across several distance values this fixture deliberately
// snaps onto the bead lattice (singleNeighborHolder) and several node-kind pairs whose
// bareNodeRadius values don't share an obvious common factor.
const tangencyEps = 1e-3

// --- direction aims at the LIVE partner center, independent of spacing/size ---
//
// offAxisFixture places the live partner center off any coordinate axis, so a chain
// placed along the wrong axis (or a made-up direction) cannot pass this test by accident.
// Spacing/size stay the fixed uniform lattice constants regardless (no per-edge sizing
// any more) — this test is purely about DIRECTION. There is no stored bearing any more to
// diverge from (MODEL.md "the polar model": no node-node stored coordinate) — the aim is
// ALWAYS this live measurement.

// offAxisFixture builds a *nodeGeometry for one edge "a"->"b" whose LIVE partnerCenters
// direction sits off to the side in the X/Z plane at colatitude ~53.13 degrees (a 3-4-5
// triangle's angle, chosen only because it is not a whole degree and not a special angle,
// so no accidental alignment). The live center is placed at EXACTLY count*BeadStepR + both
// tori (on-lattice, by this fixture's own construction) so the far-edge tangency assertion
// below has no residue to tolerate beyond float round-off.
func offAxisFixture(srcKind, dstKind string, count int) *nodeGeometry {
	selfTorus := nodegeom.NodeTorusOuterR(srcKind)
	dstTorus := nodegeom.NodeTorusOuterR(dstKind)
	dist := selfTorus + float64(count)*lattice.BeadStepR + dstTorus
	// Live direction: (3,0,4)/5 — a unit vector off any coordinate axis, with no sqrt
	// needed to state exactly (a 3-4-5 triangle), scaled to the exact required
	// center-to-center distance.
	targetCenter := vec3{X: dist * 0.6, Y: 0, Z: dist * 0.8}
	return &nodeGeometry{
		id:   "a",
		geom: nodegeom.NodeGeom{NodeIdentity: nodegeom.NodeIdentity{Kind: srcKind}}, // HasPos false -> center at origin
		outs: nodeOuts{outTargets: []string{"b"}},
		topo: neighborTopology{
			neighborKinds:  map[string]string{"b": dstKind},
			partnerCenters: map[string]vec3{"b": targetCenter},
		},
	}
}
