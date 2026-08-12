// move_dispatch_seeds.go — the seed-resolution phases NewMoveDispatch calls in order:
// resolve the SPEC order the caller passed (or fall back to sorted map keys), build the
// buffer row seeds (md.GS.NodeSeeds/EdgeSeeds) from that order and the load-time geometry,
// and build the row-identity tables (md.RT) from those same seeds. All three live here
// because each reads/writes the same GS/RT state in the same pass over nodeOrder/edgeOrder.

package dispatch

import (
	"sort"
	"strconv"

	geomseeds "github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	rowtables "github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
)

// resolveSeedOrders returns nodeOrder/edgeOrder, falling back to sorted map keys when the
// caller passed nil (test call sites that don't care about seed order) — still
// deterministic, just not necessarily spec order.
func resolveSeedOrders(geoms map[string]nodegeom.NodeGeom, edgeEndpoints map[string]inputcodec.EdgeEndpoints, nodeOrder, edgeOrder []string) ([]string, []string) {
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
	return nodeOrder, edgeOrder
}

// buildGeomSeeds builds md.GS.NodeSeeds/EdgeSeeds — the buffer row seeds — from the
// load-time geoms/edgeEndpoints, in nodeOrder/edgeOrder (the SPEC order).
func (md *MoveDispatch) buildGeomSeeds(geoms map[string]nodegeom.NodeGeom, edgeEndpoints map[string]inputcodec.EdgeEndpoints, nodeOrder, edgeOrder []string) error {
	md.GS.NodeSeeds = make([]geomseeds.NodeGeomSeed, 0, len(nodeOrder))
	for i, id := range nodeOrder {
		g, ok := geoms[id]
		if !ok {
			continue
		}
		// ROW ID = NODE ID - 1 — declared by the id, not by position in nodeOrder. Falls
		// back to positional index only for a non-numeric id, which real (loadTree-built)
		// specs never produce (loud load-time error there); this keeps synthetic-id unit
		// tests that construct a MoveDispatch directly working unchanged.
		row := i
		if n, err := strconv.Atoi(id); err == nil {
			row = n - 1
		}
		md.GS.NodeSeeds = append(md.GS.NodeSeeds, geomseeds.BuildNodeSeed(id, i, g, row))
	}
	md.GS.EdgeSeeds = make([]geomseeds.EdgeGeomSeed, 0, len(edgeOrder))
	for _, label := range edgeOrder {
		ep, ok := edgeEndpoints[label]
		if !ok {
			continue
		}
		// Real endpoint geometry, computed by geomseeds.BuildEdgeSeed: the same
		// nodegeom.EdgeSegment computation recomputeGeometry (below) uses on every live
		// move, evaluated once here against the load-time geoms so the seed row is never
		// a degenerate 0,0,0->0,0,0 segment. A missed lookup means the edge's source or
		// target node id has no geometry — a malformed topology (most commonly a stale
		// edge file left behind after its target node's directory was deleted by hand;
		// in-edges are not indexed, so nothing else catches this) — and must fail the
		// load loudly, never seed a silent 0,0,0->0,0,0 segment indistinguishable from
		// real data.
		seed, err := geomseeds.BuildEdgeSeed(label, ep, geoms)
		if err != nil {
			return err
		}
		md.GS.EdgeSeeds = append(md.GS.EdgeSeeds, seed)
	}
	return nil
}

// buildRowTables builds the row-identity tables ONCE, from nodeSeeds/edgeSeeds (each node
// seed already carries its own absolute Row = id-1) — see RowTables.Build's doc comment for
// why this is a load-time constant, not a discovery log.
func (md *MoveDispatch) buildRowTables(rowCount int) {
	rtNodeSeeds := make([]rowtables.NodeSeed, len(md.GS.NodeSeeds))
	for i, sd := range md.GS.NodeSeeds {
		rtNodeSeeds[i] = rowtables.NodeSeed{ID: sd.ID, Row: sd.Row}
	}
	rtEdgeSeeds := make([]rowtables.EdgeSeed, len(md.GS.EdgeSeeds))
	for i, sd := range md.GS.EdgeSeeds {
		rtEdgeSeeds[i] = rowtables.EdgeSeed{Label: sd.Label, SrcNode: sd.SrcNode, DstNode: sd.DstNode}
	}
	md.RT.Build(rtNodeSeeds, rtEdgeSeeds, rowCount)
}
