package dispatch

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
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
			if em, ok := md.MR.EdgeMovers()[destID]; ok {
				switch selfID {
				case em.SrcID():
					return em.TrySendFromSrc, true
				case em.DstID():
					return em.TrySendFromDst, true
				}
				return nil, false
			}
			if other, ok := md.MR.NodeGeoms()[destID]; ok {
				return other.NeighborTrySend(selfID)
			}
			return nil, false
		}
		ownGeom := ng
		commitLocal := func(_ string, newPos vec3) {
			md.LQ.CommitNodeMoveLocal(md.MR.NodeGeoms(), md.MR.EdgeMovers(), &md.UI, ownGeom, newPos)
		}
		ng.WireMessaging(resolveDest, md.MR.EnqueueFor(ng), md.MR.CenterOfNode, commitLocal)
		md.MR.NodeGeoms()[id] = ng

		md.MR.SeedCenter(id, nodegeom.NodeWorldPos(g))
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

func (md *MoveDispatch) buildEdgeMovers(edgeEndpoints map[string]inputcodec.EdgeEndpoints, geoms map[string]nodegeom.NodeGeom, tr *T.Trace, clk clock.Clock, speedSinks *[]chan float64) {
	for edgeID, ep := range edgeEndpoints {
		em := edgemover.New(edgeID, ep.Source, ep.Target, ep.SourceHandle, ep.TargetHandle, geoms[ep.Source], geoms[ep.Target], tr, clk)
		if speedSinks != nil {
			edgeSpeedCh := make(chan float64, 1)
			em.SetSpeedCh(edgeSpeedCh)
			*speedSinks = append(*speedSinks, edgeSpeedCh)
		}
		md.MR.EdgeMovers()[edgeID] = em

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
		for edgeID, em := range md.MR.EdgeMovers() {
			if em.SrcID() == id || em.DstID() == id {
				nm.AddEdgeID(edgeID)
			}
		}
	}
}
