package nodeframe

import (
	"github.com/dtauraso/wirefold/nodes/rowevent"
)

type NodeFrameInput struct {
	Tick uint32

	NodeRow int32
	NodeID  int32

	CX, CY, CZ float32

	Radius, SphereR float32

	VRX, VRY, VRZ float32
	FRX, FRY, FRZ float32

	RingAxisPhi, RingAxisTheta float32

	TopTiltVectorLen float32

	TopTiltVectorTheta float32

	BottomTiltVectorTheta float32

	CoplanarNormalTheta float32

	ReceivedVectorLen float32

	ReceivedVectorTheta float32

	Selected, KindID, Hovered, LatchedSel uint8

	LatticePoints uint8

	RoundsToParallel, MsgsToParallel int32

	Label string

	Events []rowevent.RowEvent
}

type NodeFrameBuilder func(f NodeFrameInput) []byte
