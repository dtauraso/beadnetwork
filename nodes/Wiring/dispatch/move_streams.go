// move_streams.go — SetMsgTap, the one stream/seed-wiring method that stays a MoveDispatch
// method: it writes md.tapToInstall (forbidden to export). SetEdgeStreams/SetNodeStreams
// were pure single-owner forwards onto md.Sw and were deleted
// (docs/planning/movedispatch-decomposition.md, the remainder cluster) — their two callers
// (runtopology/edge_stream.go, runtopology/node_stream.go) now call
// md.Sw.SetEdgeStreams(md.GS.EdgeSeeds, md.MR.EdgeMovers(), ...) /
// md.Sw.SetNodeStreams(md.GS.NodeSeeds, md.MR.NodeGeoms(), ...) directly. Load-time seed
// geometry (NodeGeomSeed/EdgeGeomSeed types and their NodeSeedsFn/EdgeSeedsFn/
// LoadTimeCenters accessors) lives on md.GS directly (nodes/Wiring/geomseeds) — no
// MoveDispatch delegator.

package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
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
