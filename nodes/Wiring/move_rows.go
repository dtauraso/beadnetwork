// move_rows.go — the buffer-row query API on MoveDispatch: resolving numeric buffer row
// indices (node/edge/port) to/from their string identities, built once at load
// (row_tables.go).

package Wiring

// LookupNodeRow resolves a numeric buffer NODE-ROW index to its node id via the row table
// built at load. ok=false for an out-of-range row. This is the node analogue of
// LookupPortRow/LookupEdgeRow: a numeric node-row hit (the node InstancedMesh instanceId
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

// LayoutLinkPairs returns every LAYOUT cascade-link pair (id, to), one per unordered pair
// (id is always the alphabetically-first side — mirrors loader.go's emitLayoutLinks own
// de-dup rule), by walking each nodeMover's own layoutLinkTos (seeded once at load,
// static since — see its doc comment in node_mover.go). This is the SAME set
// emitLayoutLinks streams via tr.LayoutLink for the old central accumulator's LayoutLink BLOCK
// (still the sole source of that block; unaffected by this method), reconstructed here
// so main.go can ALSO emit each pair once as a view-owner VIEW-frame event (Step C,
// memory/feedback_no_single_writer_bridge.md — LayoutLink is load-time-once, like SceneSphere, so it has no
// live per-goroutine owner to decentralize onto; this seed-once emission is its
// decentralized counterpart). Order is not guaranteed to match emitLayoutLinks' — the
// .probe log is a multiset of events, per memory/feedback_no_single_writer_bridge.md's own doc comment.
func (md *MoveDispatch) LayoutLinkPairs() [][2]string {
	var out [][2]string
	for id, nm := range md.mr.nodeMovers {
		for _, to := range nm.layoutLinkTos {
			out = append(out, [2]string{id, to})
		}
	}
	return out
}

// EdgeRowForPair resolves the buffer edge-row index of the bead edge connecting node ids
// a/b (in either direction) via the edge-endpoint table built at load — safe to call from
// any goroutine (the table is read-only after construction). Used by a dedicated nodeMover
// goroutine to resolve its own layout-link's edge row for its per-node stream frame.
// ok=false when no such edge exists.
func (md *MoveDispatch) EdgeRowForPair(a, b string) (int32, bool) {
	return md.rt.edgeRowForPair(a, b)
}

// LookupPortRow resolves a numeric buffer PORT-ROW index to its (node, port, isInput)
// identity via the port-row table built at load. ok=false for an out-of-range row. This is
// the row→(node,port) resolution the gesture FSM uses for wiring/handhold — the numeric
// buffer carries no port strings.
func (md *MoveDispatch) LookupPortRow(row int) (node, port string, isInput, ok bool) {
	return md.rt.lookupPortRow(row)
}

// PortRowFor resolves (node, port, isInput) to its buffer PORT-ROW index via the port-row
// table built at load — the REVERSE of LookupPortRow. Used by a dedicated edgeMover
// goroutine to resolve its own SrcPortRow/DstPortRow for its per-edge stream frame.
// ok=false when no port matches.
func (md *MoveDispatch) PortRowFor(node, port string, isInput bool) (int32, bool) {
	return md.rt.portRowFor(node, port, isInput)
}
