package Buffer

// bufLayoutEdge defines one row of the edges column block.
// One row per edge (wire). Matched from KindGeometry trace events.
//
// SX..EZ is the edge's own straight SEGMENT (docs/bead-model/bead-lattice.md "Ownership": the
// edgeMover publishes the segment only) — NODE SURFACE TO NODE SURFACE along the
// centre-to-centre line (edgeSegment, nodes/Wiring/port_geometry.go), the same two
// points chain_beads.go anchors bead 0 and the last bead to. There is no port row to
// reference any more: a port stopped being a place (docs/bead-model/channels-not-ports.md), so this
// column pair is the edge's own emitted endpoints, not an index into a Port block that no
// longer exists (see layout.go's "Port block is GONE" note). This DOES reintroduce a second
// copy of a world position (the node's own center + torus radius, computed once here rather
// than read live) — the prior port-row indirection existed specifically to dodge that tear,
// but the tear it was dodging was itself the bug this rewrite removes (a port's PLACE
// floating apart from the node's own surface): the edgeMover recomputes this segment on
// every endpoint move (recomputeGeometry), same as it always has, so it is never more than
// one move-event stale — no different from the Node block's own center staying fresh.
type bufLayoutEdge struct {
	SX       float32 `buf:"f32"` // segment start x (source node surface, world)
	SY       float32 `buf:"f32"` // segment start y
	SZ       float32 `buf:"f32"` // segment start z
	EX       float32 `buf:"f32"` // segment end x (target node surface, world)
	EY       float32 `buf:"f32"` // segment end y
	EZ       float32 `buf:"f32"` // segment end z
	Selected uint8   `buf:"u8"`  // persistent: 1 = this edge is the click-selected edge
	// EdgeLabelOff/EdgeLabelLen are this edge's slice into the snapshot's trailing EDGE-LABEL
	// BYTES section (the label-section analogue for edges): EdgeLabelOff is the byte offset,
	// EdgeLabelLen the UTF-8 byte length. Edge labels are carried ONLY for the .probe buffer-
	// decoded log (geometry `edge`, select-edge) — the render/bridge path
	// still resolves an edge hit by row index (LookupEdgeRow), never by this string.
	// Concatenated in the same stable edge-row order as the Edge block.
	EdgeLabelOff uint32 `buf:"u32"` // byte offset into the edge-label-bytes section
	EdgeLabelLen uint32 `buf:"u32"` // edge-label UTF-8 byte length
}
