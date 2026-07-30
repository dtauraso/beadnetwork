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
// load — the REVERSE of LookupNodeRow. Used by a dedicated nodeMover goroutine
// (memory/feedback_no_single_writer_bridge.md) to resolve its own layout-link's dst node
// row for its per-node stream frame. ok=false when nodeID is not a registered node.
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

// LayoutLinkPairs returns every cascade-link pair (id, to), one per unordered pair (id is
// always the lexicographically-smaller side — mirrors nodeMover.emitGeometry's own de-dup
// rule), by walking each nodeMover's own cascadeEdges (loaded once at construction from
// nodes/<id>/cascade-edges.json — see its doc comment in node_mover.go). This is used so
// main.go can ALSO emit each pair once as a view-owner VIEW-frame event (Step C,
// memory/feedback_no_single_writer_bridge.md — the cascade-link overlay pairs are
// load-time-once, like SceneSphere, so they have no live per-goroutine owner to
// decentralize onto; this seed-once emission is its decentralized counterpart). Order is
// not guaranteed — the .probe log is a multiset of events, per
// memory/feedback_no_single_writer_bridge.md's own doc comment.
func (md *MoveDispatch) LayoutLinkPairs() [][2]string {
	var out [][2]string
	for id, nm := range md.mr.nodeMovers {
		for _, to := range nm.cascadeEdges {
			if id < to {
				out = append(out, [2]string{id, to})
			}
		}
	}
	return out
}

// EdgeRowForPair resolves the buffer edge-row index of the bead edge connecting node ids
// a/b (in either direction) via the edge-endpoint table built at load — safe to call from
// any goroutine (the table is read-only after construction). General-purpose accessor; no
// longer used by the cascade-link overlay (which draws node-center to node-center, not
// along a bead edge). ok=false when no such edge exists.
func (md *MoveDispatch) EdgeRowForPair(a, b string) (int32, bool) {
	return md.rt.edgeRowForPair(a, b)
}
