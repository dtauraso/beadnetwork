package TiltVectors

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/beadnetwork/Categories/NodeKinds/NodePhi/tiltring"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type Tilt struct {
	topTiltVectorPhiIdx  int32
	receivedVectorPhiIdx int32
	receivedVectorSet    bool

	latticePoints int32

	tickCount int32

	startVectors []polarindex.Offset
}

func (t *Tilt) SetStartVectors(v []polarindex.Offset) { t.startVectors = v }

func (t *Tilt) StartVectors() []polarindex.Offset { return t.startVectors }

func (t *Tilt) SetTickCount(count int32) { t.tickCount = count }

func (t *Tilt) TickCount() int32 { return t.tickCount }

func NewTilt(latticePoints int32) Tilt {
	return Tilt{latticePoints: latticePoints}
}

func (t *Tilt) SetTopTiltVectorPhiIdx(idx int32) {
	t.topTiltVectorPhiIdx = idx
}

func (t *Tilt) BumpTopTiltVectorPhiIdx(delta int32) {
	t.topTiltVectorPhiIdx += delta
}

func (t *Tilt) ResetTopTiltVectorPhiIdx() {
	t.topTiltVectorPhiIdx = 0
}

func (t *Tilt) TopTiltVectorPhiIdx() int32 { return t.topTiltVectorPhiIdx }

func (t *Tilt) SetTiltIndex(topIdx int32) {
	t.topTiltVectorPhiIdx = topIdx
}

func (t *Tilt) points() int32 {
	if t.latticePoints <= 0 {
		return TiltPanel.FullTurnPhiIdx
	}
	return t.latticePoints
}

func (t *Tilt) SetReceivedVector(theta int32, set bool) {
	t.receivedVectorPhiIdx = theta
	t.receivedVectorSet = set
}

func (t *Tilt) SetLatticePoints(points int32) {
	t.latticePoints = points
}

func (t *Tilt) FrameGeometryFields() (topIdx, bottomIdx, normalIdx, receivedIdx int32, receivedSet bool, latticePoints int32) {
	top := t.topTiltVectorPhiIdx
	tau := t.points()
	return top, tiltring.Bottom(top, tau), tiltring.Sent(top, tau),
		t.receivedVectorPhiIdx, t.receivedVectorSet, t.latticePoints
}
