package Wiring

// gesture_hit.go — resolves a raw hit-test (numeric buffer-row index) back to Go's own
// topology identity (node id / edge label / port name). Go owns the topology and wrote the
// Node/Edge/Port blocks in row order, so a raw hit's row index is looked up through the same
// row tables built at load (buildRowTables).

// nodeFromHit resolves a node hit to its node id. A node hit carries only a numeric buffer
// NODE-ROW index (the node InstancedMesh instanceId == its buffer node row); Go maps it back
// through its own node-row table (built at load — see buildRowTables), since Go owns the
// topology and wrote the Node block in that same row order.
func (md *MoveDispatch) nodeFromHit(h rawHit) (node string, ok bool) {
	if h.NodeRow >= 0 {
		return md.LookupNodeRow(h.NodeRow)
	}
	return "", false
}

// edgeFromHit resolves an edge hit to its edge label. An edge hit carries only a numeric
// buffer EDGE-ROW index (no label string); Go maps it back through its own edge-row table
// (built at load — see buildRowTables), since Go owns the topology and wrote the Edge block
// in that same row order.
func (md *MoveDispatch) edgeFromHit(h rawHit) (label string, ok bool) {
	if h.EdgeRow >= 0 {
		return md.LookupEdgeRow(h.EdgeRow)
	}
	return "", false
}

// portConnected reports whether the named port has at least one incident edge. It scans the
// edge movers' endpoints (the held topology) — the FSM's own state, not a fact carried on
// the wire from TS.
func (md *MoveDispatch) portConnected(node, port string, isInput bool) bool {
	for _, em := range md.mr.edgeMovers {
		if isInput {
			if em.dstID == node && em.dstH == port {
				return true
			}
		} else {
			if em.srcID == node && em.srcH == port {
				return true
			}
		}
	}
	return false
}

// portFromHit resolves a port hit to its (node, port, isInput) identity. A port hit
// carries only a numeric buffer PORT-ROW index (no name string); Go maps it back through
// its own port-row table (built at load — see buildRowTables), since Go owns the topology
// and wrote the Port block in that same row order.
func (md *MoveDispatch) portFromHit(h rawHit) (node, port string, isInput, ok bool) {
	if h.PortRow >= 0 {
		return md.LookupPortRow(h.PortRow)
	}
	return "", "", false, false
}
