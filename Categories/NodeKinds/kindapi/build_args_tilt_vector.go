package kindapi

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/Categories/Node/TiltVectors"
)

func (a BuildArgs) TiltVectorAngleSeed() (theta int32) {
	return a.TiltPhiIdx
}

func (a BuildArgs) TiltEditIn() <-chan TiltVectors.TiltEditMsg {
	if a.Deps.ClaimTiltEditIn == nil {
		return make(chan TiltVectors.TiltEditMsg)
	}
	ch, ok := a.Deps.ClaimTiltEditIn(a.Name).(chan TiltVectors.TiltEditMsg)
	if !ok {
		panic("TiltEditIn: the scene handed this kind something that is not a tilt-edit channel")
	}
	return ch
}

func (a BuildArgs) VectorOut() chan<- TiltPanel.TiltVectorMsg {
	if a.PB.VectorOut == nil {
		return nil
	}
	return a.PB.VectorOut[a.Name]
}

func (a BuildArgs) VectorIn() <-chan TiltPanel.TiltVectorMsg {
	if a.PB.VectorIn == nil {
		return nil
	}
	return a.PB.VectorIn[a.Name]
}
