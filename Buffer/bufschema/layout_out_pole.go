package bufschema

// bufLayoutOutPole is one row per OUTGOING neighbour of a node: that
// neighbour's stored path vector, as stored. It is the +y pole of the per-edge
// frame the editor draws on the node — n outgoing neighbours, n poles. The
// renderer scales it to unit length when it builds the frame's orientation.
type bufLayoutOutPole struct {
	DX float32 `buf:"f32"`
	DY float32 `buf:"f32"`
	DZ float32 `buf:"f32"`
}
