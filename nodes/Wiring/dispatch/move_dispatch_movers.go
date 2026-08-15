package dispatch

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgetable"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	geomseeds "github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/clock"
)

func (md *MoveDispatch) buildNodeMovers(geoms map[string]nodegeom.NodeGeom, tr *T.Trace, clk clock.Clock) {
	for id, g := range geoms {
		ng := nodeactor.NewNodeGeometry(id, g, tr, clk)

		selfID := id
		resolveDest := func(destID string) (func(movemsg.Msg) bool, bool) {

			if other, ok := md.MR.NodeGeoms()[destID]; ok {
				return other.NeighborTrySend(selfID)
			}
			return nil, false
		}
		ownGeom := ng
		commitLocal := func(_ string, newPos vec3, targetPolar *polar.Polar) {
			md.LQ.CommitNodeMoveLocal(md.MR.NodeGeoms(), md.MR.Edges(), &md.UI, ownGeom, newPos, targetPolar)
		}
		ng.WireMessaging(resolveDest, md.MR.EnqueueFor(ng), commitLocal)
		md.MR.NodeGeoms()[id] = ng

		md.MR.SeedCenter(id, nodegeom.NodeWorldPos(g))
	}
}

func (md *MoveDispatch) wireRuleMesh() {
	for id := range md.MR.NodeGeoms() {
		md.Rules.Claim(id)
	}
	rules := md.Rules.All()
	for id, rn := range rules {
		for peerID, peer := range rules {
			if peerID == id {
				continue
			}

			rn.LinkRuleDown(peerID, peer.RuleBackChannel(id))
		}
		md.MR.NodeGeoms()[id].LinkRuleState(rn.Out(), rn.Wake())
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
