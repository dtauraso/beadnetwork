package pacer

import (
	"context"

	clock "github.com/dtauraso/wirefold/src/Clock"
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/NodeKinds/nodeapi"

	interior "github.com/dtauraso/wirefold/src/Node/Interior"
	Wiring "github.com/dtauraso/wirefold/src/NodeKinds/kindapi"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeactor"
)

const noValue = interior.NoValue

type Node struct {
	Fire         func()
	EmitGeometry func()
	EmitHeldBead func(held int)
	Held         int `wire:"data.state"`

	Self *nodeactor.PairNodeSelf

	Clock clock.Clock

	SpeedCh <-chan float64

	In          *beadanimation.Receiver
	FeedbackOut *beadanimation.Sender
}

func (p *Node) Update(ctx context.Context) {
	nodeapi.TryEmit(p.EmitGeometry)
	p.Self.EmitGeometryOnce()

	held := noValue
	if p.EmitHeldBead != nil {
		p.EmitHeldBead(held)
	}

	clk := p.Clock.Copy()
	p.Self.StartRule(ctx, clk)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		clock.ApplySpeedNonBlocking(clk, p.SpeedCh)
		p.Self.Step(ctx, clk.Tick())
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
		func(a Wiring.BuildArgs) (nodeapi.Node, error) {
			n := &Node{

				Held: a.StateSeed("held", noValue),
			}
			n.Fire = a.Fire()
			n.EmitHeldBead = a.EmitHeldBead()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.Self = a.ClaimSelfDrive()
			n.In = a.In("In")
			n.FeedbackOut = a.Out("FeedbackOut")

			return n, nil
		})
}
