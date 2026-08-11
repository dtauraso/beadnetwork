// move_streams.go — stream/seed wiring & setup: dedicated per-mover fd wiring
// (SetMsgTap/SetEdgeStreams/SetNodeStreams). Load-time seed geometry (NodeGeomSeed/
// EdgeGeomSeed types and their NodeSeedsFn/EdgeSeedsFn/LoadTimeCenters accessors) lives
// on md.GS directly (nodes/Wiring/geomseeds) — no MoveDispatch delegator.

package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// SetMsgTap installs (or clears, with nil) the test-only message-trace hook, on md.tapToInstall
// AND on every already-constructed nodeMover's own nm.msg.tap field. MUST be called before
// Start (a setup-goroutine write to each mover's plain field is safe only because it
// happens-before the mover goroutines are launched; there is no concurrent access once
// Start has run). Test-only — production code never calls this.
func (md *MoveDispatch) SetMsgTap(tap func(destID string, msg movemsg.Msg)) {
	md.tapToInstall = tap
	for _, nm := range md.MR.NodeGeoms() {
		nm.SetMsgTap(tap)
	}
}

// SetEdgeStreams wires every edgeMover to ITS OWN dedicated fd — the per-edge stream
// (memory/feedback_no_single_writer_bridge.md): fd = baseFd + row, where row is the
// STABLE edge-seed order (md.GS.EdgeSeeds, the same spec order the Edge
// block uses — see runtopology's md.GS.EdgeSeedsFn() seed loop). buildFrame is an injected func
// (not a Buffer import) so this package stays Buffer-independent, packing the combined
// per-edge frame bytes (Buffer.BuildEdgeStreamFrame) straight from the edge's own SEGMENT
// endpoints (docs/bead-model/channels-not-ports.md — there is no port row to resolve any more).
// Edge selection is NOT injected: each edgeMover owns its OWN selected bit, set via a
// movemsg.KindSelect message on its extIn (md.sendEdgeSelect), not a lookup. Call once at
// startup after LoadTopology, before Start — mirrors SetPortRowResolver/
// SetEdgeRowResolver's call site in main.go. A missing edgeMover for a seed row (should
// not happen) is skipped rather than panicking.
func (md *MoveDispatch) SetEdgeStreams(
	baseFd int,
	nodeRowFor func(id string) (int32, bool),
	buildFrame func(tick uint32, sx, sy, sz, ex, ey, ez float32, selected uint8, label string, events []wire.RowEvent) []byte,
) {
	md.Sw.SetEdgeStreams(md.GS.EdgeSeeds, md.MR.EdgeMovers(), baseFd, nodeRowFor, buildFrame)
}

// SetNodeStreams wires every nodeMover to ITS OWN dedicated node-fd (geometry+ports+
// label), AND wires the interiorOuts directory + buildInteriorFrame func every node's own
// Update-loop closures (builders.go's injectClosures) look up for its own dedicated
// interior-fd — the two emitting goroutines per node (memory/feedback_no_single_writer_bridge.md).
// nodeBase/interiorBase are the two fd ranges' base fds; row is the STABLE node-seed
// order (md.GS.NodeSeeds, the same spec order the Node block uses — see
// runtopology's md.GS.NodeSeedsFn() seed loop). nodeRowFor/buildFrame/
// buildInteriorFrame are injected funcs (not a Buffer import), matching SetEdgeStreams'
// existing pattern. Selection/hover/abc-drag/kind are NOT injected lookups: each nodeMover
// owns its OWN selected/hovered/latchedSel/gotDragMsg/dragDelta*/kindID fields, set via
// movemsg.KindSelect/Hover/Latched/AbcReset messages (or, for kindID, once here at
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
	buildFrame nodeactor.NodeFrameBuilder,
	buildInteriorFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte,
	kindIDFor func(kind string) uint8,
) {
	md.Sw.SetNodeStreams(md.GS.NodeSeeds, md.MR.NodeGeoms(), nodeBase, interiorBase, driveBase, driveWired, nodeRowFor, buildFrame, buildInteriorFrame, kindIDFor)
}

// NodeSeeds/EdgeSeeds/loadTimeCenters were deleted as pure one-line forwards to md.GS
// (nodes/Wiring/geomseeds). GS is an exported field, so every caller — in-package and
// out-of-package (runtopology) alike — now calls md.GS.NodeSeedsFn() /
// md.GS.EdgeSeedsFn() / md.GS.LoadTimeCenters() directly with no new export needed.
