// move_streams.go — stream/seed wiring & setup: dedicated per-mover fd wiring
// (SetMsgTap/SetEdgeStreams/SetNodeStreams) and load-time seed geometry accessors
// (NodeGeomSeed/EdgeGeomSeed, NodeSeeds/EdgeSeeds/loadTimeCenters).

package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// SetMsgTap installs (or clears, with nil) the test-only message-trace hook, on md.tapToInstall
// AND on every already-constructed nodeMover's own nm.tap field. MUST be called before
// Start (a setup-goroutine write to each mover's plain field is safe only because it
// happens-before the mover goroutines are launched; there is no concurrent access once
// Start has run). Test-only — production code never calls this.
func (md *MoveDispatch) SetMsgTap(tap func(destID string, msg moveMsg)) {
	md.tapToInstall = tap
	for _, nm := range md.mr.nodeMovers {
		nm.tap = tap
	}
}

// SetEdgeStreams wires every edgeMover to ITS OWN dedicated fd — the per-edge stream
// (memory/feedback_no_single_writer_bridge.md): fd = baseFd + row, where row is the
// STABLE edge-seed order (md.gs.edgeSeeds, the same spec order the Edge
// block uses — see main.go's md.EdgeSeeds() seed loop). buildFrame is an injected func
// (not a Buffer import) so this package stays Buffer-independent, packing the combined
// per-edge frame bytes (Buffer.BuildEdgeStreamFrame) straight from the edge's own SEGMENT
// endpoints (docs/channels-not-ports.md — there is no port row to resolve any more).
// Edge selection is NOT injected: each edgeMover owns its OWN selected bit, set via a
// moveMsgKindSelect message on its extIn (md.sendEdgeSelect), not a lookup. Call once at
// startup after LoadTopology, before Start — mirrors SetPortRowResolver/
// SetEdgeRowResolver's call site in main.go. A missing edgeMover for a seed row (should
// not happen) is skipped rather than panicking.
func (md *MoveDispatch) SetEdgeStreams(
	baseFd int,
	nodeRowFor func(id string) (int32, bool),
	buildFrame func(tick uint32, sx, sy, sz, ex, ey, ez float32, selected uint8, label string, events []wire.RowEvent) []byte,
) {
	md.sw.setEdgeStreams(md.gs.edgeSeeds, md.mr.edgeMovers, baseFd, nodeRowFor, buildFrame)
}

// SetNodeStreams wires every nodeMover to ITS OWN dedicated node-fd (geometry+ports+
// label), AND wires the interiorOuts directory + buildInteriorFrame func every node's own
// Update-loop closures (builders.go's injectClosures) look up for its own dedicated
// interior-fd — the two emitting goroutines per node (memory/feedback_no_single_writer_bridge.md).
// nodeBase/interiorBase are the two fd ranges' base fds; row is the STABLE node-seed
// order (md.gs.nodeSeeds, the same spec order the Node block uses — see
// main.go's md.NodeSeeds() seed loop). nodeRowFor/buildFrame/
// buildInteriorFrame are injected funcs (not a Buffer import), matching SetEdgeStreams'
// existing pattern. Selection/hover/abc-drag/kind are NOT injected lookups: each nodeMover
// owns its OWN selected/hovered/latchedSel/gotDragMsg/dragDelta*/kindID fields, set via
// moveMsgKindSelect/Hover/Latched/AbcReset messages (or, for kindID, once here at
// construction — a node's kind never changes after load, so there is no lookup to
// perform on every emit). Call once at startup after LoadTopology, before Start — mirrors
// SetEdgeStreams' call site in main.go. A missing nodeMover for a seed row (should not
// happen) is skipped rather than panicking.
// driveBase/driveWired add the per-DriveHeld-goroutine "drive" fd range (see
// setNodeStreams' own doc comment and Buffer.StreamKindDrive) — driveWired false leaves
// every node's driveOuts slots nil (no dedicated fd, fallback no-op writes).
func (md *MoveDispatch) SetNodeStreams(
	nodeBase, interiorBase, driveBase int,
	driveWired bool,
	nodeRowFor func(id string) (int32, bool),
	buildFrame func(tick uint32, nodeRow int32, nodeID int32, cx, cy, cz, radius, sphereR float32, vrx, vry, vrz, frx, fry, frz float32, poleTheta, polePhi, ringAxisTheta, ringAxisPhi, vectorLen, vectorTheta, vectorPhi float32, selected, kindID, hovered, latchedSel uint8, label string, chainBeadOX, chainBeadOY, chainBeadOZ []float32, chainBeadLit []uint8, chainBeadLitValue []int32, events []wire.RowEvent) []byte,
	buildInteriorFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte,
	kindIDFor func(kind string) uint8,
) {
	md.sw.setNodeStreams(md.gs.nodeSeeds, md.mr.nodeMovers, nodeBase, interiorBase, driveBase, driveWired, nodeRowFor, buildFrame, buildInteriorFrame, kindIDFor)
}

// NodeGeomSeed is one node's load-time seed geometry, exported in spec order and consumed
// by main.go's pre-launch tr.NodeGeometry loop (see the row-seeding comment in main.go).
// No port geometry rides here any more (docs/channels-not-ports.md — a port carries none).
type NodeGeomSeed struct {
	ID, Label, Kind              string
	CX, CY, CZ, Radius, SphereR  float64
	VRX, VRY, VRZ, FRX, FRY, FRZ float64
	// Row is this node's buffer ROW INDEX: id-1 (ROW ID = NODE ID - 1 — the row is declared
	// by the id, never derived by sorting/position in this slice). Two nodes are never at
	// the same Row (loadTree rejects a duplicate id); a gap in the id space is simply a row
	// no seed claims, not a shift of later rows.
	Row int
}

// EdgeGeomSeed is one edge's load-time topology AND its real segment endpoints — the same
// edgeSegment(srcGeom, dstGeom) computation the edge's own live recomputeGeometry
// (node_move.go) uses, evaluated here against the load-time geoms so the seed row is never a
// degenerate 0,0,0→0,0,0 segment.
type EdgeGeomSeed struct {
	Label, SrcNode, DstNode string
	SX, SY, SZ, EX, EY, EZ  float64
}

// NodeSeeds returns every node's load-time seed geometry in SPEC ORDER (see
// geomSeeds.nodeSeeds). Call after LoadTopology returns, before launching any node
// goroutine, and stream each entry via tr.NodeGeometry (main.go). Thin delegator to
// md.gs (geom_seeds.go).
func (md *MoveDispatch) NodeSeeds() []NodeGeomSeed { return md.gs.nodeSeedsFn() }

// loadTimeCenters returns the node-id → LOAD-TIME world center map. Thin delegator to
// md.gs (geom_seeds.go).
func (md *MoveDispatch) loadTimeCenters() map[string]vec3 { return md.gs.loadTimeCenters() }

// EdgeSeeds returns every edge's load-time seed topology (with real endpoint geometry) in
// SPEC ORDER. Call alongside NodeSeeds; stream each entry via tr.Geometry (main.go). Thin
// delegator to md.gs (geom_seeds.go).
func (md *MoveDispatch) EdgeSeeds() []EdgeGeomSeed { return md.gs.edgeSeedsFn() }
