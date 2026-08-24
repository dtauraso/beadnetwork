package timeend

import (
	"context"

	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"

	interior "github.com/dtauraso/beadnetwork/Categories/Node/Interior"
)

const noValue = interior.NoValue

type TimeEnd struct {
	Fire         func()
	EmitGeometry func()
	EmitHeldBead func(held int)
	Held         int `wire:"data.state"`

	Self *Self

	Clock clock.Clock

	SpeedCh <-chan float64
	In      *beadanimation.Receiver
}

func (h *TimeEnd) Update(ctx context.Context) {
	tryEmit(h.EmitGeometry)
	h.Self.EmitGeometryOnce()

	held := noValue
	if h.EmitHeldBead != nil {
		h.EmitHeldBead(held)
	}

	clk := h.Clock.Copy()
	clk.SpeedFrom(h.SpeedCh)
	h.Self.StartRule(ctx, clk)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

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

var Builder = BuilderFor("TimeEnd",
	func(a BuildArgs) (any, error) {
		n := &TimeEnd{

			Held: a.StateSeed("held", noValue),
		}
		n.Fire = a.Fire()
		n.EmitHeldBead = a.EmitHeldBead()
		n.Clock = a.Clock()
		n.SpeedCh = a.SpeedCh()
		n.Self = claimSelfDrive(a)
		n.In = a.In("In")

		return n, nil
	})
