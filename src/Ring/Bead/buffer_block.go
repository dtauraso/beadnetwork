package bead

var _ = bufLayoutEdgeBead{}

type bufLayoutEdgeBead struct {
	X float32 `buf:"f32"`
	Y float32 `buf:"f32"`
	Z float32 `buf:"f32"`

	Value int32 `buf:"i32"`

	EdgeRow int32 `buf:"i32"`

	RingM0  float32 `buf:"f32"`
	RingM1  float32 `buf:"f32"`
	RingM2  float32 `buf:"f32"`
	RingM3  float32 `buf:"f32"`
	RingM4  float32 `buf:"f32"`
	RingM5  float32 `buf:"f32"`
	RingM6  float32 `buf:"f32"`
	RingM7  float32 `buf:"f32"`
	RingM8  float32 `buf:"f32"`
	RingM9  float32 `buf:"f32"`
	RingM10 float32 `buf:"f32"`
	RingM11 float32 `buf:"f32"`
	RingM12 float32 `buf:"f32"`
	RingM13 float32 `buf:"f32"`
	RingM14 float32 `buf:"f32"`
	RingM15 float32 `buf:"f32"`
}
