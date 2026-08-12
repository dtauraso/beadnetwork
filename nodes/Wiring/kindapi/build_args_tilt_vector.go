package kindapi

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

func (a BuildArgs) TiltVectorAngleSeed() (theta int32) {
	return a.tiltThetaIdx
}

func (a BuildArgs) TiltEditIn() <-chan movemsg.TiltEditMsg {
	if a.deps.ClaimTiltEditIn == nil {
		return make(chan movemsg.TiltEditMsg)
	}
	return a.deps.ClaimTiltEditIn(a.name)
}

func (a BuildArgs) VectorOut() chan<- tiltvector.TiltVectorMsg {
	if a.pb.VectorOut == nil {
		return nil
	}
	return a.pb.VectorOut[a.name]
}

func (a BuildArgs) VectorIn() <-chan tiltvector.TiltVectorMsg {
	if a.pb.VectorIn == nil {
		return nil
	}
	return a.pb.VectorIn[a.name]
}
