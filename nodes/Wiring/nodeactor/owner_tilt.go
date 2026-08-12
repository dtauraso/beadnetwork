package nodeactor

func (t *nodeTilt) SetTopTiltVectorThetaIdx(idx int32) {
	t.topTiltVectorThetaIdx = idx
}

func (t *nodeTilt) BumpTopTiltVectorThetaIdx(delta int32) {
	t.topTiltVectorThetaIdx += delta
}

func (t *nodeTilt) ResetTopTiltVectorThetaIdx() {
	t.topTiltVectorThetaIdx = 0
}

func (t *nodeTilt) TopTiltVectorThetaIdx() int32 { return t.topTiltVectorThetaIdx }

func (t *nodeTilt) SetTiltIndex(topIdx, normalIdx, bottomIdx int32) {
	t.topTiltVectorThetaIdx = topIdx
	t.normalThetaIdx = normalIdx
	t.bottomThetaIdx = bottomIdx
}

func (t *nodeTilt) SetReceivedVector(theta int32, set bool) {
	t.receivedVectorThetaIdx = theta
	t.receivedVectorSet = set
}

func (t *nodeTilt) SetLatticePoints(points int32) {
	t.latticePoints = points
}

func (t *nodeTilt) FrameGeometryFields() (topIdx, bottomIdx, normalIdx, receivedIdx int32, receivedSet bool, latticePoints int32) {
	return t.topTiltVectorThetaIdx, t.bottomThetaIdx, t.normalThetaIdx, t.receivedVectorThetaIdx, t.receivedVectorSet, t.latticePoints
}
