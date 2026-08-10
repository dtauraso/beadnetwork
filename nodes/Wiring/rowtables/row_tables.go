// Package rowtables holds the row-identity table owner split out of MoveDispatch (god-object
// decomposition), as a pure move (no logic changes): RowTables owns the three row-identity
// tables (node/edge/edge-endpoint) built once at load and read-only afterward. MoveDispatch
// keeps its public LookupNodeRow/NodeRowFor/LookupEdgeRow/EdgeRowForPair as thin delegators to
// its exported RT field so the external API is unchanged.
package rowtables

// EdgeEndpoint is one edge row's endpoint node ids.
type EdgeEndpoint struct {
	SrcNode, DstNode string
}

// NodeSeed is the subset of a node's load-time seed geometry that row-table construction
// needs: its id and its row (id-1). Callers build this from their own NodeGeomSeed slice
// so this package needs no import back onto Wiring's seed types (which would cycle, since
// Wiring imports this package for RowTables itself).
type NodeSeed struct {
	ID  string
	Row int
}

// EdgeSeed is the subset of an edge's load-time seed geometry that row-table construction
// needs: its label and its endpoint node ids, in spec order.
type EdgeSeed struct {
	Label            string
	SrcNode, DstNode string
}

// RowTables owns the three row-identity tables (hit-testing + mover row resolution). A
// port has no row of its own any more (docs/bead-model/channels-not-ports.md — no Port block, no
// port hit kind), so this is three tables, not four.
//
// These tables used to live on a central accumulator, built as a side effect of
// the Trace-drain goroutine observing the FIRST geometry event for each node/edge — a
// discovery log. Node/edge row order is actually a LOAD-TIME CONSTANT (spec
// order, md.gs.nodeSeeds/md.gs.edgeSeeds — see their doc comments): nodes/edges are only
// ever added via respawn, which re-runs load from scratch. So the tables are built
// ONCE here, in newMoveDispatch, from the SAME nodeSeeds/edgeSeeds order the seed loop
// in main.go streams through tr.NodeGeometry/tr.Geometry — reproducing byte-for-byte
// the row order the old central accumulator used to discover independently. Built
// before Start (and never mutated afterward), so a plain slice/map here is already
// safe for every reader goroutine (gesture, movers) to read concurrently: the write
// happened-before every goroutine that could read it (Go launches nodeMover/edgeMover
// goroutines only in Start, which runs after newMoveDispatch returns).
type RowTables struct {
	// NodeRowTable: node ids in stable row order (== md.gs.nodeSeeds order).
	NodeRowTable []string
	// EdgeRowTable: edge labels in stable row order (== md.gs.edgeSeeds order).
	EdgeRowTable []string
	// EdgeEndpointRowTable: each edge row's (srcNode, dstNode) ids, same order as EdgeRowTable.
	EdgeEndpointRowTable []EdgeEndpoint
}

// Build constructs the row-identity tables from the caller's own node/edge seeds. Called
// once, before any mover goroutine exists.
//
// rowCount sizes the NODE row table (topoSpec.RowCount — the largest node id found, not the
// node count): rows 0..rowCount-1. Each seed places itself at its OWN Row (id-1), not
// at its position in nodeSeeds — a gap in the id space is a row no seed ever writes, so it
// stays the empty string ("") rather than a later seed sliding up to fill it. 0 (only test
// call sites that pass no rowCount) falls back to len(nodeSeeds), i.e. no gaps.
func (rt *RowTables) Build(nodeSeeds []NodeSeed, edgeSeeds []EdgeSeed, rowCount int) {
	if rowCount == 0 {
		rowCount = len(nodeSeeds)
	}
	rt.NodeRowTable = make([]string, rowCount)
	for _, sd := range nodeSeeds {
		if sd.Row < 0 || sd.Row >= rowCount {
			continue
		}
		rt.NodeRowTable[sd.Row] = sd.ID
	}
	rt.EdgeRowTable = make([]string, len(edgeSeeds))
	rt.EdgeEndpointRowTable = make([]EdgeEndpoint, len(edgeSeeds))
	for i, sd := range edgeSeeds {
		rt.EdgeRowTable[i] = sd.Label
		rt.EdgeEndpointRowTable[i] = EdgeEndpoint{SrcNode: sd.SrcNode, DstNode: sd.DstNode}
	}
}

// LookupNodeRow resolves a numeric buffer NODE-ROW index to its node id via the row table
// built at load. ok=false for an out-of-range row OR an in-range but EMPTY row — a gap left
// by a deleted node id (GAPS ARE LEGAL: they never shift later rows), which is exactly as
// unresolvable as out-of-range from a caller's point of view.
func (rt *RowTables) LookupNodeRow(row int) (nodeID string, ok bool) {
	if row < 0 || row >= len(rt.NodeRowTable) || rt.NodeRowTable[row] == "" {
		return "", false
	}
	return rt.NodeRowTable[row], true
}

// NodeRowFor resolves nodeID to its buffer NODE-ROW index via the row table built at
// load — the REVERSE of LookupNodeRow.
func (rt *RowTables) NodeRowFor(nodeID string) (int32, bool) {
	for i, id := range rt.NodeRowTable {
		if id == nodeID {
			return int32(i), true
		}
	}
	return -1, false
}

// LookupEdgeRow resolves a numeric buffer EDGE-ROW index to its edge label via the row
// table built at load. ok=false for an out-of-range row.
func (rt *RowTables) LookupEdgeRow(row int) (label string, ok bool) {
	if row < 0 || row >= len(rt.EdgeRowTable) {
		return "", false
	}
	return rt.EdgeRowTable[row], true
}

// EdgeRowForPair resolves the buffer edge-row index of the bead edge connecting node ids
// a/b (in either direction) via the edge-endpoint table built at load. ok=false when no
// such edge exists.
func (rt *RowTables) EdgeRowForPair(a, b string) (int32, bool) {
	for i, e := range rt.EdgeEndpointRowTable {
		if (e.SrcNode == a && e.DstNode == b) || (e.SrcNode == b && e.DstNode == a) {
			return int32(i), true
		}
	}
	return -1, false
}
