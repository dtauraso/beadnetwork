package gatecommon

import (
	clock "github.com/dtauraso/wirefold/Clock"
	beadanimation "github.com/dtauraso/wirefold/Node/BeadAnimation"

	interior "github.com/dtauraso/wirefold/Node/Interior"
	"github.com/dtauraso/wirefold/Node/nodeactor"
)

const WindowMs = 3000

const PollIntervalTicks = 1

const FireDwellMs = 800

const NoValue = interior.NoValue

type GateNode struct {
	Fire           func()
	EmitGeometry   func()
	EmitInputBeads func(left, right int)

	Self *nodeactor.PairNodeSelf

	Clock clock.Clock

	SpeedCh   <-chan float64
	Left      int
	HasLeft   bool
	Right     int
	HasRight  bool
	FromLeft  *beadanimation.Receiver
	FromRight *beadanimation.Receiver
	ToPassed  *beadanimation.Sender
}

const windowTicks = int64(WindowMs / clock.MsPerTick)

const fireDwellTicks = int64(FireDwellMs / clock.MsPerTick)

type gateWindow struct {
	t0         int64
	t0Set      bool
	dwellStart int64
	dwellSet   bool
}

func emitInputs(g *GateNode) {
	l, r := NoValue, NoValue
	if g.HasLeft {
		l = g.Left
	}
	if g.HasRight {
		r = g.Right
	}
	if g.EmitInputBeads != nil {
		g.EmitInputBeads(l, r)
	}
}
