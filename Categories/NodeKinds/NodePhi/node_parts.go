package NodePhi

import (
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"github.com/dtauraso/beadnetwork/Categories/NodeKinds/NodePhi/tiltring"
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

	Self *Self
}

type tiltHeld struct {
	Top *tiltring.State

	Machine tiltring.Machine

	TiltEditIn <-chan TiltVectors.TiltEditMsg

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
