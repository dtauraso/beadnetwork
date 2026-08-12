package owners

type Tilt struct {
	topTiltVectorThetaIdx  int32
	normalThetaIdx         int32
	bottomThetaIdx         int32
	receivedVectorThetaIdx int32
	receivedVectorSet      bool

	latticePoints int32
}

func NewTilt(latticePoints int32) Tilt {
	return Tilt{latticePoints: latticePoints}
}

func (t *Tilt) SetTopTiltVectorThetaIdx(idx int32) {
	t.topTiltVectorThetaIdx = idx
}

func (t *Tilt) BumpTopTiltVectorThetaIdx(delta int32) {
	t.topTiltVectorThetaIdx += delta
}

func (t *Tilt) ResetTopTiltVectorThetaIdx() {
	t.topTiltVectorThetaIdx = 0
}

func (t *Tilt) TopTiltVectorThetaIdx() int32 { return t.topTiltVectorThetaIdx }

func (t *Tilt) SetTiltIndex(topIdx, normalIdx, bottomIdx int32) {
	t.topTiltVectorThetaIdx = topIdx
	t.normalThetaIdx = normalIdx
	t.bottomThetaIdx = bottomIdx
}

func (t *Tilt) SetReceivedVector(theta int32, set bool) {
	t.receivedVectorThetaIdx = theta
	t.receivedVectorSet = set
}

func (t *Tilt) SetLatticePoints(points int32) {
	t.latticePoints = points
}

func (t *Tilt) FrameGeometryFields() (topIdx, bottomIdx, normalIdx, receivedIdx int32, receivedSet bool, latticePoints int32) {
	return t.topTiltVectorThetaIdx, t.bottomThetaIdx, t.normalThetaIdx, t.receivedVectorThetaIdx, t.receivedVectorSet, t.latticePoints
}
