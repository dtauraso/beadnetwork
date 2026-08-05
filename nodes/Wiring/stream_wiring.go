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
// driveSlotsPerNode is a local copy of Buffer.DriveSlotsPerNode's value (2), kept here
// rather than importing Buffer — same precedent as port_wiring.go's bufInteriorSlotsPerNode
// (this package stays Buffer-independent; buildFrame/buildInteriorFrame are injected
// funcs for the same reason).
const driveSlotsPerNode = 2

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
	// driveOuts holds DriveSlotsPerNode dedicated per-node DRIVE fds, keyed by node id —
	// one PER gatecommon.DriveHeld goroutine that node spawns (Buffer.StreamKindDrive;
	// docs/interior-stream-framing.md). Populated ONCE by setNodeStreams alongside
	// interiorOuts, BEFORE any node's Update/DriveHeld goroutines launch. A nil slot
	// entry (no WIREFOLD_STREAM_FDS "drive" entry, or a kind that doesn't use that slot)
	// means writes through it are simply never made — nil-safe, same fallback shape as
	// interiorOuts.
	driveOuts map[string][driveSlotsPerNode]io.Writer
	// --- the dedicated VIEW stream (memory/feedback_no_single_writer_bridge.md,
	// memory/feedback_no_single_writer_bridge.md Step C) --- see view_stream.go.
	//
	// viewOut, when Ok(), is the VIEW stream's OWN dedicated fd (see SetViewStream /
	// Buffer/stream_fds.go's StreamKindView). A dead claimedStream (the default — no
	// WIREFOLD_STREAM_FDS "view" entry, e.g. headless tests, OR a rejected second
	// SetViewStream claim — see stream_claim.go) means emitViewFrame is a no-op: nothing
	// here ever writes, and camera/overlay/scene are simply never emitted. Written ONLY
	// by the gesture/stdin-reader goroutine (the sole caller of every MoveDispatch
	// method that can change camera/overlay/scene/selection/hover).
	viewOut claimedStream
	// viewBuildFrame packs this goroutine's own VIEW frame (Buffer.BuildViewStreamFrame),
	// injected via SetViewStream so this package stays Buffer-independent, mirroring
	// buildInteriorFrame/buildFrame's existing interface-injection pattern.
	viewBuildFrame ViewFrameBuilder
	// viewTick is a purely local frame-sequence counter for the VIEW stream (not shared
	// with any other stream's tick) — written only by the gesture/stdin-reader goroutine.
	viewTick uint32

	// claims is the wiring-time claim registry backing every claimedStream this struct
	// hands out (streamOut on each nodeMover/edgeMover, viewOut here) — see
	// stream_claim.go's header comment. Lazily allocated (newStreamClaims) by whichever
	// of setEdgeStreams/setNodeStreams/SetViewStream runs first, so a bare streamWiring
	// zero value (test construction that never wires any stream) never allocates it.
	claims streamClaims
}

// ensureClaims lazily allocates sw.claims on first use — see its doc comment.
func (sw *streamWiring) ensureClaims() streamClaims {
	if sw.claims == nil {
		sw.claims = newStreamClaims()
	}
	return sw.claims
}

