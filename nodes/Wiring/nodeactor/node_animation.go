package nodeactor

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/beadindex"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/owners"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

// NodeAnimation is the peer goroutine that owns a node's outgoing
// PacedWires and the human-speed clock that paces their beads. It never
// touches geometry state directly — the only things that cross to/from its
// geometry peer are step counts (in, geometry -> animation) and gathered
// pulses (out, animation -> geometry), both sent non-blocking so neither
// peer ever waits on the other.
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

func (a *NodeAnimation) AddOutWire(pw *wire.PacedWire, target string, o *outport.Out, sendSteps func(int)) {
	a.outs.AddOutWire(pw, target, o, sendSteps)
}

func (a *NodeAnimation) ClearOutWires() {
	a.outs.ClearOutWires()
}

// drainStepCounts applies the most recently received step counts before
// driving wires this cycle — the geometry side's answer to "how many steps
// to the target", non-blocking exactly like DrainPending.
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

// sendPulses gathers this cycle's in-flight bead fractions per target and
// hands them to geometry non-blocking; a slow reader drops the update
// rather than stalling this clock.
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

// Run free-runs on the speed-scaled cycle: apply speed, drive beads, hand
// their fractions to geometry, sleep. It never touches DrainPending or the
// node stream — those belong to the geometry peer alone.
func (a *NodeAnimation) Run(ctx context.Context) {
	a.clocks.CopyClockSrc()
	for {
		a.clocks.ApplySpeed(a.speedCh)
		a.drainStepCounts()

		tick := a.clocks.Tick()
		a.driveOutWires(ctx, tick)
		a.sendPulses(tick)

		if err := a.clocks.SleepCycle(ctx); err != nil {
			return
		}
	}
}
