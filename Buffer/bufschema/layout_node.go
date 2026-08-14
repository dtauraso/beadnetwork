package bufschema

type bufLayoutNode struct {
	NodeId  int32   `buf:"i32"`
	CX      float32 `buf:"f32"`
	CY      float32 `buf:"f32"`
	CZ      float32 `buf:"f32"`
	Radius  float32 `buf:"f32"`
	SphereR float32 `buf:"f32"`

	VRX float32 `buf:"f32"`
	VRY float32 `buf:"f32"`
	VRZ float32 `buf:"f32"`
	FRX float32 `buf:"f32"`
	FRY float32 `buf:"f32"`
	FRZ float32 `buf:"f32"`

	PolePhi   float32 `buf:"f32"`
	PoleTheta float32 `buf:"f32"`

	RingAxisPhi   float32 `buf:"f32"`
	RingAxisTheta float32 `buf:"f32"`

	TopTiltVectorLen float32 `buf:"f32"`

	TopTiltVectorTheta float32 `buf:"f32"`

	BottomTiltVectorTheta float32 `buf:"f32"`
	CoplanarNormalTheta   float32 `buf:"f32"`

	ReceivedVectorLen   float32 `buf:"f32"`
	ReceivedVectorTheta float32 `buf:"f32"`
	Selected            uint8   `buf:"u8"`

	KindId uint8 `buf:"u8"`

	LabelOff uint32 `buf:"u32"`
	LabelLen uint32 `buf:"u32"`

	Hovered uint8 `buf:"u8"`

	LatchedSel uint8 `buf:"u8"`

	LatticePoints uint8 `buf:"u8"`

	RoundsToParallel int32 `buf:"i32"`

	MsgsToParallel int32 `buf:"i32"`
}
