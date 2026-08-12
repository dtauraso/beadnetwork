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
