package RingPoint

var _ = bufLayoutNodeRingPoint{}

type bufLayoutNodeRingPoint struct {
	X float32 `buf:"f32"`
	Y float32 `buf:"f32"`
	Z float32 `buf:"f32"`
}

var _ = bufLayoutBeadRingPoint{}

type bufLayoutBeadRingPoint struct {
	X float32 `buf:"f32"`
	Y float32 `buf:"f32"`
	Z float32 `buf:"f32"`
}
