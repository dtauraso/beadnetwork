package streamframe

import (
	"encoding/binary"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	B "github.com/dtauraso/wirefold/src/Buffer"
)

type NodeStreamFrame struct {
	Tick uint32

	NodeRow int32

	NodeID int32

	IndexR, IndexPhi, IndexTheta int32
	HasPos                       uint8

	Radius float32

	NavTubeR                                 float32
	PoleAnchorX, PoleAnchorY, PoleAnchorZ    float32
	LabelAnchorX, LabelAnchorY, LabelAnchorZ float32

	PolePhi, PoleTheta float32

	RingMatrix [16]float32

	TopTiltVectorLen float32

	TopTiltVectorIdx int32

	Selected, KindID, Hovered, LatchedSel uint8

	LatticePoints uint8

	RoundsToParallel, MsgsToParallel int32

	DragRLocked, DragPhiLocked uint8
	DragThetaMax               float32
	DragActive                 uint8
	HasKindRule                uint8
	KindRuleActive             uint8
	PoleRingR                  float32

	SelfRLocked, SelfPhiLocked uint8
	SelfThetaMax               float32
	SelfActive                 uint8

	RuleGroupID, RuleGroupSize int32

	Label string

	TiltArrows []TiltArrow

	ChannelVectors []ChannelVector

	Events []rowevent.RowEvent
}

type TiltArrow struct {
	Received uint8
	Shaft    [16]float32
	Head     [16]float32
}

type ChannelVector struct {
	Shaft [16]float32
	Head  [16]float32
}

func BuildNodeStreamFrame(f NodeStreamFrame) []byte {
	buf := make([]byte, B.BufNodeStreamFrameHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:], f.Tick)
	return append(buf, BuildEventsSection(f.Events)...)
}

func BuildInteriorStreamFrame(tick uint32, events []rowevent.RowEvent) []byte {
	buf := make([]byte, B.BufInteriorStreamFrameHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	return append(buf, BuildEventsSection(events)...)
}
