package PairNode

import (
	clock "github.com/dtauraso/wirefold/src/Clock"
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Node/nodeactor"
	"github.com/dtauraso/wirefold/src/NodeKinds/PairNode/tiltring"
)

type nodePlumbing struct {
	PairID int32

	Fire         func()
	EmitGeometry func()

	Clock clock.Clock

	SpeedCh <-chan float64

	In *beadanimation.Receiver

	Out *beadanimation.Sender

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
