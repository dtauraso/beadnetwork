package gatecommon

import (
	"github.com/dtauraso/wirefold/src/Bead/inport"
	"github.com/dtauraso/wirefold/src/Bead/outport"
	"github.com/dtauraso/wirefold/src/Node/clock"

	"github.com/dtauraso/wirefold/src/Node/Wiring/interior"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeactor"
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
	FromLeft  *inport.In
	FromRight *inport.In
	ToPassed  *outport.Out
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
