package pacer

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
	"github.com/dtauraso/wirefold/nodes/wire/outport"

	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	Wiring "github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
)

const noValue = interior.NoValue

type Node struct {
	Fire         func()
	EmitGeometry func()
	EmitHeldBead func(held int)
	Held         int `wire:"data.state"`

	Clock clock.Clock

	SpeedCh <-chan float64

	In          *wire.In
	FeedbackOut *outport.Out
}

func (p *Node) Update(ctx context.Context) {
	wire.TryEmit(p.EmitGeometry)

	held := noValue
	if p.EmitHeldBead != nil {
		p.EmitHeldBead(held)
	}

	clk := p.Clock.Copy()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		clock.ApplySpeedNonBlocking(clk, p.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}

		if value, ok := p.In.PollRecv(); ok {
			if p.Fire != nil {
				p.Fire()
			}

			heldChanged := value != held
			held = value
			if heldChanged && p.EmitHeldBead != nil {
				p.EmitHeldBead(value)
			}

			step := 0
			if heldChanged {
				step = 1
			}
			p.Held = value

			p.FeedbackOut.PlaceDrivenAt(step, clk.Tick())
		}
	}
}

func init() {

	Wiring.RegisterBuilder("Pacer",
		[]portwiring.PortSpec{
			{Name: "In", Dir: portwiring.PortIn},
			{Name: "FeedbackOut", Dir: portwiring.PortOut},
		},
		func(a Wiring.BuildArgs) (wire.Node, error) {
			n := &Node{

				Held: a.StateSeed("held", noValue),
			}
			n.Fire = a.Fire()
			n.EmitHeldBead = a.EmitHeldBead()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.In = a.In("In")
			n.FeedbackOut = a.Out("FeedbackOut")

			return n, nil
		})
}
