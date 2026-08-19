package Scene

var _ = bufLayoutScene{}

type bufLayoutScene struct {
	CX     float32 `buf:"f32"`
	CY     float32 `buf:"f32"`
	CZ     float32 `buf:"f32"`
	Radius float32 `buf:"f32"`

	ConstantR     float32 `buf:"f32"`
	MaxIndexPhi   int32   `buf:"i32"`
	MaxIndexTheta int32   `buf:"i32"`

	NodeCount int32 `buf:"i32"`
	EdgeCount int32 `buf:"i32"`
}
