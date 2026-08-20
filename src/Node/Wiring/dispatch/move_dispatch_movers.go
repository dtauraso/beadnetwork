package dispatch

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/edgetable"
	geomseeds "github.com/dtauraso/wirefold/src/Node/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/src/Node/Wiring/polarindex"
	"github.com/dtauraso/wirefold/src/Node/Wiring/rulenode"
	"github.com/dtauraso/wirefold/src/Clock"
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

		md.MR.SeedCenter(id, nodegeom.NodeWorldPos(g))
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

func (md *MoveDispatch) wireMutualPairs(edgeEndpoints map[string]inputcodec.EdgeEndpoints) {
	for src, targets := range geomseeds.MutualPairs(edgeEndpoints) {
		if nm, ok := md.MR.NodeGeoms()[src]; ok {
			for target := range targets {
				nm.AddMutualTarget(target)
			}
		}
	}
}

func (md *MoveDispatch) buildEdgeTable(edgeEndpoints map[string]inputcodec.EdgeEndpoints) {
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
				nm.AddEdgeID(edgeID)
			}
		}
	}
}
