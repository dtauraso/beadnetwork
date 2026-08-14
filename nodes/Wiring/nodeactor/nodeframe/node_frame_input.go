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

	PolePhi, PoleTheta float32

	RingAxisPhi, RingAxisTheta float32

	TopTiltVectorLen float32

	TopTiltVectorIdx int32

	TopTiltVectorPhi, TopTiltVectorTheta float32

	BottomTiltVectorPhi, BottomTiltVectorTheta float32

	CoplanarNormalPhi, CoplanarNormalTheta float32

	ReceivedVectorLen float32

	ReceivedVectorPhi, ReceivedVectorTheta float32

	Selected, KindID, Hovered, LatchedSel uint8

	LatticePoints uint8

	RoundsToParallel, MsgsToParallel int32

	Label string

	Events []rowevent.RowEvent
}

type NodeFrameBuilder func(f NodeFrameInput) []byte
