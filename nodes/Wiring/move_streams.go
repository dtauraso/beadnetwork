// move_streams.go — stream/seed wiring & setup: dedicated per-mover fd wiring
// (SetMsgTap/SetEdgeStreams/SetNodeStreams) and load-time seed geometry accessors
// (NodeSeeds/EdgeSeeds/loadTimeCenters — the NodeGeomSeed/EdgeGeomSeed types themselves
// now live in nodes/Wiring/geomseeds).

package Wiring

import (
	geomseeds "github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// SetMsgTap installs (or clears, with nil) the test-only message-trace hook, on md.tapToInstall
// AND on every already-constructed nodeMover's own nm.msg.tap field. MUST be called before
// Start (a setup-goroutine write to each mover's plain field is safe only because it
// happens-before the mover goroutines are launched; there is no concurrent access once
// Start has run). Test-only — production code never calls this.
func (md *MoveDispatch) SetMsgTap(tap func(destID string, msg moveMsg)) {
	md.tapToInstall = tap
	for _, nm := range md.mr.nodeGeoms {
		nm.msg.tap = tap
	}
}

// SetEdgeStreams wires every edgeMover to ITS OWN dedicated fd — the per-edge stream
// (memory/feedback_no_single_writer_bridge.md): fd = baseFd + row, where row is the
// STABLE edge-seed order (md.GS.EdgeSeeds, the same spec order the Edge
// block uses — see main.go's md.EdgeSeeds() seed loop). buildFrame is an injected func
// (not a Buffer import) so this package stays Buffer-independent, packing the combined
// per-edge frame bytes (Buffer.BuildEdgeStreamFrame) straight from the edge's own SEGMENT
// endpoints (docs/bead-model/channels-not-ports.md — there is no port row to resolve any more).
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
	md.sw.setEdgeStreams(md.GS.EdgeSeeds, md.mr.edgeMovers, baseFd, nodeRowFor, buildFrame)
}

// SetNodeStreams wires every nodeMover to ITS OWN dedicated node-fd (geometry+ports+
// label), AND wires the interiorOuts directory + buildInteriorFrame func every node's own
// Update-loop closures (builders.go's injectClosures) look up for its own dedicated
// interior-fd — the two emitting goroutines per node (memory/feedback_no_single_writer_bridge.md).
// nodeBase/interiorBase are the two fd ranges' base fds; row is the STABLE node-seed
// order (md.GS.NodeSeeds, the same spec order the Node block uses — see
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
	buildFrame NodeFrameBuilder,
	buildInteriorFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte,
	kindIDFor func(kind string) uint8,
) {
	md.sw.setNodeStreams(md.GS.NodeSeeds, md.mr.nodeGeoms, nodeBase, interiorBase, driveBase, driveWired, nodeRowFor, buildFrame, buildInteriorFrame, kindIDFor)
}

// NodeSeeds returns every node's load-time seed geometry in SPEC ORDER (see
// geomseeds.GeomSeeds.NodeSeeds). Call after LoadTopology returns, before launching any
// node goroutine, and stream each entry via tr.NodeGeometry (main.go). Thin delegator to
// md.GS (nodes/Wiring/geomseeds).
func (md *MoveDispatch) NodeSeeds() []geomseeds.NodeGeomSeed { return md.GS.NodeSeedsFn() }

// loadTimeCenters returns the node-id → LOAD-TIME world center map. Thin delegator to
// md.GS (nodes/Wiring/geomseeds).
func (md *MoveDispatch) loadTimeCenters() map[string]vec3 { return md.GS.LoadTimeCenters() }

// EdgeSeeds returns every edge's load-time seed topology (with real endpoint geometry) in
// SPEC ORDER. Call alongside NodeSeeds; stream each entry via tr.Geometry (main.go). Thin
// delegator to md.GS (nodes/Wiring/geomseeds).
func (md *MoveDispatch) EdgeSeeds() []geomseeds.EdgeGeomSeed { return md.GS.EdgeSeedsFn() }
