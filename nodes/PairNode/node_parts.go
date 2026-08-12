package PairNode

import (
	"github.com/dtauraso/wirefold/nodes/wire/clock"
	"github.com/dtauraso/wirefold/nodes/wire/inport"
	"github.com/dtauraso/wirefold/nodes/wire/outport"

	"github.com/dtauraso/wirefold/nodes/PairNode/tiltring"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

type nodePlumbing struct {
	PairID int32

	Fire         func()
	EmitGeometry func()

	Clock clock.Clock

	SpeedCh <-chan float64

	In *inport.In

	Out *outport.Out

	ClearOutBeads func()

	Self *nodeactor.PairNodeSelf
}

type tiltHeld struct {
	Top *tiltring.State

	Bottom *tiltring.State

	Machine tiltring.Machine

	TiltEditIn <-chan movemsg.TiltEditMsg

	SyncTiltIndex func(theta, normalTheta, bottomTheta int32)
}

type latticeState struct {
	Ring *tiltring.Ring

	LatticeIn <-chan int32

	SyncLatticePoints func(points int32)
}

type vectorExchange struct {
	VectorOut chan<- tiltvector.TiltVectorMsg
	VectorIn  <-chan tiltvector.TiltVectorMsg

	ReceivedThetaIdx int32
	ReceivedSet      bool

	SyncReceivedVector func(theta int32, set bool)
}

type restCounters struct {
	msgsSinceOpen int32

	roundsSinceOpen int32

	roundsAtRest int32
	msgsAtRest   int32
	restReported bool

	restedThisCycle bool
}
