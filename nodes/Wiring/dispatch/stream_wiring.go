// stream_wiring.go — the per-node interior-stream fd-wiring owner split out of
// MoveDispatch (god-object decomposition), as a pure move (no logic changes): streamWiring
// owns interiorOuts/buildInteriorFrame (the per-node interior fd directory) and driveOuts.
// MoveDispatch's public SetEdgeStreams/SetNodeStreams methods stay as thin delegators so
// the external API is unchanged. The dedicated VIEW stream's own fd/frame-builder/tick used
// to live here too (viewOut/viewBuildFrame/viewTick) — lifted onto
// nodes/Wiring/viewstate.UIState per docs/planning/gesture-actor.md, since its one caller
// (emitViewFrame) needs md.UI.vp/ov/sceneSphere just as much as it needs these fields, and
// both now live together in the same package.

package dispatch

import (
	"fmt"
	"io"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	geomseeds "github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	wire "github.com/dtauraso/wirefold/nodes/wire"
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
	// docs/investigations/interior-stream-framing.md). Populated ONCE by setNodeStreams alongside
	// interiorOuts, BEFORE any node's Update/DriveHeld goroutines launch. A nil slot
	// entry (no WIREFOLD_STREAM_FDS "drive" entry, or a kind that doesn't use that slot)
	// means writes through it are simply never made — nil-safe, same fallback shape as
	// interiorOuts.
	driveOuts map[string][driveSlotsPerNode]io.Writer

	// nodeClaims is the node-stream claim registry, using package nodeactor's own
	// ClaimRegistry type (nodeactor/stream_claim.go) — nodeactor cannot import package
	// Wiring's unexported claimedStream/streamClaims types (the type left this package
	// in docs/planning/movedispatch-decomposition.md §20), so it keeps a small duplicate
	// of the same mechanism, the same precedent §17 set for nodes/Wiring/edgemover's own
	// stream_claim.go. Lazily allocated by setNodeStreams. The VIEW stream's own claim
	// registry is separate (viewstate's own viewClaimedStream), and the EDGE stream's
	// own claim registry is separate too (edgeClaims below, package edgemover's own
	// ClaimRegistry). The three can never collide (disjoint namespaces: node ids, edge
	// labels, and the VIEW stream's own singleton), so splitting them cost nothing.
	nodeClaims nodeactor.ClaimRegistry
	// edgeClaims is the edge-stream counterpart of nodeClaims above, using package
	// edgemover's own ClaimRegistry type. Lazily allocated by setEdgeStreams.
	edgeClaims edgemover.ClaimRegistry
}

// ensureNodeClaims lazily allocates sw.nodeClaims on first use — see its doc comment.
func (sw *streamWiring) ensureNodeClaims() nodeactor.ClaimRegistry {
	if sw.nodeClaims == nil {
		sw.nodeClaims = nodeactor.NewClaimRegistry()
	}
	return sw.nodeClaims
}

// ensureEdgeClaims lazily allocates sw.edgeClaims on first use — see its doc comment.
func (sw *streamWiring) ensureEdgeClaims() edgemover.ClaimRegistry {
	if sw.edgeClaims == nil {
		sw.edgeClaims = edgemover.NewClaimRegistry()
	}
	return sw.edgeClaims
}

// setEdgeStreams wires every edgeMover in edgeMovers to ITS OWN dedicated fd — the
// per-edge stream (memory/feedback_no_single_writer_bridge.md): fd = baseFd + row, where
// row is the STABLE edge-seed order (edgeSeeds, the same spec order the Edge block uses).
// buildFrame is an injected func (not a Buffer import) so this package stays
// Buffer-independent. Edge selection is NOT injected: each edgeMover owns its OWN
// selected bit, set via a movemsg.KindSelect message on its extIn (md.sendEdgeSelect), not
// a lookup. A missing edgeMover for a seed row (should not happen) is skipped rather than
// panicking.
func (sw *streamWiring) setEdgeStreams(
	edgeSeeds []geomseeds.EdgeGeomSeed,
	edgeMovers map[string]*edgemover.EdgeMover,
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
		handle := edgemover.Claim(sw.ensureEdgeClaims(), seed.Label, rawOut)
		em.SetStream(handle, int32(row), nodeRowFor, buildFrame)
		// This edge now has a real consumer wired: let its PacedWire accumulate
		// pending events (nodes/wire/wire_readout.go's StreamsActive doc comment) —
		// set here, before this edge's mover goroutine launches, same "wire
		// before launch" ordering as everything else in this function.
		if dest := em.Dest(); dest != nil {
			dest.SetStreamsActive(true)
		}
	}
}

// setNodeStreams wires every nodeMover in nodeMovers to ITS OWN dedicated node-fd
// (geometry+ports+label), AND wires sw.interiorOuts + sw.buildInteriorFrame — every
// node's own Update-loop closures (builders.go's injectClosures) look these up for its
// own dedicated interior-fd — the two emitting goroutines per node (memory/
// feedback_no_single_writer_bridge.md). It ALSO wires sw.driveOuts, one dedicated fd PER
// (node row, drive slot) for each gatecommon.DriveHeld goroutine that node spawns (a
// THIRD-and-beyond kind of emitting goroutine per node — docs/investigations/interior-stream-framing.md,
// Buffer.StreamKindDrive), when driveWired is true; driveBase is then the drive fd
// range's base fd. driveWired false leaves sw.driveOuts populated with nil-slot entries
// only (see the loop body) — main.go requires "drive" present exactly when "node"/
// "interior" are, so this parameter is a defense against a caller that doesn't. nodeBase/
// interiorBase are the other two fd ranges' base fds; row is the STABLE node-seed order
// (nodeSeeds, the same spec order the Node block uses). nodeRowFor/buildFrame/
// buildInteriorFrame are injected funcs (not a Buffer import), matching setEdgeStreams'
// existing pattern. Selection/hover/
// kind are NOT injected lookups: each nodeMover owns its OWN selected/hovered/
// latchedSel/kindID fields, set via movemsg.KindSelect/Hover/Latched messages (or, for
// kindID, once here at construction — a node's kind never changes after load, so there
// is no lookup to perform on every emit). A missing nodeMover for a seed row (should not
// happen) is skipped rather than panicking.
func (sw *streamWiring) setNodeStreams(
	nodeSeeds []geomseeds.NodeGeomSeed,
	nodeMovers map[string]*nodeactor.NodeGeometry,
	nodeBase, interiorBase int,
	driveBase int, driveWired bool,
	nodeRowFor func(id string) (int32, bool),
	buildFrame nodeactor.NodeFrameBuilder,
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
		streamOut := nodeactor.Claim(sw.ensureNodeClaims(), seed.ID, rawNodeOut)
		// kindID is static per node (never changes after load) — resolved once here,
		// directly onto the mover's own field, not via a per-emit lookup func.
		var kindID uint8
		if kindIDFor != nil {
			kindID = kindIDFor(seed.Kind)
		}
		nm.WireStream(streamOut, int32(row), kindID, nodeRowFor, buildFrame)

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
