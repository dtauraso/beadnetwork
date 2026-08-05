// move_rows.go — the buffer-row query API on MoveDispatch: resolving numeric buffer row
// indices (node/edge/port) to/from their string identities, built once at load
// (row_tables.go).

package Wiring

// LookupNodeRow resolves a numeric buffer NODE-ROW index to its node id via the row table
// built at load. ok=false for an out-of-range row. This is the node analogue of
// LookupEdgeRow: a numeric node-row hit (the node InstancedMesh instanceId
// == its buffer node row) resolves back to the node id here in Go, so the numeric buffer
// carries no node id strings and the webview forwards only the row.
func (md *MoveDispatch) LookupNodeRow(row int) (nodeID string, ok bool) {
	return md.rt.lookupNodeRow(row)
}

// NodeRowFor resolves nodeID to its buffer NODE-ROW index via the row table built at
// load — the REVERSE of LookupNodeRow. Used by a dedicated nodeMover/edgeMover goroutine
// (memory/feedback_no_single_writer_bridge.md) to resolve a neighbor's node row for its
// own per-goroutine stream frame (breadcrumb TargetRow, edge endpoint rows). ok=false when
// nodeID is not a registered node.
func (md *MoveDispatch) NodeRowFor(nodeID string) (int32, bool) {
	return md.rt.nodeRowFor(nodeID)
}

// LookupEdgeRow resolves a numeric buffer EDGE-ROW index to its edge label via the row
// table built at load. ok=false for an out-of-range row. This is the row→edge resolution
// the gesture FSM uses to mark the Go-owned edge selection — the numeric buffer carries
// no edge label strings.
func (md *MoveDispatch) LookupEdgeRow(row int) (label string, ok bool) {
	return md.rt.lookupEdgeRow(row)
}

// EdgeRowForPair resolves the buffer edge-row index of the bead edge connecting node ids
// a/b (in either direction) via the edge-endpoint table built at load — safe to call from
// any goroutine (the table is read-only after construction). General-purpose accessor.
// ok=false when no such edge exists.
func (md *MoveDispatch) EdgeRowForPair(a, b string) (int32, bool) {
	return md.rt.edgeRowForPair(a, b)
}
