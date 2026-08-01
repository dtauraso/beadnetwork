// row_tables.go — the row-identity table owner split out of MoveDispatch (god-object
// decomposition), as a pure move (no logic changes): rowTables owns the four row-identity
// tables (node/edge/port/edge-endpoint) built once at load and read-only afterward.
// MoveDispatch keeps its public Lookup*/…RowFor methods as thin delegators so the
// external API is unchanged.

package Wiring

import ()

// moveDispatchEdgeEndpoint is one edge row's endpoint node ids.
type moveDispatchEdgeEndpoint struct {
	srcNode, dstNode string
}

// rowTables owns the three row-identity tables (hit-testing + mover row resolution). A
// port has no row of its own any more (docs/channels-not-ports.md — no Port block, no
// port hit kind), so this is three tables, not four.
//
// These tables used to live on a central accumulator, built as a side effect of
// the Trace-drain goroutine observing the FIRST geometry event for each node/edge — a
// discovery log. Node/edge row order is actually a LOAD-TIME CONSTANT (spec
// order, md.nodeSeeds/md.edgeSeeds — see their doc comments): nodes/edges are only
// ever added via respawn, which re-runs load from scratch. So the tables are built
// ONCE here, in newMoveDispatch, from the SAME nodeSeeds/edgeSeeds order the seed loop
// in main.go streams through tr.NodeGeometry/tr.Geometry — reproducing byte-for-byte
// the row order the old central accumulator used to discover independently. Built
// before Start (and never mutated afterward), so a plain slice/map here is already
// safe for every reader goroutine (gesture, movers) to read concurrently: the write
// happened-before every goroutine that could read it (Go launches nodeMover/edgeMover
// goroutines only in Start, which runs after newMoveDispatch returns).
type rowTables struct {
	// nodeRowTable: node ids in stable row order (== md.nodeSeeds order).
	nodeRowTable []string
	// edgeRowTable: edge labels in stable row order (== md.edgeSeeds order).
	edgeRowTable []string
	// edgeEndpointRowTable: each edge row's (srcNode, dstNode) ids, same order as edgeRowTable.
	edgeEndpointRowTable []moveDispatchEdgeEndpoint
}

// buildRowTables constructs the row-identity tables from md.nodeSeeds/md.edgeSeeds. Called
// once, before any mover goroutine exists.
//
// rowCount sizes the NODE row table (topoSpec.RowCount — the largest node id found, not the
// node count): rows 0..rowCount-1. Each seed places itself at its OWN seed.Row (id-1), not
// at its position in nodeSeeds — a gap in the id space is a row no seed ever writes, so it
// stays the empty string ("") rather than a later seed sliding up to fill it. 0 (only test
// call sites that pass no rowCount) falls back to len(nodeSeeds), i.e. no gaps.
func (rt *rowTables) buildRowTables(nodeSeeds []NodeGeomSeed, edgeSeeds []EdgeGeomSeed, rowCount int) {
	if rowCount == 0 {
		rowCount = len(nodeSeeds)
	}
	rt.nodeRowTable = make([]string, rowCount)
	for _, sd := range nodeSeeds {
		if sd.Row < 0 || sd.Row >= rowCount {
			continue
		}
		rt.nodeRowTable[sd.Row] = sd.ID
	}
	rt.edgeRowTable = make([]string, len(edgeSeeds))
	rt.edgeEndpointRowTable = make([]moveDispatchEdgeEndpoint, len(edgeSeeds))
	for i, sd := range edgeSeeds {
		rt.edgeRowTable[i] = sd.Label
		rt.edgeEndpointRowTable[i] = moveDispatchEdgeEndpoint{srcNode: sd.SrcNode, dstNode: sd.DstNode}
	}
}

// lookupNodeRow resolves a numeric buffer NODE-ROW index to its node id via the row table
// built at load. ok=false for an out-of-range row OR an in-range but EMPTY row — a gap left
// by a deleted node id (GAPS ARE LEGAL: they never shift later rows), which is exactly as
// unresolvable as out-of-range from a caller's point of view.
func (rt *rowTables) lookupNodeRow(row int) (nodeID string, ok bool) {
	if row < 0 || row >= len(rt.nodeRowTable) || rt.nodeRowTable[row] == "" {
		return "", false
	}
	return rt.nodeRowTable[row], true
}

// nodeRowFor resolves nodeID to its buffer NODE-ROW index via the row table built at
// load — the REVERSE of lookupNodeRow.
func (rt *rowTables) nodeRowFor(nodeID string) (int32, bool) {
	for i, id := range rt.nodeRowTable {
		if id == nodeID {
			return int32(i), true
		}
	}
	return -1, false
}

// lookupEdgeRow resolves a numeric buffer EDGE-ROW index to its edge label via the row
// table built at load. ok=false for an out-of-range row.
func (rt *rowTables) lookupEdgeRow(row int) (label string, ok bool) {
	if row < 0 || row >= len(rt.edgeRowTable) {
		return "", false
	}
	return rt.edgeRowTable[row], true
}

// edgeRowForPair resolves the buffer edge-row index of the bead edge connecting node ids
// a/b (in either direction) via the edge-endpoint table built at load. ok=false when no
// such edge exists.
func (rt *rowTables) edgeRowForPair(a, b string) (int32, bool) {
	for i, e := range rt.edgeEndpointRowTable {
		if (e.srcNode == a && e.dstNode == b) || (e.srcNode == b && e.dstNode == a) {
			return int32(i), true
		}
	}
	return -1, false
}
