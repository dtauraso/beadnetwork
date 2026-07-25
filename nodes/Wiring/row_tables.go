// row_tables.go — the row-identity table owner split out of MoveDispatch (god-object
// decomposition), as a pure move (no logic changes): rowTables owns the four row-identity
// tables (node/edge/port/edge-endpoint) built once at load and read-only afterward.
// MoveDispatch keeps its public Lookup*/…RowFor methods as thin delegators so the
// external API is unchanged.

package Wiring

// moveDispatchPortRow is one row of the port-row table — the (node, port) identity a
// numeric buffer PORT-ROW index resolves to. Mirrors Buffer.PortRowEntry's shape (kept
// as a local type so this package stays Buffer-independent).
type moveDispatchPortRow struct {
	node    string
	port    string
	isInput bool
}

// moveDispatchEdgeEndpoint is one edge row's endpoint node ids.
type moveDispatchEdgeEndpoint struct {
	srcNode, dstNode string
}

// rowTables owns the four row-identity tables (hit-testing + mover row resolution).
//
// These four tables used to live on a central accumulator, built as a side effect of
// the Trace-drain goroutine observing the FIRST geometry event for each node/edge — a
// discovery log. Node/edge/port row order is actually a LOAD-TIME CONSTANT (spec
// order, md.nodeSeeds/md.edgeSeeds — see their doc comments): nodes/edges are only
// ever added via respawn, which re-runs load from scratch. So the tables are built
// ONCE here, in newMoveDispatch, from the SAME nodeSeeds/edgeSeeds order the seed loop
// in main.go streams through tr.NodeGeometry/tr.Geometry — reproducing byte-for-byte
// the row order the old central accumulator used to discover independently. Built
// before Start (and never mutated afterward), so — unlike that accumulator's
// atomic.Pointer tables — a plain
// slice/map here is already safe for every reader goroutine (gesture, movers) to read
// concurrently: the write happened-before every goroutine
// that could read it (Go launches nodeMover/edgeMover goroutines only in Start, which
// runs after newMoveDispatch returns).
type rowTables struct {
	// nodeRowTable: node ids in stable row order (== md.nodeSeeds order).
	nodeRowTable []string
	// edgeRowTable: edge labels in stable row order (== md.edgeSeeds order).
	edgeRowTable []string
	// portRowTable: the flattened port-row table in the SAME order Buffer's Port block is
	// written — node-row order × each node's Ports order (== md.nodeSeeds[i].Ports order).
	portRowTable []moveDispatchPortRow
	// edgeEndpointRowTable: each edge row's (srcNode, dstNode) ids, same order as edgeRowTable.
	edgeEndpointRowTable []moveDispatchEdgeEndpoint
}

// buildRowTables constructs the four row-identity tables from md.nodeSeeds/md.edgeSeeds
// (already in stable spec order at this point in newMoveDispatch). Called once, before
// any mover goroutine exists.
func (rt *rowTables) buildRowTables(nodeSeeds []NodeGeomSeed, edgeSeeds []EdgeGeomSeed) {
	rt.nodeRowTable = make([]string, len(nodeSeeds))
	for i, sd := range nodeSeeds {
		rt.nodeRowTable[i] = sd.ID
	}
	rt.portRowTable = rt.portRowTable[:0]
	for _, sd := range nodeSeeds {
		for _, p := range sd.Ports {
			rt.portRowTable = append(rt.portRowTable, moveDispatchPortRow{node: sd.ID, port: p.Name, isInput: p.IsInput})
		}
	}
	rt.edgeRowTable = make([]string, len(edgeSeeds))
	rt.edgeEndpointRowTable = make([]moveDispatchEdgeEndpoint, len(edgeSeeds))
	for i, sd := range edgeSeeds {
		rt.edgeRowTable[i] = sd.Label
		rt.edgeEndpointRowTable[i] = moveDispatchEdgeEndpoint{srcNode: sd.SrcNode, dstNode: sd.DstNode}
	}
}

// lookupNodeRow resolves a numeric buffer NODE-ROW index to its node id via the row table
// built at load. ok=false for an out-of-range row.
func (rt *rowTables) lookupNodeRow(row int) (nodeID string, ok bool) {
	if row < 0 || row >= len(rt.nodeRowTable) {
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

// lookupPortRow resolves a numeric buffer PORT-ROW index to its (node, port, isInput)
// identity via the port-row table built at load. ok=false for an out-of-range row.
func (rt *rowTables) lookupPortRow(row int) (node, port string, isInput, ok bool) {
	if row < 0 || row >= len(rt.portRowTable) {
		return "", "", false, false
	}
	e := rt.portRowTable[row]
	return e.node, e.port, e.isInput, true
}

// portRowFor resolves (node, port, isInput) to its buffer PORT-ROW index via the port-row
// table built at load — the REVERSE of lookupPortRow.
func (rt *rowTables) portRowFor(node, port string, isInput bool) (int32, bool) {
	for i, e := range rt.portRowTable {
		if e.node == node && e.port == port && e.isInput == isInput {
			return int32(i), true
		}
	}
	return -1, false
}
