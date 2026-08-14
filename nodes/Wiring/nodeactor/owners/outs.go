package owners

import (
	"context"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

type Outs struct {
	outWires []*wire.PacedWire

	outBeadRowsIn []func([]wire.LiveBeadRow)
}

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

func (o *Outs) AddOutWire(pw *wire.PacedWire, sendBeadRows func([]wire.LiveBeadRow)) {
	o.outWires = append(o.outWires, pw)
	o.outBeadRowsIn = append(o.outBeadRowsIn, sendBeadRows)
}

func (o *Outs) SendBeadRows(tick int64) {
	for i, pw := range o.outWires {
		if pw == nil || i >= len(o.outBeadRowsIn) || o.outBeadRowsIn[i] == nil {
			continue
		}
		o.outBeadRowsIn[i](pw.LiveBeadRows(tick))
	}
}
