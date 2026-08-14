package nodeactor

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/owners"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

type NodeAnimation struct {
	id string

	outs   owners.Outs
	clocks owners.Clocks

	speedCh chan float64
}

func (a *NodeAnimation) SetSpeedCh(ch chan float64) {
	a.speedCh = ch
}

func (a *NodeAnimation) AddOutWire(pw *wire.PacedWire, sendBeadRows func([]wire.LiveBeadRow)) {
	a.outs.AddOutWire(pw, sendBeadRows)
}

func (a *NodeAnimation) ClearOutWires() {
	a.outs.ClearOutWires()
}

func (a *NodeAnimation) driveOutWires(ctx context.Context, tick int64) {
	a.outs.DriveOutWires(ctx, tick)
}

func (a *NodeAnimation) Run(ctx context.Context) {
	a.clocks.CopyClockSrc()
	for {
		a.clocks.ApplySpeed(a.speedCh)

		tick := a.clocks.Tick()
		a.driveOutWires(ctx, tick)
		a.outs.SendBeadRows(tick)

		if err := a.clocks.SleepCycle(ctx); err != nil {
			return
		}
	}
}
