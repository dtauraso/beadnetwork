package bufschema

type bufLayoutNodeRingPoint struct {
	X float32 `buf:"f32"`
	Y float32 `buf:"f32"`
	Z float32 `buf:"f32"`
}

type bufLayoutBeadRingPoint struct {
	X float32 `buf:"f32"`
	Y float32 `buf:"f32"`
	Z float32 `buf:"f32"`
}
