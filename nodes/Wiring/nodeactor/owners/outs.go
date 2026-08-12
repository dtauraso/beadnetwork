package owners

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/beadindex"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

type Outs struct {
	outTargets     []string
	outWires       []*wire.PacedWire
	outWireTargets []string
	outWireOuts    []*outport.Out

	outStepsIn []func(int)
}

func (o *Outs) AddOutTarget(target string) {
	o.outTargets = append(o.outTargets, target)
}

func (o *Outs) OutTargets() []string { return o.outTargets }

func (o *Outs) HasOutWires() bool { return len(o.outWires) > 0 }

func (o *Outs) DriveOutWires(ctx context.Context, tick int64) {
	for _, pw := range o.outWires {
		pw.DriveOneCycle(ctx, tick)
	}
}

func (o *Outs) ClearOutWires() {
	for _, pw := range o.outWires {
		pw.ClearInFlight()
	}
}

func (o *Outs) AddOutWire(pw *wire.PacedWire, target string, out *outport.Out, sendSteps func(int)) {
	o.outWires = append(o.outWires, pw)
	o.outWireTargets = append(o.outWireTargets, target)
	o.outWireOuts = append(o.outWireOuts, out)
	o.outStepsIn = append(o.outStepsIn, sendSteps)
}

func (o *Outs) PublishStepCount(to string, count int) {
	for i, wt := range o.outWireTargets {
		if wt != to {
			continue
		}
		if i < len(o.outWireOuts) && o.outWireOuts[i] != nil {
			o.outWireOuts[i].PublishSteps(count)
		}
		if i < len(o.outStepsIn) && o.outStepsIn[i] != nil {
			o.outStepsIn[i](count)
		}
	}
}

func (o *Outs) GatherPulses(to string, tick int64) []beadindex.Pulse {
	var pulses []beadindex.Pulse
	for i, wt := range o.outWireTargets {
		if wt != to || o.outWires[i] == nil {
			continue
		}
		for _, p := range o.outWires[i].LiveBeadFractions(tick) {
			if p.T < 0 || p.T >= 1 || p.Steps <= 0 {
				continue
			}

			pulses = append(pulses, beadindex.Pulse{T: p.T, Steps: p.Steps, Val: int32(p.Val)})
		}
	}
	return pulses
}
