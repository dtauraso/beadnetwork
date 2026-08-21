package Node

var _ = bufLayoutNode{}

type bufLayoutNode struct {

	IndexR     int32 `buf:"i32"`
	IndexPhi   int32 `buf:"i32"`
	IndexTheta int32 `buf:"i32"`

	HasPos uint8 `buf:"u8"`

	Radius float32 `buf:"f32"`

	NavTubeR float32 `buf:"f32"`

	PoleAnchorX float32 `buf:"f32"`
	PoleAnchorY float32 `buf:"f32"`
	PoleAnchorZ float32 `buf:"f32"`

	LabelAnchorX float32 `buf:"f32"`
	LabelAnchorY float32 `buf:"f32"`
	LabelAnchorZ float32 `buf:"f32"`

	PolePhi   float32 `buf:"f32"`
	PoleTheta float32 `buf:"f32"`

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


	TopTiltVectorIdx int32 `buf:"i32"`

	Selected uint8 `buf:"u8"`

	KindId uint8 `buf:"u8"`

	Label []byte `buf:"bytes"`

	Hovered uint8 `buf:"u8"`

	LatchedSel uint8 `buf:"u8"`


	RoundsToParallel int32 `buf:"i32"`

	MsgsToParallel int32 `buf:"i32"`

	DragRLocked    uint8   `buf:"u8"`
	DragPhiLocked  uint8   `buf:"u8"`
	DragThetaMax   float32 `buf:"f32"`
	DragActive     uint8   `buf:"u8"`
	KindRuleActive uint8   `buf:"u8"`

	PoleRingR float32 `buf:"f32"`

	SelfRLocked   uint8   `buf:"u8"`
	SelfPhiLocked uint8   `buf:"u8"`
	SelfThetaMax  float32 `buf:"f32"`
	SelfActive    uint8   `buf:"u8"`

	RuleGroupId   int32 `buf:"i32"`
	RuleGroupSize int32 `buf:"i32"`
}
