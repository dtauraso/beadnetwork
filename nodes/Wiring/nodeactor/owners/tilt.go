package owners

type Tilt struct {
	topTiltVectorPhiIdx  int32
	normalPhiIdx         int32
	bottomPhiIdx         int32
	receivedVectorPhiIdx int32
	receivedVectorSet    bool

	latticePoints int32
}

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

func (t *Tilt) SetTiltIndex(topIdx, normalIdx, bottomIdx int32) {
	t.topTiltVectorPhiIdx = topIdx
	t.normalPhiIdx = normalIdx
	t.bottomPhiIdx = bottomIdx
}

func (t *Tilt) SetReceivedVector(theta int32, set bool) {
	t.receivedVectorPhiIdx = theta
	t.receivedVectorSet = set
}

func (t *Tilt) SetLatticePoints(points int32) {
	t.latticePoints = points
}

func (t *Tilt) FrameGeometryFields() (topIdx, bottomIdx, normalIdx, receivedIdx int32, receivedSet bool, latticePoints int32) {
	return t.topTiltVectorPhiIdx, t.bottomPhiIdx, t.normalPhiIdx, t.receivedVectorPhiIdx, t.receivedVectorSet, t.latticePoints
}
