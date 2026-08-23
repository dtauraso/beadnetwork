package nodeframe

import (
	"github.com/dtauraso/wirefold/Categories/Node/ChannelVectors"
	"github.com/dtauraso/wirefold/Categories/Node/TiltVectors"
)

type TiltArrow struct {
	Received bool

	Shaft [16]float32
	Head  [16]float32
}

type NodeFrameInput struct {
	Tick uint32

	NodeRow int32
	NodeID  int32

	IndexR, IndexPhi, IndexTheta int32
	HasPos                       uint8

	Radius float32

	PoleRingR float32

	NavTubeR                                 float32
	PoleAnchorX, PoleAnchorY, PoleAnchorZ    float32
	LabelAnchorX, LabelAnchorY, LabelAnchorZ float32

	PolePhi, PoleTheta float32

	RingMatrix [16]float32

	TopTiltVectorLen float32

	TopTiltVectorIdx int32

	TiltArrows []TiltVectors.TiltArrow

	ChannelVectors []ChannelVectors.ChannelVector

	Selected, KindID, Hovered, LatchedSel uint8

	LatticePoints uint8

	RoundsToParallel, MsgsToParallel int32

	DragRLocked, DragPhiLocked uint8
	DragThetaMax               float32
	DragActive                 uint8
	HasKindRule                uint8
	KindRuleActive             uint8
	SelfRLocked, SelfPhiLocked uint8
	SelfThetaMax               float32
	SelfActive                 uint8

	RuleGroupID, RuleGroupSize int32

	Label string
}

type NodeFrameBuilder func(f NodeFrameInput)
