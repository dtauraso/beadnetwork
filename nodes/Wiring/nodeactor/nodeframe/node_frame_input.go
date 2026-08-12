package nodeframe

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

type NodeFrameInput struct {
	Tick uint32

	NodeRow int32
	NodeID  int32

	CX, CY, CZ float32

	Radius, SphereR float32

	VRX, VRY, VRZ float32
	FRX, FRY, FRZ float32

	PoleTheta, PolePhi float32

	RingAxisTheta, RingAxisPhi float32

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

	ChainBeadOX, ChainBeadOY, ChainBeadOZ []float32
	ChainBeadLit                          []uint8
	ChainBeadLitValue                     []int32

	Events []wire.RowEvent
}

type NodeFrameBuilder func(f NodeFrameInput) []byte
