// move_dispatch_movers.go — the actor-construction phases NewMoveDispatch calls in order:
// build one nodeMover per node (wiring its messaging closures), copy the load-time
// mutual-pair fact onto each node, build one edgeMover per edge (wiring speedSinks and the
// dedicated neighbor channels), seed every nodeMover's own partnerCenters mirror, and give
// every nodeMover the ids of its own incident edges. All five live here because they are
// the single pass, in order, that turns load-time geometry into live movers wired to each
// other.

package dispatch

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	geomseeds "github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// buildNodeMovers creates one nodeMover per node, wires its messaging closures, and seeds
// the dispatch goroutine's center mirror from the load-time geom.
func (md *MoveDispatch) buildNodeMovers(geoms map[string]nodegeom.NodeGeom, tr *T.Trace, clk clock.Clock) {
	for id, g := range geoms {
		ng := nodeactor.NewNodeGeometry(id, g, tr, clk)
		// resolveDest resolves the ONE dedicated directed channel FROM this node
		// (selfID, captured below) TO destID: another node's own neighborIn[selfID]
		// slot, or an incident edge's srcIn/dstIn depending on which endpoint this
		// node is. There is no shared dispatch map to look up — md.MR.NodeGeoms()/md.MR.EdgeMovers() are the
		// read-only directories, safe to read from any goroutine once construction
		// finishes.
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
		// Seed the dispatch goroutine's center mirror from the same load-time geom
		// (single-threaded setup, before md.Start — no driving goroutine is running yet)
		// so the first framing read has every center before any push arrives (mirrors
		// the partnerCenters seed below).
		md.MR.SeedCenter(id, nodegeom.NodeWorldPos(g))
	}
}

// wireMutualPairs copies each mutual-pair fact (a pair that points an edge at each other) —
// a LOAD-TIME fact of the edge set, resolved here (single-threaded setup, before any mover
// goroutine exists) — into each node's own field. Without it a node cannot know whether its
// target also aims back — its own outTargets say nothing about the other node's — and the
// two chains would draw along the identical centre line (nodegeom.ParallelChainOffset,
// nodegeom/port_geometry.go).
func (md *MoveDispatch) wireMutualPairs(edgeEndpoints map[string]inputcodec.EdgeEndpoints) {
	for src, targets := range geomseeds.MutualPairs(edgeEndpoints) {
		if nm, ok := md.MR.NodeGeoms()[src]; ok {
			for target := range targets {
				nm.AddMutualTarget(target)
			}
		}
	}
}

// buildEdgeMovers creates one edgeMover per edge, wires speedSinks (if non-nil, the
// loader's build-wide accumulator — every clock-owning goroutine must not be left behind),
// and ensures the dedicated node-to-node neighbor channels each edge's two endpoints share.
func (md *MoveDispatch) buildEdgeMovers(edgeEndpoints map[string]inputcodec.EdgeEndpoints, geoms map[string]nodegeom.NodeGeom, tr *T.Trace, clk clock.Clock, speedSinks *[]chan float64) {
	for edgeID, ep := range edgeEndpoints {
		em := edgemover.New(edgeID, ep.Source, ep.Target, ep.SourceHandle, ep.TargetHandle, geoms[ep.Source], geoms[ep.Target], tr, clk)
		if speedSinks != nil {
			edgeSpeedCh := make(chan float64, 1)
			em.SetSpeedCh(edgeSpeedCh)
			*speedSinks = append(*speedSinks, edgeSpeedCh)
		}
		md.MR.EdgeMovers()[edgeID] = em
		// This edge's two nodes each get a dedicated channel TO this edge (already
		// created above, srcIn/dstIn) — and each other's own dedicated channel for
		// node-to-node traffic (neighborIn, the plain-neighbor/partner-reemit fan):
		// two directed channels per ordered pair, never a shared inbox.
		if srcNM, ok := md.MR.NodeGeoms()[ep.Source]; ok {
			if dstNM, ok := md.MR.NodeGeoms()[ep.Target]; ok {
				dstNM.EnsureNeighborChannel(ep.Source)
				srcNM.EnsureNeighborChannel(ep.Target)
			}
		}
	}
}

// seedPartnerCenters seeds every nodeMover's own partnerCenters map: quantized_move.go's
// neighbor-move math (neighborSetCReposition et al.) reads a direct neighbor's CURRENT
// world center off THIS node's OWN partnerCenters map (owned, written only by this node's
// own goroutine), kept current thereafter by each neighbor's own movemsg.KindNeighborCenter
// push (applyCenter) — one hop, no cascade.
func (md *MoveDispatch) seedPartnerCenters() {
	for _, nm := range md.MR.NodeGeoms() {
		// Seed partnerCenters at construction (single-threaded setup, before md.Start —
		// no mover goroutine is running yet, so reading a neighbor's geom directly here
		// is safe) with the SAME value the old snap seed used (newNodeMover seeds snap
		// from nodegeom.NodeWorldPos(geom)), so the first emit reproduces today's center exactly.
		// A node's neighbor set is nm.NeighborIDs() (populated above from
		// edgeEndpoints — one dedicated channel per adjacent node, both directions).
		for _, neighborID := range nm.NeighborIDs() {
			if other, ok := md.MR.NodeGeoms()[neighborID]; ok {
				nm.SeedPartnerCenter(neighborID, other.WorldCenter())
			}
		}
	}
}

// wireNodeEdgeIDs gives every nodeMover the ids of its OWN incident edges, so a
// lock-driven move can notify its edges via sendMove (resolveDest's per-pair channel
// lookup) — no cached channel slice.
func (md *MoveDispatch) wireNodeEdgeIDs() {
	for id, nm := range md.MR.NodeGeoms() {
		for edgeID, em := range md.MR.EdgeMovers() {
			if em.SrcID() == id || em.DstID() == id {
				nm.AddEdgeID(edgeID)
			}
		}
	}
}
