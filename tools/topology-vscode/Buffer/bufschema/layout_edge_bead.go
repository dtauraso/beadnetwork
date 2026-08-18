package bufschema

type bufLayoutEdgeBead struct {
	X float32 `buf:"f32"`
	Y float32 `buf:"f32"`
	Z float32 `buf:"f32"`

	Value int32 `buf:"i32"`

	EdgeRow int32 `buf:"i32"`

	RingMatrix []float32 `buf:"f32run"`
}
