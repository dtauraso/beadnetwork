package holdflip

import (
	"context"

	clock "github.com/dtauraso/wirefold/Categories/Clock"
	Speed "github.com/dtauraso/wirefold/Categories/Speed"
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/nodeapi"

	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/gatecommon"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/helddrive"
	Wiring "github.com/dtauraso/wirefold/Categories/NodeKinds/kindapi"
)

type Node struct {
	Fire         func()
	EmitGeometry func()

	EmitHeldBead func(held int)

	Self *nodeactor.PairNodeSelf

	Clock clock.Clock

	SpeedCh <-chan float64
	In      *beadanimation.Receiver
	Out     Wiring.DrivenOut
}

func (g *Node) Update(ctx context.Context) {
	nodeapi.TryEmit(g.EmitGeometry)
	g.Self.EmitGeometryOnce()

	if g.EmitHeldBead != nil {
		g.EmitHeldBead(gatecommon.NoValue)
	}

	driver := helddrive.New(g.Out, func(h int64) int {
		if h == gatecommon.NoValue {
			return gatecommon.NoValue
		}
		return 1 - int(h)
	})

	var lastDisplayed int64 = gatecommon.NoValue
	consume := func() {
		v, ok := g.In.PollRecv()
		if !ok {
			return
		}

		for {
			next, ok := g.In.PollRecv()
			if !ok {
				break
			}
			if next != gatecommon.NoValue {
				v = next
			}
		}
		if g.Fire != nil {
			g.Fire()
		}
		newHeld := int64(v)
		driver.Set(newHeld)
		if newHeld != lastDisplayed && g.EmitHeldBead != nil {
			g.EmitHeldBead(v)
		}
		lastDisplayed = newHeld
	}

	clk := g.Clock.Copy()
	g.Self.StartRule(ctx, clk)

	for {
		if ctx.Err() != nil {
			return
		}
		consume()
		Speed.ApplySpeedNonBlocking(clk, g.SpeedCh)
		driver.Step(clk.Tick())
		g.Self.Step(ctx, clk.Tick())
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

func init() {

	Wiring.RegisterBuilder("HoldFlip",
		func(a Wiring.BuildArgs) (nodeapi.Node, error) {
			n := &Node{}
			n.Fire = a.Fire()
			n.EmitHeldBead = a.EmitHeldBead()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.In = a.In("In")
			n.Self = a.ClaimSelfDrive()

			n.Out = a.DriveOut("Out")

			return n, nil
		})
}
