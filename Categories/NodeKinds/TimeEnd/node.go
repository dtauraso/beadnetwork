package timeend

import (
	"context"

	clock "github.com/dtauraso/wirefold/Categories/Clock"
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/nodeapi"
	Speed "github.com/dtauraso/wirefold/Categories/Speed"

	interior "github.com/dtauraso/wirefold/Categories/Node/Interior"
	Wiring "github.com/dtauraso/wirefold/Categories/NodeKinds/kindapi"
)

const noValue = interior.NoValue

type TimeEnd struct {
	Fire         func()
	EmitGeometry func()
	EmitHeldBead func(held int)
	Held         int `wire:"data.state"`

	Self *Wiring.Self

	Clock clock.Clock

	SpeedCh <-chan float64
	In      *beadanimation.Receiver
}

func (h *TimeEnd) Update(ctx context.Context) {
	nodeapi.TryEmit(h.EmitGeometry)
	h.Self.EmitGeometryOnce()

	held := noValue
	if h.EmitHeldBead != nil {
		h.EmitHeldBead(held)
	}

	clk := h.Clock.Copy()
	h.Self.StartRule(ctx, clk)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		Speed.ApplySpeedNonBlocking(clk, h.SpeedCh)
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

var Builder = Wiring.BuilderFor("TimeEnd",
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
