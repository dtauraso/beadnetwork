package bufschema

// One row per tilt arrow: the two transforms Go composed for it, and which mesh draws it.
// The render tree used to build these from the node centre, the vector length and a phi
// -- geometry derived in TS, and the reason TiltVectors read the node centre column.
type bufLayoutTiltArrow struct {
	Received uint8 `buf:"u8"`

	ShaftM0  float32 `buf:"f32"`
	ShaftM1  float32 `buf:"f32"`
	ShaftM2  float32 `buf:"f32"`
	ShaftM3  float32 `buf:"f32"`
	ShaftM4  float32 `buf:"f32"`
	ShaftM5  float32 `buf:"f32"`
	ShaftM6  float32 `buf:"f32"`
	ShaftM7  float32 `buf:"f32"`
	ShaftM8  float32 `buf:"f32"`
	ShaftM9  float32 `buf:"f32"`
	ShaftM10 float32 `buf:"f32"`
	ShaftM11 float32 `buf:"f32"`
	ShaftM12 float32 `buf:"f32"`
	ShaftM13 float32 `buf:"f32"`
	ShaftM14 float32 `buf:"f32"`
	ShaftM15 float32 `buf:"f32"`

	HeadM0  float32 `buf:"f32"`
	HeadM1  float32 `buf:"f32"`
	HeadM2  float32 `buf:"f32"`
	HeadM3  float32 `buf:"f32"`
	HeadM4  float32 `buf:"f32"`
	HeadM5  float32 `buf:"f32"`
	HeadM6  float32 `buf:"f32"`
	HeadM7  float32 `buf:"f32"`
	HeadM8  float32 `buf:"f32"`
	HeadM9  float32 `buf:"f32"`
	HeadM10 float32 `buf:"f32"`
	HeadM11 float32 `buf:"f32"`
	HeadM12 float32 `buf:"f32"`
	HeadM13 float32 `buf:"f32"`
	HeadM14 float32 `buf:"f32"`
	HeadM15 float32 `buf:"f32"`
}
