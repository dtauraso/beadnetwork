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
