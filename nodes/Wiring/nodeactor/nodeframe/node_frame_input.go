package nodeframe

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/framegeom"
	"github.com/dtauraso/wirefold/nodes/rowevent"
)

type NodeFrameInput struct {
	Tick uint32

	NodeRow int32
	NodeID  int32

	CX, CY, CZ float32

	Radius float32

	PoleRingR float32

	PolePhi, PoleTheta float32

	RingMatrix [16]float32

	TopTiltVectorLen float32

	TopTiltVectorIdx int32

	TiltArrows []framegeom.TiltArrow

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

	Events []rowevent.RowEvent
}

type NodeFrameBuilder func(f NodeFrameInput) []byte
