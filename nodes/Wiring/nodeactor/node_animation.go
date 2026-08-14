package nodeactor

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/beadindex"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/owners"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

type NodeAnimation struct {
	id string

	outs   owners.Outs
	clocks owners.Clocks

	speedCh chan float64

	stepIn   <-chan map[string]int
	pulseOut chan<- map[string][]beadindex.Pulse
}

func (a *NodeAnimation) SetSpeedCh(ch chan float64) {
	a.speedCh = ch
}

func (a *NodeAnimation) AddOutWire(pw *wire.PacedWire, target string, o *outport.Out, sendSteps func(int), sendBeadRows func([]wire.LiveBeadRow)) {
	a.outs.AddOutWire(pw, target, o, sendSteps, sendBeadRows)
}

func (a *NodeAnimation) ClearOutWires() {
	a.outs.ClearOutWires()
}

func (a *NodeAnimation) drainStepCounts() {
	for {
		select {
		case counts := <-a.stepIn:
			for to, count := range counts {
				a.outs.PublishStepCount(to, count)
			}
		default:
			return
		}
	}
}

func (a *NodeAnimation) sendPulses(tick int64) {
	if !a.outs.HasOutWires() {
		return
	}
	targets := a.outs.WireTargets()
	pulses := make(map[string][]beadindex.Pulse, len(targets))
	for _, to := range targets {
		pulses[to] = a.outs.GatherPulses(to, tick)
	}
	select {
	case a.pulseOut <- pulses:
	default:
	}
}

func (a *NodeAnimation) driveOutWires(ctx context.Context, tick int64) {
	a.outs.DriveOutWires(ctx, tick)
}

func (a *NodeAnimation) Run(ctx context.Context) {
	a.clocks.CopyClockSrc()
	for {
		a.clocks.ApplySpeed(a.speedCh)
		a.drainStepCounts()

		tick := a.clocks.Tick()
		a.driveOutWires(ctx, tick)
		a.sendPulses(tick)
		a.outs.SendBeadRows(tick)

		if err := a.clocks.SleepCycle(ctx); err != nil {
			return
		}
	}
}
