package bufschema

// bufLayoutOutPole is one row per OUTGOING neighbour of a node: the unit
// direction of that neighbour's stored path vector. It is the +y pole of the
// per-edge frame the editor draws on the node — n outgoing neighbours, n poles.
type bufLayoutOutPole struct {
	DX float32 `buf:"f32"`
	DY float32 `buf:"f32"`
	DZ float32 `buf:"f32"`
}
