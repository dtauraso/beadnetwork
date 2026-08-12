package Buffer

type bufLayoutEvent struct {
	Kind          uint8  `buf:"u8"`
	NodeRow       int32  `buf:"i32"`
	PortRow       int32  `buf:"i32"`
	TargetRow     int32  `buf:"i32"`
	TargetPortRow int32  `buf:"i32"`
	EdgeRow       int32  `buf:"i32"`
	Slot          int32  `buf:"i32"`
	Value         int32  `buf:"i32"`
	Bead          uint32 `buf:"u32"`

	BeadSteps    float32 `buf:"f32"`
	SimLatencyMs float32 `buf:"f32"`
	X            float32 `buf:"f32"`
	Y            float32 `buf:"f32"`
	Z            float32 `buf:"f32"`
	F            float32 `buf:"f32"`

	Label uint8 `buf:"u8"`

	Debug uint8 `buf:"u8"`

	TextOff uint32 `buf:"u32"`
	TextLen uint32 `buf:"u32"`
}
