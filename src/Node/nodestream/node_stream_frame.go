package nodestream

import (
	"encoding/binary"
	"github.com/dtauraso/wirefold/src/Node/rowevent"

	TiltB "github.com/dtauraso/wirefold/src/Scene/TiltVectors"
	VecB "github.com/dtauraso/wirefold/src/Scene/Vectors"
	T "github.com/dtauraso/wirefold/src/Trace"
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
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

	TiltArrows []TiltB.TiltArrow

	ChannelVectors []VecB.ChannelVector

	Events []rowevent.RowEvent
}

func BuildNodeStreamFrame(f NodeStreamFrame) []byte {
	buf := make([]byte, B.BufNodeStreamFrameHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:], f.Tick)
	return append(buf, T.BuildEventsSection(f.Events)...)
}

func BuildInteriorStreamFrame(tick uint32, events []rowevent.RowEvent) []byte {
	buf := make([]byte, B.BufInteriorStreamFrameHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	return append(buf, T.BuildEventsSection(events)...)
}
