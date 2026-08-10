// move_dispatch_construct.go — builds a MoveDispatch from load-time geometry: one
// nodeMover per node, one edgeMover per edge, and the dedicated directed channels wiring
// adjacent movers together (see move_dispatch.go's doc comment for the model this
// reproduces per-goroutine). This is the one-time, single-threaded setup step; the
// MoveDispatch struct itself lives in move_dispatch.go and its public delegator API lives
// in move_dispatch_api.go.

package Wiring

import (
	"fmt"
	"sort"
	"strconv"

	geomseeds "github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	rowtables "github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

// newMoveDispatch builds the registry from per-node geometry and per-edge endpoints.
// It creates one nodeMover per node and one edgeMover per edge, registering each under
// its key (node id / edge id) in md.mr.nodeGeoms/md.mr.edgeMovers, and wires the dedicated
// directed channels between adjacent movers. Outs and dest wires are bound later by Bind once node
// construction has populated them. nodeOrder/edgeOrder are the
// SPEC order (deterministic directory-sorted order, not map iteration order) used to
// build md.GS.NodeSeeds/EdgeSeeds for buffer row seeding.
//
// speedSinks, when non-nil, is the loader's build-wide accumulator
// (buildCtx.speedSinks): each nodeMover AND each edgeMover created below gets its own
// fresh buffered-1 speed channel (per-goroutine-clock.md "Delivery" — every
// clock-owning goroutine must not be left behind), and that channel's SEND end is
// appended here.
// nil in test call sites that construct a MoveDispatch directly with no
// loader — those edgeMovers then simply have no speed channel to poll.
// rowCount is the buffer's node-row space (topoSpec.RowCount — the largest node id found,
// not the node count): rows 0..rowCount-1, ROW ID = NODE ID - 1. 0 (test call sites that
// don't pass one) falls back to the number of resolved seeds, i.e. no gaps.
// NewMoveDispatchForTest exposes newMoveDispatch to an EXTERNAL test package (one that
// cannot import this package's own _test.go helpers) that needs a real, bare MoveDispatch
// without going through LoadTopology (which requires the full kind registry populated by
// kinds_generated.go's blank imports — package main only). Same escape-hatch shape as
// NewDrivenOutForTest (driven_out.go): a clearly-named, test-only constructor, no
// production caller.
func NewMoveDispatchForTest(geoms map[string]nodegeom.NodeGeom, edgeEndpoints map[string]inputcodec.EdgeEndpoints, tr *T.Trace, nodeOrder, edgeOrder []string, clk clock.Clock, speedSinks *[]chan float64, rowCount int) (*MoveDispatch, error) {
	return newMoveDispatch(geoms, edgeEndpoints, tr, nodeOrder, edgeOrder, clk, speedSinks, rowCount)
}

func newMoveDispatch(geoms map[string]nodegeom.NodeGeom, edgeEndpoints map[string]inputcodec.EdgeEndpoints, tr *T.Trace, nodeOrder, edgeOrder []string, clk clock.Clock, speedSinks *[]chan float64, rowCount int) (*MoveDispatch, error) {
	// nil order (test call sites that don't care about seed order) falls back to sorted
	// map keys — still deterministic, just not necessarily spec order.
	if nodeOrder == nil {
		nodeOrder = make([]string, 0, len(geoms))
		for id := range geoms {
			nodeOrder = append(nodeOrder, id)
		}
		sort.Strings(nodeOrder)
	}
	if edgeOrder == nil {
		edgeOrder = make([]string, 0, len(edgeEndpoints))
		for label := range edgeEndpoints {
			edgeOrder = append(edgeOrder, label)
		}
		sort.Strings(edgeOrder)
	}
	md := &MoveDispatch{
		tr: tr,
	}
	md.mr.nodeGeoms = map[string]*nodeGeometry{}
	md.mr.edgeMovers = map[string]*edgeMover{}
	md.mr.edgeOut = map[string]*wire.Out{}
	md.mr.centerMirror = map[string]vec3{}
	md.ui.ov = defaultOverlayState()
	md.ui.speed = 1                            // default playback multiplier; LoadSpeed overwrites from view/speed.json if present
	md.ui.clockDivisor = 1                     // no scaling until LoadSpeed resolves the loaded scene's own divisor
	md.ui.latticePoints = defaultLatticePoints // LoadLatticePoints overwrites from view/lattice.json if present
	md.GS.NodeSeeds = make([]geomseeds.NodeGeomSeed, 0, len(nodeOrder))
	for i, id := range nodeOrder {
		g, ok := geoms[id]
		if !ok {
			continue
		}
		label := g.Label
		if label == "" {
			label = id
		}
		var cx, cy, cz float64
		if g.HasPos {
			c := nodegeom.NodeWorldPos(g)
			cx, cy, cz = c.X, c.Y, c.Z
		}
		// ROW ID = NODE ID - 1 — declared by the id, not by position in nodeOrder. Falls
		// back to positional index only for a non-numeric id, which real (loadTree-built)
		// specs never produce (loud load-time error there); this keeps synthetic-id unit
		// tests that construct a MoveDispatch directly working unchanged.
		row := i
		if n, err := strconv.Atoi(id); err == nil {
			row = n - 1
		}
		md.GS.NodeSeeds = append(md.GS.NodeSeeds, geomseeds.NodeGeomSeed{
			ID: id, Label: label, Kind: g.Kind,
			CX: cx, CY: cy, CZ: cz,
			Radius: nodegeom.NodeRadius(g.Kind), SphereR: nodegeom.EffectiveRadius(g),
			VRX: loadspec.VerticalRingNormalX, VRY: loadspec.VerticalRingNormalY, VRZ: loadspec.VerticalRingNormalZ,
			FRX: loadspec.FlatRingNormalX, FRY: loadspec.FlatRingNormalY, FRZ: loadspec.FlatRingNormalZ,
			Row: row,
		})
	}
	md.GS.EdgeSeeds = make([]geomseeds.EdgeGeomSeed, 0, len(edgeOrder))
	for _, label := range edgeOrder {
		ep, ok := edgeEndpoints[label]
		if !ok {
			continue
		}
		// Real endpoint geometry: the same nodegeom.EdgeSegment computation recomputeGeometry
		// (below) uses on every live move, evaluated once here against the load-time
		// geoms so the seed row is never a degenerate 0,0,0->0,0,0 segment. A missed
		// lookup means the edge's source or target node id has no geometry — a
		// malformed topology (most commonly a stale edge file left behind after its
		// target node's directory was deleted by hand; in-edges are not indexed, so
		// nothing else catches this) — and must fail the load loudly, never seed a
		// silent 0,0,0->0,0,0 segment indistinguishable from real data.
		srcG, srcOK := geoms[ep.Source]
		if !srcOK {
			return nil, fmt.Errorf("newMoveDispatch: edge %q references missing source node %q (no geometry loaded for it)", label, ep.Source)
		}
		dstG, dstOK := geoms[ep.Target]
		if !dstOK {
			return nil, fmt.Errorf("newMoveDispatch: edge %q references missing target node %q (no geometry loaded for it)", label, ep.Target)
		}
		seg := nodegeom.EdgeSegment(srcG, dstG)
		sx, sy, sz := seg.Start.X, seg.Start.Y, seg.Start.Z
		ex, ey, ez := seg.End.X, seg.End.Y, seg.End.Z
		md.GS.EdgeSeeds = append(md.GS.EdgeSeeds, geomseeds.EdgeGeomSeed{
			Label: label, SrcNode: ep.Source, DstNode: ep.Target,
			SX: sx, SY: sy, SZ: sz, EX: ex, EY: ey, EZ: ez,
		})
	}
	for id, g := range geoms {
		ng := newNodeGeometry(id, g, tr, clk)
		// resolveDest resolves the ONE dedicated directed channel FROM this node
		// (selfID, captured below) TO destID: another node's own neighborIn[selfID]
		// slot, or an incident edge's srcIn/dstIn depending on which endpoint this
		// node is. There is no shared dispatch map to look up — md.mr.nodeGeoms/md.mr.edgeMovers are the
		// read-only directories, safe to read from any goroutine once construction
		// finishes.
		selfID := id
		ng.msg.resolveDest = func(destID string) (chan movemsg.Msg, bool) {
			if em, ok := md.mr.edgeMovers[destID]; ok {
				switch selfID {
				case em.srcID:
					return em.srcIn, true
				case em.dstID:
					return em.dstIn, true
				}
				return nil, false
			}
			if other, ok := md.mr.nodeGeoms[destID]; ok {
				if ch, ok := other.msg.neighborIn[selfID]; ok {
					return ch, true
				}
			}
			return nil, false
		}
		ng.msg.sendMove = md.enqueueFor(ng)
		ng.msg.tap = md.tapToInstall
		ng.msg.centerOf = md.centerOfNode
		ownGeom := ng
		ng.msg.commitLocal = func(_ string, newPos vec3) { md.commitNodeMoveLocal(ownGeom, newPos) }
		md.mr.nodeGeoms[id] = ng
		// Seed the dispatch goroutine's center mirror from the same load-time geom
		// (single-threaded setup, before md.Start — no driving goroutine is running yet)
		// so the first framing read has every center before any push arrives (mirrors
		// the partnerCenters seed below).
		md.mr.centerMirror[id] = nodegeom.NodeWorldPos(g)
	}
	// A pair that points at each other is a LOAD-TIME fact of the edge set, resolved here
	// (single-threaded setup, before any mover goroutine exists) and copied into each
	// node's own field. Without it a node cannot know whether its target also aims back —
	// its own outTargets say nothing about the other node's — and the two chains would draw
	// along the identical centre line (nodegeom.ParallelChainOffset, nodegeom/port_geometry.go).
	hasEdge := make(map[string]bool, len(edgeEndpoints))
	for _, ep := range edgeEndpoints {
		hasEdge[ep.Source+"\x00"+ep.Target] = true
	}
	for _, ep := range edgeEndpoints {
		if !hasEdge[ep.Target+"\x00"+ep.Source] {
			continue
		}
		if nm, ok := md.mr.nodeGeoms[ep.Source]; ok {
			if nm.topo.mutualTargets == nil {
				nm.topo.mutualTargets = map[string]bool{}
			}
			nm.topo.mutualTargets[ep.Target] = true
		}
	}
	for edgeID, ep := range edgeEndpoints {
		em := newEdgeMover(ep, edgeID, geoms[ep.Source], geoms[ep.Target], tr, clk)
		if speedSinks != nil {
			edgeSpeedCh := make(chan float64, 1)
			em.speedCh = edgeSpeedCh
			*speedSinks = append(*speedSinks, edgeSpeedCh)
		}
		md.mr.edgeMovers[edgeID] = em
		// This edge's two nodes each get a dedicated channel TO this edge (already
		// created above, srcIn/dstIn) — and each other's own dedicated channel for
		// node-to-node traffic (neighborIn, the plain-neighbor/partner-reemit fan):
		// two directed channels per ordered pair, never a shared inbox.
		if srcNM, ok := md.mr.nodeGeoms[ep.Source]; ok {
			if dstNM, ok := md.mr.nodeGeoms[ep.Target]; ok {
				if _, exists := dstNM.msg.neighborIn[ep.Source]; !exists {
					dstNM.msg.neighborIn[ep.Source] = make(chan movemsg.Msg, moverInboxDepth)
				}
				if _, exists := srcNM.msg.neighborIn[ep.Target]; !exists {
					srcNM.msg.neighborIn[ep.Target] = make(chan movemsg.Msg, moverInboxDepth)
				}
			}
		}
	}
	// Seed every nodeMover's own partnerCenters map: quantized_move.go's neighbor-move
	// math (neighborSetCReposition et al.) reads a direct neighbor's CURRENT world center
	// off THIS node's OWN partnerCenters map (owned, written only by this node's own
	// goroutine), kept current thereafter by each neighbor's own movemsg.KindNeighborCenter
	// push (applyCenter) — one hop, no cascade.
	for _, nm := range md.mr.nodeGeoms {
		// Seed partnerCenters at construction (single-threaded setup, before md.Start —
		// no mover goroutine is running yet, so reading a neighbor's geom directly here
		// is safe) with the SAME value the old snap seed used (newNodeMover seeds snap
		// from nodegeom.NodeWorldPos(geom)), so the first emit reproduces today's center exactly.
		// A node's neighbor set is nm.msg.neighborIn's key set (populated above from
		// edgeEndpoints — one dedicated channel per adjacent node, both directions).
		for neighborID := range nm.msg.neighborIn {
			if other, ok := md.mr.nodeGeoms[neighborID]; ok {
				nm.topo.partnerCenters[neighborID] = nodegeom.NodeWorldPos(other.geom)
			}
		}
	}
	// Give every nodeMover the ids of its OWN incident edges, so a lock-driven move can
	// notify its edges via sendMove (resolveDest's per-pair channel lookup) — no cached
	// channel slice.
	for id, nm := range md.mr.nodeGeoms {
		for edgeID, em := range md.mr.edgeMovers {
			if em.srcID == id || em.dstID == id {
				nm.topo.edgeIDs = append(nm.topo.edgeIDs, edgeID)
			}
		}
	}
	// Row-identity tables: built ONCE here, from nodeSeeds/edgeSeeds (each node seed already
	// carries its own absolute Row = id-1) — see RowTables.Build's doc comment for why this is
	// a load-time constant, not a discovery log.
	rtNodeSeeds := make([]rowtables.NodeSeed, len(md.GS.NodeSeeds))
	for i, sd := range md.GS.NodeSeeds {
		rtNodeSeeds[i] = rowtables.NodeSeed{ID: sd.ID, Row: sd.Row}
	}
	rtEdgeSeeds := make([]rowtables.EdgeSeed, len(md.GS.EdgeSeeds))
	for i, sd := range md.GS.EdgeSeeds {
		rtEdgeSeeds[i] = rowtables.EdgeSeed{Label: sd.Label, SrcNode: sd.SrcNode, DstNode: sd.DstNode}
	}
	md.RT.Build(rtNodeSeeds, rtEdgeSeeds, rowCount)
	return md, nil
}
