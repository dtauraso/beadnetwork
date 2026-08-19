package kindapi

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/movemsg"
	"github.com/dtauraso/wirefold/src/TiltPanel"
)

func (a BuildArgs) TiltVectorAngleSeed() (theta int32) {
	return a.tiltPhiIdx
}

func (a BuildArgs) TiltEditIn() <-chan movemsg.TiltEditMsg {
	if a.deps.ClaimTiltEditIn == nil {
		return make(chan movemsg.TiltEditMsg)
	}
	return a.deps.ClaimTiltEditIn(a.name)
}

func (a BuildArgs) VectorOut() chan<- TiltPanel.TiltVectorMsg {
	if a.pb.VectorOut == nil {
		return nil
	}
	return a.pb.VectorOut[a.name]
}

func (a BuildArgs) VectorIn() <-chan TiltPanel.TiltVectorMsg {
	if a.pb.VectorIn == nil {
		return nil
	}
	return a.pb.VectorIn[a.name]
}
