package bufschema

type bufLayoutTiltArrow struct {
	Received uint8 `buf:"u8"`

	Shaft []float32 `buf:"f32run"`

	Head []float32 `buf:"f32run"`
}
