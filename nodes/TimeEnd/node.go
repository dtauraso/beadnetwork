package timeend

import (
	"context"
	"github.com/dtauraso/wirefold/nodes/clock"
	"github.com/dtauraso/wirefold/nodes/nodeapi"
	"github.com/dtauraso/wirefold/nodes/wire/inport"

	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	Wiring "github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
)

const noValue = interior.NoValue

type TimeEnd struct {
	Fire         func()
	EmitGeometry func()
	EmitHeldBead func(held int)
	Held         int `wire:"data.state"`

	Self *nodeactor.PairNodeSelf

	Clock clock.Clock

	SpeedCh <-chan float64
	In      *inport.In
}

func (h *TimeEnd) Update(ctx context.Context) {
	nodeapi.TryEmit(h.EmitGeometry)
	h.Self.EmitGeometryOnce()

	held := noValue
	if h.EmitHeldBead != nil {
		h.EmitHeldBead(held)
	}

	clk := h.Clock.Copy()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		clock.ApplySpeedNonBlocking(clk, h.SpeedCh)
		h.Self.Step(ctx, clk.Tick())
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}

		if value, ok := h.In.PollRecv(); ok {
			if h.Fire != nil {
				h.Fire()
			}
			if value != held && h.EmitHeldBead != nil {
				h.EmitHeldBead(value)
			}
			held = value
			h.Held = value
		}
	}
}

func init() {

	Wiring.RegisterBuilder("TimeEnd",
		[]portwiring.PortSpec{
			{Name: "In", Dir: portwiring.PortIn},
		},
		func(a Wiring.BuildArgs) (nodeapi.Node, error) {
			n := &TimeEnd{

				Held: a.StateSeed("held", noValue),
			}
			n.Fire = a.Fire()
			n.EmitHeldBead = a.EmitHeldBead()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.Self = a.ClaimSelfDrive()
			n.In = a.In("In")

			return n, nil
		})
}
