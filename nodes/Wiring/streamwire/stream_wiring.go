package streamwire

import (
	"fmt"
	"io"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	geomseeds "github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/streamclaim"
	"github.com/dtauraso/wirefold/nodes/rowevent"
)

const DriveSlotsPerNode = 2

type StreamWiring struct {
	interiorOuts map[string]io.Writer

	buildInteriorFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) []byte

	driveOuts map[string][DriveSlotsPerNode]io.Writer

	nodeClaims streamclaim.ClaimRegistry

	edgeClaims edgemover.ClaimRegistry
}

func (sw *StreamWiring) InteriorOutsPtr() *map[string]io.Writer { return &sw.interiorOuts }

func (sw *StreamWiring) DriveOutsPtr() *map[string][DriveSlotsPerNode]io.Writer {
	return &sw.driveOuts
}

func (sw *StreamWiring) BuildInteriorFramePtr() *func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) []byte {
	return &sw.buildInteriorFrame
}

func (sw *StreamWiring) ensureNodeClaims() streamclaim.ClaimRegistry {
	if sw.nodeClaims == nil {
		sw.nodeClaims = streamclaim.NewClaimRegistry()
	}
	return sw.nodeClaims
}

func (sw *StreamWiring) ensureEdgeClaims() edgemover.ClaimRegistry {
	if sw.edgeClaims == nil {
		sw.edgeClaims = edgemover.NewClaimRegistry()
	}
	return sw.edgeClaims
}

func nodeDrawsOwnOutEdges(nodeID string) bool { return nodeID == "1" }

func kindOf(nodeGeoms map[string]*nodeactor.NodeGeometry, nodeID string) string {
	if nm, ok := nodeGeoms[nodeID]; ok {
		return nm.SelfKind()
	}
	return ""
}

func (sw *StreamWiring) SetEdgeStreams(
	edgeSeeds []geomseeds.EdgeGeomSeed,
	edgeMovers map[string]*edgemover.EdgeMover,
	nodeGeoms map[string]*nodeactor.NodeGeometry,
	baseFd int,
	nodeRowFor func(id string) (int32, bool),
	buildFrame func(tick uint32, sx, sy, sz, ex, ey, ez float32, srcNodeRow int32, label string, events []rowevent.RowEvent) []byte,
) {
	for row, seed := range edgeSeeds {
		em, ok := edgeMovers[seed.Label]
		if !ok {
			continue
		}
		fd := baseFd + row
		rawOut := os.NewFile(uintptr(fd), fmt.Sprintf("edge-fd%d", fd))

		if srcNM, sourceDraws := nodeGeoms[seed.SrcNode]; sourceDraws && nodeDrawsOwnOutEdges(seed.SrcNode) {
			srcRow := int32(-1)
			if r, ok := nodeRowFor(seed.SrcNode); ok {
				srcRow = r
			}
			srcNM.WireOutEdgeStream(seed.Label, int32(row), seed.DstNode, kindOf(nodeGeoms, seed.DstNode), rawOut, srcRow, buildFrame)
			if dest := em.Dest(); dest != nil {
				dest.SetStreamsActive(true)
			}
			continue
		}

		handle := edgemover.Claim(sw.ensureEdgeClaims(), seed.Label, rawOut)
		em.SetStream(handle, int32(row), nodeRowFor, buildFrame)

		if dest := em.Dest(); dest != nil {
			dest.SetStreamsActive(true)
		}
	}
}

func (sw *StreamWiring) SetNodeStreams(
	nodeSeeds []geomseeds.NodeGeomSeed,
	nodeMovers map[string]*nodeactor.NodeGeometry,
	nodeBase, interiorBase int,
	driveBase int, driveWired bool,
	beadBase int, beadWired bool,
	buildBeadFrame owners.BeadFrameBuilder,
	nodeRowFor func(id string) (int32, bool),
	buildFrame nodeframe.NodeFrameBuilder,
	buildInteriorFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) []byte,
	kindIDFor func(kind string) uint8,
) {
	sw.interiorOuts = map[string]io.Writer{}
	sw.buildInteriorFrame = buildInteriorFrame

	sw.driveOuts = map[string][DriveSlotsPerNode]io.Writer{}
	for _, seed := range nodeSeeds {
		nm, ok := nodeMovers[seed.ID]
		if !ok {
			continue
		}
		row := seed.Row
		nFd := nodeBase + row
		rawNodeOut := os.NewFile(uintptr(nFd), fmt.Sprintf("node-fd%d", nFd))
		streamOut := streamclaim.Claim(sw.ensureNodeClaims(), seed.ID, rawNodeOut)

		var kindID uint8
		if kindIDFor != nil {
			kindID = kindIDFor(seed.Kind)
		}
		nm.WireStream(streamOut, int32(row), kindID, nodeRowFor, buildFrame)

		if beadWired {
			bFd := beadBase + row
			nm.WireBeadStream(os.NewFile(uintptr(bFd), fmt.Sprintf("bead-fd%d", bFd)), int32(row), buildBeadFrame)
		}

		iFd := interiorBase + row
		sw.interiorOuts[seed.ID] = os.NewFile(uintptr(iFd), fmt.Sprintf("interior-fd%d", iFd))

		if driveWired {
			var slots [DriveSlotsPerNode]io.Writer
			for slot := 0; slot < DriveSlotsPerNode; slot++ {
				dFd := driveBase + row*DriveSlotsPerNode + slot
				slots[slot] = os.NewFile(uintptr(dFd), fmt.Sprintf("drive-fd%d", dFd))
			}
			sw.driveOuts[seed.ID] = slots
		}
	}
}
