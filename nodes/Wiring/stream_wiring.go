// stream_wiring.go — the per-node interior-stream and dedicated VIEW-stream fd-wiring
// owner split out of MoveDispatch (god-object decomposition), as a pure move (no logic
// changes): streamWiring owns interiorOuts/buildInteriorFrame (the per-node interior fd
// directory) and viewOut/viewBuildFrame/viewTick (the VIEW stream's fd + frame builder +
// local tick counter). MoveDispatch's public SetEdgeStreams/SetNodeStreams methods stay
// as thin delegators so the external API is unchanged. view_stream.go's emitViewFrame
// reads md.sw.viewOut/viewBuildFrame/viewTick directly — it ALSO reads
// md.ui.vp/md.ui.ov/md.ui.sceneSphere, which are owned elsewhere and are NOT part
// of this extraction, so emitViewFrame itself stays a MoveDispatch method rather than
// moving here.

package Wiring

import (
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"io"
	"os"
)

// streamWiring owns the fd directories the two dedicated-per-goroutine emitting streams
// (interior, view) need — see MoveDispatch.sw's doc comment.
type streamWiring struct {
	// interiorOuts holds ONE dedicated per-node interior-bead fd, keyed by node id — the
	// SECOND emitting goroutine per node (its own Update loop, not its nodeMover) writes
	// here (memory/feedback_no_single_writer_bridge.md). Populated ONCE by
	// SetNodeStreams, BEFORE any node's Update goroutine launches (mirrors
	// SetEdgeStreams' "wire before launch" ordering) — read-only afterward, so a node's
	// own Update-loop closures (builders.go's injectClosures) can look it up by name.
	// nil map entries / a nil map itself (no WIREFOLD_STREAM_FDS "interior"
	// entry) mean the interior frame is simply never written for that node — tr.NodeBead
	// is a cheap no-op on the live path (neither its sink nor its onEvent hook is wired
	// in production).
	interiorOuts map[string]io.Writer
	// buildInteriorFrame packs one node's fixed-slot interior frame bytes
	// (Buffer.BuildInteriorStreamFrame), injected here (rather than importing Buffer) so
	// this package stays Buffer-independent, matching portRowFor/buildFrame's existing
	// interface-injection pattern on edgeMover.
	buildInteriorFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte
	// --- the dedicated VIEW stream (memory/feedback_no_single_writer_bridge.md,
	// memory/feedback_no_single_writer_bridge.md Step C) --- see view_stream.go.
	//
	// viewOut, when non-nil, is the VIEW stream's OWN dedicated fd (see SetViewStream /
	// Buffer/stream_fds.go's StreamKindView). Nil (the default — no WIREFOLD_STREAM_FDS
	// "view" entry, e.g. headless tests) means emitViewFrame is a no-op: nothing here
	// ever writes, and camera/overlay/scene are simply never emitted. Written ONLY by
	// the gesture/stdin-reader
	// goroutine (the sole caller of every MoveDispatch method that can change camera/
	// overlay/scene/selection/hover).
	viewOut io.Writer
	// viewBuildFrame packs this goroutine's own VIEW frame (Buffer.BuildViewStreamFrame),
	// injected via SetViewStream so this package stays Buffer-independent, mirroring
	// buildInteriorFrame/buildFrame's existing interface-injection pattern.
	viewBuildFrame ViewFrameBuilder
	// viewTick is a purely local frame-sequence counter for the VIEW stream (not shared
	// with any other stream's tick) — written only by the gesture/stdin-reader goroutine.
	viewTick uint32
}

