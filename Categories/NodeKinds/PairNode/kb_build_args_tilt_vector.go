package PairNode

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/Categories/Node/TiltVectors"
)

func (a BuildArgs) TiltVectorAngleSeed() (theta int32) {
	return a.TiltPhiIdx
}

func (a BuildArgs) TiltEditIn() <-chan TiltVectors.TiltEditMsg {
	if a.Deps == nil {
		return make(chan TiltVectors.TiltEditMsg)
	}
	ch, ok := a.Deps.TiltEditChan(a.Name).(chan TiltVectors.TiltEditMsg)
	if !ok {
		panic("TiltEditIn: the scene handed this kind something that is not a tilt-edit channel")
	}
	return ch
}

func (a BuildArgs) VectorOut() chan<- TiltPanel.TiltVectorMsg {
	return a.PB.VectorOutOf(a.Name)
}

func (a BuildArgs) VectorIn() <-chan TiltPanel.TiltVectorMsg {
	return a.PB.VectorInOf(a.Name)
}
