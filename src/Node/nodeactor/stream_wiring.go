package nodeactor

import (
	"fmt"

	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/Node/Edge/edgegeom"
	"github.com/dtauraso/wirefold/src/Node/Edge/edgetable"
	interior "github.com/dtauraso/wirefold/src/Node/Interior"
	"github.com/dtauraso/wirefold/src/Node/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/src/Node/nodeactor/owners"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
)

type StreamWiring struct {
	interiorEmitters map[string]*interior.Emitter
}

func (sw *StreamWiring) InteriorEmittersPtr() *map[string]*interior.Emitter {
	return &sw.interiorEmitters
}

func kindOf(nodeGeoms map[string]*NodeGeometry, nodeID string) string {
	if nm, ok := nodeGeoms[nodeID]; ok {
		return nm.SelfKind()
	}
	return ""
}

func (sw *StreamWiring) SetEdgeStreams(
	edgeSeeds []edgegeom.Seed,
	edgeTable map[string]*edgetable.Edge,
	nodeGeoms map[string]*NodeGeometry,
	nodeRowFor func(id string) (int32, bool),
	buildFrame owners.EdgeFrameBuilder,
) {
	for row, seed := range edgeSeeds {
		em, ok := edgeTable[seed.Label]
		if !ok {
			continue
		}

		srcNM, ok := nodeGeoms[seed.SrcNode]
		if !ok {
			panic(fmt.Sprintf(
				"nodeactor.SetEdgeStreams: edge %q leaves node %q, which has no node geometry — a node draws its OWN out-edges, so an edge with no source node has no writer",
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
		srcNM.WireOutEdgeStream(seed.Label, int32(row), seed.DstNode, kindOf(nodeGeoms, seed.DstNode), srcRow, dstRow, buildFrame)

		if dest := em.Dest(); dest != nil {
			dest.SetStreamsActive(true)
		}
	}
}

func (sw *StreamWiring) SetNodeStreams(
	nodeSeeds []nodegeom.Seed,
	nodeMovers map[string]*NodeGeometry,
	sceneRoot string,
	buildBeadFrame beadanimation.BeadFrameBuilder,
	nodeRowFor func(id string) (int32, bool),
	buildFrame nodeframe.NodeFrameBuilder,
	kindIDFor func(kind string) uint8,
) {
	sw.interiorEmitters = map[string]*interior.Emitter{}

	for _, seed := range nodeSeeds {
		nm, ok := nodeMovers[seed.ID]
		if !ok {
			continue
		}
		row := seed.Row

		var kindID uint8
		if kindIDFor != nil {
			kindID = kindIDFor(seed.Kind)
		}
		nm.WireStream(int32(row), kindID, nodeRowFor, buildFrame, sceneRoot)

		nm.WireBeadStream(int32(row), buildBeadFrame, sceneRoot)

		sw.interiorEmitters[seed.ID] = nm.WireInteriorStream(int32(row), nil, sceneRoot)
	}
}
