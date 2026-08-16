package bufschema

type bufLayoutNode struct {
	NodeId int32   `buf:"i32"`
	CX     float32 `buf:"f32"`
	CY     float32 `buf:"f32"`
	CZ     float32 `buf:"f32"`
	Radius float32 `buf:"f32"`

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

	TopTiltVectorLen float32 `buf:"f32"`

	TopTiltVectorIdx int32 `buf:"i32"`

	TopTiltVectorPhi float32 `buf:"f32"`

	BottomTiltVectorPhi float32 `buf:"f32"`
	CoplanarNormalPhi   float32 `buf:"f32"`

	ReceivedVectorLen float32 `buf:"f32"`
	ReceivedVectorPhi float32 `buf:"f32"`
	Selected          uint8   `buf:"u8"`

	KindId uint8 `buf:"u8"`

	LabelOff uint32 `buf:"u32"`
	LabelLen uint32 `buf:"u32"`

	Hovered uint8 `buf:"u8"`

	LatchedSel uint8 `buf:"u8"`

	LatticePoints uint8 `buf:"u8"`

	RoundsToParallel int32 `buf:"i32"`

	MsgsToParallel int32 `buf:"i32"`

	DragRLocked    uint8   `buf:"u8"`
	DragPhiLocked  uint8   `buf:"u8"`
	DragThetaMax   float32 `buf:"f32"`
	DragActive     uint8   `buf:"u8"`
	HasKindRule    uint8   `buf:"u8"`
	KindRuleActive uint8   `buf:"u8"`

	PoleRingR float32 `buf:"f32"`

	SelfRLocked   uint8   `buf:"u8"`
	SelfPhiLocked uint8   `buf:"u8"`
	SelfThetaMax  float32 `buf:"f32"`
	SelfActive    uint8   `buf:"u8"`

	RuleGroupId   int32 `buf:"i32"`
	RuleGroupSize int32 `buf:"i32"`
}
