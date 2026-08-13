package owners

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/beadindex"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

type Outs struct {
	outWires       []*wire.PacedWire
	outWireTargets []string
	outWireOuts    []*outport.Out

	outStepsIn    []func(int)
	outBeadRowsIn []func([]wire.LiveBeadRow)
}

func (o *Outs) HasOutWires() bool { return len(o.outWires) > 0 }

// WireTargets returns each distinct target id an outgoing PacedWire drives —
// the animation job's own iteration list, separate from NodeGeometry's
// outTargets (which draws chain beads toward every declared edge target,
// wired or not).
func (o *Outs) WireTargets() []string {
	seen := map[string]bool{}
	var targets []string
	for _, wt := range o.outWireTargets {
		if seen[wt] {
			continue
		}
		seen[wt] = true
		targets = append(targets, wt)
	}
	return targets
}

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

func (o *Outs) AddOutWire(pw *wire.PacedWire, target string, out *outport.Out, sendSteps func(int), sendBeadRows func([]wire.LiveBeadRow)) {
	o.outWires = append(o.outWires, pw)
	o.outWireTargets = append(o.outWireTargets, target)
	o.outWireOuts = append(o.outWireOuts, out)
	o.outStepsIn = append(o.outStepsIn, sendSteps)
	o.outBeadRowsIn = append(o.outBeadRowsIn, sendBeadRows)
}

// SendBeadRows hands each wire's in-flight beads to the edge that owns it.
//
// The beads are read HERE, on the goroutine that steps the wire, because that
// is the only goroutine allowed to look at its in-flight slice. The edge is
// the one that draws them, so what crosses is the finished rows — positions
// already placed along that edge's own segment — and never the wire itself.
func (o *Outs) SendBeadRows(tick int64) {
	for i, pw := range o.outWires {
		if pw == nil || i >= len(o.outBeadRowsIn) || o.outBeadRowsIn[i] == nil {
			continue
		}
		o.outBeadRowsIn[i](pw.LiveBeadRows(tick))
	}
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
