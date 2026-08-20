package PairNode

import (
	"github.com/dtauraso/wirefold/src/Node/wire/inport"
	"github.com/dtauraso/wirefold/src/Node/wire/outport"
	"github.com/dtauraso/wirefold/src/Node/clock"

	"github.com/dtauraso/wirefold/src/NodeKinds/PairNode/tiltring"
	"github.com/dtauraso/wirefold/src/Node/Wiring/movemsg"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/src/TiltPanel"
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

	Machine tiltring.Machine

	TiltEditIn <-chan movemsg.TiltEditMsg

	SyncTiltIndex func(theta int32)
}

type latticeState struct {
	Ring *tiltring.Ring

	LatticeIn <-chan int32

	SyncLatticePoints func(points int32)
}

type vectorExchange struct {
	VectorOut chan<- TiltPanel.TiltVectorMsg
	VectorIn  <-chan TiltPanel.TiltVectorMsg

	ReceivedPhiIdx int32
	ReceivedSet    bool

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
