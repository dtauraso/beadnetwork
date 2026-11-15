package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/AngleArcs"
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
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
	BodyMatrix [16]float32

	TopTiltVectorLen float32

	TopTiltVectorIdx int32

	TiltArrows []TiltVectors.TiltArrow


	TickMarks []TickMark

	AngleArcs AngleArcs.Arcs

	Selected, KindID, LatchedSel uint8

	LatticePoints uint8



	Label string
}

type NodeFrameBuilder func(f NodeFrameInput)
