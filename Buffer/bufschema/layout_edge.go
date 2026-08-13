package bufschema

type bufLayoutEdge struct {
	SX float32 `buf:"f32"`
	SY float32 `buf:"f32"`
	SZ float32 `buf:"f32"`
	EX float32 `buf:"f32"`
	EY float32 `buf:"f32"`
	EZ float32 `buf:"f32"`

	// SrcNodeRow is the row of the node this edge LEAVES. The segment says
	// where the edge is drawn but not whose it is, so nothing downstream
	// could tell an edge apart by the node it comes from — which is what
	// deciding an edge's appearance from its source node's kind needs.
	SrcNodeRow int32 `buf:"i32"`

	EdgeLabelOff uint32 `buf:"u32"`
	EdgeLabelLen uint32 `buf:"u32"`
}
