// move_streams.go — stream/seed wiring & setup: dedicated per-mover fd wiring
// (SetMsgTap/SetEdgeStreams/SetNodeStreams) and load-time seed geometry accessors
// (NodeGeomSeed/EdgeGeomSeed, NodeSeeds/EdgeSeeds/loadTimeCenters).

package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
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
// block uses — see main.go's md.EdgeSeeds() seed loop). portRowFor/buildFrame are
// injected funcs (not a Buffer import) so this package stays Buffer-independent,
// matching PortRowResolver/EdgeRowResolver's existing pattern: portRowFor resolves
// (node,port,isInput) to a buffer PORT-ROW index (mirroring the old central accumulator's PortRowFor), and
// buildFrame packs the combined per-edge frame bytes (Buffer.BuildEdgeStreamFrame).
// Edge selection is NOT injected: each edgeMover owns its OWN selected bit, set via a
// moveMsgKindSelect message on its extIn (md.sendEdgeSelect), not a lookup. Call once at
// startup after LoadTopology, before Start — mirrors SetPortRowResolver/
// SetEdgeRowResolver's call site in main.go. A missing edgeMover for a seed row (should
// not happen) is skipped rather than panicking.
func (md *MoveDispatch) SetEdgeStreams(
	baseFd int,
	portRowFor func(node, port string, isInput bool) (int32, bool),
	nodeRowFor func(id string) (int32, bool),
	buildFrame func(tick uint32, srcPortRow, dstPortRow int32, selected uint8, label string, edgeLen float32, groupIdx int32, beadVal []int32, beadX, beadY, beadZ []float32, events []wire.RowEvent) []byte,
) {
	md.sw.setEdgeStreams(md.gs.edgeSeeds, md.mr.edgeMovers, baseFd, portRowFor, nodeRowFor, buildFrame)
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
func (md *MoveDispatch) SetNodeStreams(
	nodeBase, interiorBase int,
	nodeRowFor func(id string) (int32, bool),
	buildFrame func(tick uint32, nodeRow int32, cx, cy, cz, radius, sphereR float32, vrx, vry, vrz, frx, fry, frz float32, selected, kindID, hovered, latchedSel, gotDragMsg uint8, dragDeltaA, dragDeltaB, dragDeltaC, dragRequantCount int32, gotForwardMsg uint8, forwardDeltaA, forwardDeltaB, forwardDeltaC, forwardFromRow int32, label string, portNames []string, portDX, portDY, portDZ, portPX, portPY, portPZ []float32, portIsInput, portHovered []uint8, dstNodeRows []int32, events []wire.RowEvent) []byte,
	buildInteriorFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte,
	kindIDFor func(kind string) uint8,
) {
	md.sw.setNodeStreams(md.gs.nodeSeeds, md.mr.nodeMovers, nodeBase, interiorBase, nodeRowFor, buildFrame, buildInteriorFrame, kindIDFor)
}

// NodeGeomSeed is one node's load-time seed geometry, exported in spec order and consumed
// by main.go's pre-launch tr.NodeGeometry loop (see the row-seeding comment in main.go).
// Ports are the SAME aimed-vs-static port geometry (buildPortGeoms/aimedPortPosDir,
// builders.go) the node's own live emit later produces — computed here from the same
// load-time geoms map, since every node's center is already known at load (buildPartnerCenterFn
// resolves partner centers straight off geoms, no goroutine needed). main.go copies these
// fields into the tr.NodeGeometry call (which additionally resolves the numeric KindID,
// since that table lives in Buffer).
type NodeGeomSeed struct {
	ID, Label, Kind              string
	CX, CY, CZ, Radius, SphereR  float64
	Ports                        []T.PortGeom
	VRX, VRY, VRZ, FRX, FRY, FRZ float64
}

// EdgeGeomSeed is one edge's load-time topology AND its real segment endpoints — the same
// edgeSegment(srcGeom, dstGeom, srcH, dstH) computation the edge's own live recomputeGeometry
// (node_move.go) uses, evaluated here against the load-time geoms so the seed row is never a
// degenerate 0,0,0→0,0,0 segment.
type EdgeGeomSeed struct {
	Label, SrcNode, DstNode string
	// SrcPort/DstPort are the edge's source (output) and dest (input) port NAMES —
	// the edgeMover's own dedicated stream frame resolves these to buffer PORT-ROW
	// indices (see Trace.Geometry).
	SrcPort, DstPort       string
	SX, SY, SZ, EX, EY, EZ float64
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