// setEdgeStreams wires every edgeMover in edgeMovers to ITS OWN dedicated fd — the
// per-edge stream (memory/feedback_no_single_writer_bridge.md): fd = baseFd + row, where
// row is the STABLE edge-seed order (edgeSeeds, the same spec order the Edge block uses).
// buildFrame is an injected func (not a Buffer import) so this package stays
// Buffer-independent. Edge selection is NOT injected: each edgeMover owns its OWN
// selected bit, set via a moveMsgKindSelect message on its extIn (md.sendEdgeSelect), not
// a lookup. A missing edgeMover for a seed row (should not happen) is skipped rather than
// panicking.
func (sw *streamWiring) setEdgeStreams(
	edgeSeeds []EdgeGeomSeed,
	edgeMovers map[string]*edgeMover,
	baseFd int,
	nodeRowFor func(id string) (int32, bool),
	buildFrame func(tick uint32, sx, sy, sz, ex, ey, ez float32, selected uint8, label string, events []wire.RowEvent) []byte,
) {
	for row, seed := range edgeSeeds {
		em, ok := edgeMovers[seed.Label]
		if !ok {
			continue
		}
		fd := baseFd + row
		rawOut := os.NewFile(uintptr(fd), fmt.Sprintf("edge-fd%d", fd))
		em.streamOut = newClaimedStream(sw.ensureClaims(), "edge", seed.Label, rawOut)
		em.edgeRow = int32(row)
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
// feedback_no_single_writer_bridge.md). It ALSO wires sw.driveOuts, one dedicated fd PER
// (node row, drive slot) for each gatecommon.DriveHeld goroutine that node spawns (a
// THIRD-and-beyond kind of emitting goroutine per node — docs/interior-stream-framing.md,
// Buffer.StreamKindDrive), when driveWired is true; driveBase is then the drive fd
// range's base fd. driveWired false leaves sw.driveOuts populated with nil-slot entries
// only (see the loop body) — main.go requires "drive" present exactly when "node"/
// "interior" are, so this parameter is a defense against a caller that doesn't. nodeBase/
// interiorBase are the other two fd ranges' base fds; row is the STABLE node-seed order
// (nodeSeeds, the same spec order the Node block uses). nodeRowFor/buildFrame/
// buildInteriorFrame are injected funcs (not a Buffer import), matching setEdgeStreams'
// existing pattern. Selection/hover/
// kind are NOT injected lookups: each nodeMover owns its OWN selected/hovered/
// latchedSel/kindID fields, set via moveMsgKindSelect/Hover/Latched messages (or, for
// kindID, once here at construction — a node's kind never changes after load, so there
// is no lookup to perform on every emit). A missing nodeMover for a seed row (should not
// happen) is skipped rather than panicking.
func (sw *streamWiring) setNodeStreams(
	nodeSeeds []NodeGeomSeed,
	nodeMovers map[string]*nodeMover,
	nodeBase, interiorBase int,
	driveBase int, driveWired bool,
	nodeRowFor func(id string) (int32, bool),
	buildFrame func(tick uint32, nodeRow int32, nodeID int32, cx, cy, cz, radius, sphereR float32, vrx, vry, vrz, frx, fry, frz float32, poleTheta, polePhi, ringAxisTheta, ringAxisPhi float32, selected, kindID, hovered, latchedSel uint8, label string, dstNodeRows []int32, chainBeadOX, chainBeadOY, chainBeadOZ []float32, chainBeadLit []uint8, chainBeadLitValue []int32, events []wire.RowEvent) []byte,
	buildInteriorFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte,
	kindIDFor func(kind string) uint8,
) {
	sw.interiorOuts = map[string]io.Writer{}
	sw.buildInteriorFrame = buildInteriorFrame
	// driveOuts is populated regardless of driveBase (0 when WIREFOLD_STREAM_FDS carries
	// no "drive" entry): its slots then simply stay nil (io.Writer zero value), matching
	// interiorOuts' own "populate the map, leave entries nil when the fd is absent"
	// shape rather than leaving the whole map nil.
	sw.driveOuts = map[string][driveSlotsPerNode]io.Writer{}
	for _, seed := range nodeSeeds {
		nm, ok := nodeMovers[seed.ID]
		if !ok {
			continue
		}
		row := seed.Row
		nFd := nodeBase + row
		rawNodeOut := os.NewFile(uintptr(nFd), fmt.Sprintf("node-fd%d", nFd))
		nm.streamOut = newClaimedStream(sw.ensureClaims(), "node", seed.ID, rawNodeOut)
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

		// One dedicated DRIVE fd per (node row, slot) — see Buffer.StreamKindDrive's
		// doc comment for why this exists and driveBase's absence handling (driveBase==0
		// only when the caller passed no "drive" WIREFOLD_STREAM_FDS entry; main.go
		// requires "node"/"interior"/"drive" present together, so in production either
		// all three resolve or none do).
		if driveWired {
			var slots [driveSlotsPerNode]io.Writer
			for slot := 0; slot < driveSlotsPerNode; slot++ {
				dFd := driveBase + row*driveSlotsPerNode + slot
				slots[slot] = os.NewFile(uintptr(dFd), fmt.Sprintf("drive-fd%d", dFd))
			}
			sw.driveOuts[seed.ID] = slots
		}
	}
}