// setEdgeStreams wires every edgeMover in edgeMovers to ITS OWN dedicated fd — the
// per-edge stream (memory/feedback_no_single_writer_bridge.md): fd = baseFd + row, where
// row is the STABLE edge-seed order (edgeSeeds, the same spec order the Edge block uses).
// portRowFor/buildFrame are injected funcs (not a Buffer import) so this package stays
// Buffer-independent. Edge selection is NOT injected: each edgeMover owns its OWN
// selected bit, set via a moveMsgKindSelect message on its extIn (md.sendEdgeSelect), not
// a lookup. A missing edgeMover for a seed row (should not happen) is skipped rather than
// panicking.
func (sw *streamWiring) setEdgeStreams(
	edgeSeeds []EdgeGeomSeed,
	edgeMovers map[string]*edgeMover,
	baseFd int,
	portRowFor func(node, port string, isInput bool) (int32, bool),
	nodeRowFor func(id string) (int32, bool),
	buildFrame func(tick uint32, srcPortRow, dstPortRow int32, selected uint8, label string, edgeLen float32, groupIdx int32, beadVal []int32, beadX, beadY, beadZ []float32, events []wire.RowEvent) []byte,
) {
	for row, seed := range edgeSeeds {
		em, ok := edgeMovers[seed.Label]
		if !ok {
			continue
		}
		fd := baseFd + row
		em.streamOut = os.NewFile(uintptr(fd), fmt.Sprintf("edge-fd%d", fd))
		em.edgeRow = int32(row)
		em.portRowFor = portRowFor
		em.nodeRowFor = nodeRowFor
		em.buildFrame = buildFrame
		// This edge now has a real consumer wired: let its PacedWire accumulate
		// pending events (nodes/wire/paced_wire.go's StreamsActive doc comment) —
		// set here, before this edge's mover goroutine launches, same "wire
		// before launch" ordering as everything else in this function.
		if em.dest != nil {
			em.dest.StreamsActive = true
		}
	}
}

// setNodeStreams wires every nodeMover in nodeMovers to ITS OWN dedicated node-fd
// (geometry+ports+label), AND wires sw.interiorOuts + sw.buildInteriorFrame — every
// node's own Update-loop closures (builders.go's injectClosures) look these up for its
// own dedicated interior-fd — the two emitting goroutines per node (memory/
// feedback_no_single_writer_bridge.md). nodeBase/interiorBase are the two fd ranges' base
// fds; row is the STABLE node-seed order (nodeSeeds, the same spec order the Node block
// uses). nodeRowFor/buildFrame/buildInteriorFrame are injected funcs (not
// a Buffer import), matching setEdgeStreams' existing pattern. Selection/hover/abc-drag/
// kind are NOT injected lookups: each nodeMover owns its OWN selected/hovered/
// latchedSel/gotDragMsg/dragDelta*/kindID fields, set via moveMsgKindSelect/Hover/
// Latched/AbcReset messages (or, for kindID, once here at construction — a node's kind
// never changes after load, so there is no lookup to perform on every emit). A missing
// nodeMover for a seed row (should not happen) is skipped rather than panicking.
func (sw *streamWiring) setNodeStreams(
	nodeSeeds []NodeGeomSeed,
	nodeMovers map[string]*nodeMover,
	nodeBase, interiorBase int,
	nodeRowFor func(id string) (int32, bool),
	buildFrame func(tick uint32, nodeRow int32, cx, cy, cz, radius, sphereR float32, vrx, vry, vrz, frx, fry, frz float32, selected, kindID, hovered, latchedSel, gotDragMsg uint8, dragDeltaA, dragDeltaB, dragDeltaC, dragRequantCount int32, gotForwardMsg uint8, forwardDeltaA, forwardDeltaB, forwardDeltaC, forwardFromRow int32, label string, portNames []string, portDX, portDY, portDZ, portPX, portPY, portPZ []float32, portIsInput, portHovered []uint8, dstNodeRows []int32, events []wire.RowEvent) []byte,
	buildInteriorFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte,
	kindIDFor func(kind string) uint8,
) {
	sw.interiorOuts = map[string]io.Writer{}
	sw.buildInteriorFrame = buildInteriorFrame
	for row, seed := range nodeSeeds {
		nm, ok := nodeMovers[seed.ID]
		if !ok {
			continue
		}
		nFd := nodeBase + row
		nm.streamOut = os.NewFile(uintptr(nFd), fmt.Sprintf("node-fd%d", nFd))
		nm.nodeRow = int32(row)
		// kindID is static per node (never changes after load) — resolved once here,
		// directly onto the mover's own field, not via a per-emit lookup func.
		if kindIDFor != nil {
			nm.kindID = kindIDFor(seed.Kind)
		}
		nm.nodeRowFor = nodeRowFor
		nm.buildFrame = buildFrame

		iFd := interiorBase + row
		sw.interiorOuts[seed.ID] = os.NewFile(uintptr(iFd), fmt.Sprintf("interior-fd%d", iFd))
	}
}
