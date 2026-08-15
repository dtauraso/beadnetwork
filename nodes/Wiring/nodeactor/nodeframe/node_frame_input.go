package nodeframe

import (
	"github.com/dtauraso/wirefold/nodes/rowevent"
)

type NodeFrameInput struct {
	Tick uint32

	NodeRow int32
	NodeID  int32

	CX, CY, CZ float32

	Radius float32

	VRX, VRY, VRZ float32
	FRX, FRY, FRZ float32

	PolePhi, PoleTheta float32

	RingAxisPhi, RingAxisTheta float32

	RingMatrix [16]float32

	TopTiltVectorLen float32

	TopTiltVectorIdx int32

	TopTiltVectorPhi float32

	BottomTiltVectorPhi float32

	CoplanarNormalPhi float32

	ReceivedVectorLen float32

	ReceivedVectorPhi float32

	Selected, KindID, Hovered, LatchedSel uint8

	LatticePoints uint8

	RoundsToParallel, MsgsToParallel int32

	DragRLocked, DragPhiLocked uint8
	DragThetaMax               float32
	DragActive                 uint8
	HasKindRule                uint8

	RuleGroupID, RuleGroupSize int32

	Label string

	Events []rowevent.RowEvent
}

type NodeFrameBuilder func(f NodeFrameInput) []byte
