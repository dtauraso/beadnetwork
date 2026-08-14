package streamwire

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgetable"
	geomseeds "github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/streamclaim"
	"github.com/dtauraso/wirefold/nodes/rowevent"
)

type StreamWiring struct {
	interiorEmitters map[string]*interior.Emitter

	buildInteriorFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) []byte

	nodeClaims streamclaim.ClaimRegistry
}

func (sw *StreamWiring) InteriorEmittersPtr() *map[string]*interior.Emitter {
	return &sw.interiorEmitters
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

func kindOf(nodeGeoms map[string]*nodeactor.NodeGeometry, nodeID string) string {
	if nm, ok := nodeGeoms[nodeID]; ok {
		return nm.SelfKind()
	}
	return ""
}

func (sw *StreamWiring) SetEdgeStreams(
	edgeSeeds []geomseeds.EdgeGeomSeed,
	edgeTable map[string]*edgetable.Edge,
	nodeGeoms map[string]*nodeactor.NodeGeometry,
	baseFd int,
	nodeRowFor func(id string) (int32, bool),
	buildFrame func(tick uint32, sx, sy, sz, ex, ey, ez float32, srcNodeRow, dstNodeRow int32, deltaR float32, label string, events []rowevent.RowEvent) []byte,
) {
	for row, seed := range edgeSeeds {
		em, ok := edgeTable[seed.Label]
		if !ok {
			continue
		}
		fd := baseFd + row
		rawOut := os.NewFile(uintptr(fd), fmt.Sprintf("edge-fd%d", fd))

		srcNM, ok := nodeGeoms[seed.SrcNode]
		if !ok {
			panic(fmt.Sprintf(
				"streamwire.SetEdgeStreams: edge %q leaves node %q, which has no node geometry — a node draws its OWN out-edges, so an edge with no source node has no writer",
				seed.Label, seed.SrcNode))
		}
		srcRow := int32(-1)
		if r, ok := nodeRowFor(seed.SrcNode); ok {
			srcRow = r
		}
		dstRow := int32(-1)
		if r, ok := nodeRowFor(seed.DstNode); ok {
			dstRow = r
		}
		srcNM.WireOutEdgeStream(seed.Label, int32(row), seed.DstNode, kindOf(nodeGeoms, seed.DstNode), rawOut, srcRow, dstRow, buildFrame)

		if dest := em.Dest(); dest != nil {
			dest.SetStreamsActive(true)
		}
	}
}

func (sw *StreamWiring) SetNodeStreams(
	nodeSeeds []geomseeds.NodeGeomSeed,
	nodeMovers map[string]*nodeactor.NodeGeometry,
	nodeBase, interiorBase int,
	beadBase int, beadWired bool,
	buildBeadFrame owners.BeadFrameBuilder,
	nodeRowFor func(id string) (int32, bool),
	buildFrame nodeframe.NodeFrameBuilder,
	buildInteriorFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) []byte,
	kindIDFor func(kind string) uint8,
) {
	sw.interiorEmitters = map[string]*interior.Emitter{}
	sw.buildInteriorFrame = buildInteriorFrame

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
		rawInteriorOut := os.NewFile(uintptr(iFd), fmt.Sprintf("interior-fd%d", iFd))
		sw.interiorEmitters[seed.ID] = nm.WireInteriorStream(rawInteriorOut, int32(row), buildInteriorFrame)
	}
}
