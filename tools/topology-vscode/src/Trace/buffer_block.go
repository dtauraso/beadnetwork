package Trace

var _ = bufLayoutRecv{}

type bufLayoutRecv struct {
	NodeRow int32 `buf:"i32"`
	Value   int32 `buf:"i32"`
}

var _ = bufLayoutFire{}

type bufLayoutFire struct {
	NodeRow int32 `buf:"i32"`
}

var _ = bufLayoutSend{}

type bufLayoutSend struct {
	NodeRow   int32   `buf:"i32"`
	TargetRow int32   `buf:"i32"`
	Value     int32   `buf:"i32"`
	BeadSteps float32 `buf:"f32"`
}

var _ = bufLayoutArrive{}

type bufLayoutArrive struct {
	NodeRow int32  `buf:"i32"`
	Value   int32  `buf:"i32"`
	Bead    uint32 `buf:"u32"`
}

var _ = bufLayoutBreadcrumb{}

type bufLayoutBreadcrumb struct {
	NodeRow       int32 `buf:"i32"`
	PortRow       int32 `buf:"i32"`
	TargetRow     int32 `buf:"i32"`
	TargetPortRow int32 `buf:"i32"`
	EdgeRow       int32 `buf:"i32"`
	Slot          int32 `buf:"i32"`
	Value         int32 `buf:"i32"`

	X float32 `buf:"f32"`
	Y float32 `buf:"f32"`
	Z float32 `buf:"f32"`

	Label uint8 `buf:"u8"`
	Debug uint8 `buf:"u8"`

	TextOff uint32 `buf:"u32"`
	TextLen uint32 `buf:"u32"`
}
