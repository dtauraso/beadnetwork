package scenerun

import (
	"fmt"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	edge "github.com/dtauraso/wirefold/Categories/Node/Edge"
	"github.com/dtauraso/wirefold/Categories/Node/Edge/edgegeom"
	"github.com/dtauraso/wirefold/Categories/Node/Edge/edgetable"
	interior "github.com/dtauraso/wirefold/Categories/Node/Interior"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/owners"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Node/rulenode"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (md *MoveDispatch) buildNodeMovers(geoms map[string]nodegeom.NodeGeom, clk clock.Clock, constants polarindex.SceneConstants) {
	for id, g := range geoms {
		ng := nodeactor.NewNodeGeometry(id, g, clk, constants)

		selfID := id
		resolveDest := func(destID string) (owners.Deposit, bool) {

			if other, ok := md.MR.NodeGeoms()[destID]; ok {
				return other.NeighborDeposit(selfID)
			}
			return nil, false
		}
		ownGeom := ng
		commitLocal := func(_ string, idx polarindex.Index) {
			md.Mover.CommitNodeMoveLocal(md.MR.NodeGeoms(), md.MR.Edges(), ownGeom, idx)
		}
		ng.WireMessaging(resolveDest, md.MR.EnqueueFor(ng), commitLocal)
		md.Inboxes.ClaimChannelVectorsIn(id, ng.ChannelVectorsIn())
		md.MR.NodeGeoms()[id] = ng

		md.MR.SeedCenter(id, Vec3(nodegeom.NodeWorldPos(g)))
	}
}

func (md *MoveDispatch) wireRuleMesh() {
	geoms := md.MR.NodeGeoms()
	for id, nm := range geoms {
		nm.AttachRuleNode(rulenode.New(id))
	}
	for id, nm := range geoms {
		for peerID, peer := range geoms {
			if peerID == id {
				continue
			}

			nm.RuleNode().LinkRuleDown(peerID, peer.RuleBackChannel(id))
		}
	}

}

func (md *MoveDispatch) wireRuleEditRows() {
	geoms := md.MR.NodeGeoms()
	md.Rules.SizeByNodeRows(len(md.RT.NodeRowTable))
	for row, id := range md.RT.NodeRowTable {
		nm, ok := geoms[id]
		if !ok {
			continue
		}
		md.Rules.EditsByNodeRow[row] = nm.RuleNode().Edits()
		md.Rules.KindTogglesByNodeRow[row] = nm.RuleNode().KindToggleChannel()
	}
}

func (md *MoveDispatch) wireMutualPairs(edgeEndpoints map[string]edge.EdgeEndpoints) {
	pairs := make([][2]string, 0, len(edgeEndpoints))
	for _, ep := range edgeEndpoints {
		pairs = append(pairs, [2]string{ep.Source, ep.Target})
	}
	for src, targets := range edgegeom.MutualPairs(pairs) {
		if nm, ok := md.MR.NodeGeoms()[src]; ok {
			for target := range targets {
				nm.Topo().AddMutualTarget(target)
			}
		}
	}
}

func (md *MoveDispatch) buildEdgeTable(edgeEndpoints map[string]edge.EdgeEndpoints) {
	for edgeID, ep := range edgeEndpoints {
		md.MR.Edges()[edgeID] = edgetable.New(edgeID, ep.Source, ep.Target, ep.SourceHandle, ep.TargetHandle)

		if srcNM, ok := md.MR.NodeGeoms()[ep.Source]; ok {
			if dstNM, ok := md.MR.NodeGeoms()[ep.Target]; ok {
				dstNM.EnsureNeighborChannel(ep.Source)
				srcNM.EnsureNeighborChannel(ep.Target)
			}
		}
	}
}

func (md *MoveDispatch) wireNodeEdgeIDs() {
	for id, nm := range md.MR.NodeGeoms() {
		for edgeID, e := range md.MR.Edges() {
			if e.SrcID() == id || e.DstID() == id {
				nm.Topo().AddEdgeID(edgeID)
			}
		}
	}
}

type StreamWiring struct {
	interiorEmitters map[string]*interior.Emitter
}

func (sw *StreamWiring) InteriorEmittersPtr() *map[string]*interior.Emitter {
	return &sw.interiorEmitters
}

func kindOf(nodeGeoms map[string]*nodeactor.NodeGeometry, nodeID string) string {
	if nm, ok := nodeGeoms[nodeID]; ok {
		return nm.SelfKind()
	}
	return ""
}

func (sw *StreamWiring) SetEdgeStreams(
	edgeSeeds []edgegeom.Seed,
	edgeTable map[string]*edgetable.Edge,
	nodeGeoms map[string]*nodeactor.NodeGeometry,
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
	nodeMovers map[string]*nodeactor.NodeGeometry,
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
