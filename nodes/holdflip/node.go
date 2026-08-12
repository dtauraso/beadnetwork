package holdflip

import (
	"context"
	"github.com/dtauraso/wirefold/nodes/nodeapi"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
	"github.com/dtauraso/wirefold/nodes/wire/inport"

	Wiring "github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/gatecommon"
)

type Node struct {
	Fire         func()
	EmitGeometry func()

	EmitHeldBead func(held int)

	Clock clock.Clock

	SpeedCh      <-chan float64
	DriveSpeedCh <-chan float64
	In           *inport.In
	Out          Wiring.DrivenOut
}

func (g *Node) Update(ctx context.Context) {
	nodeapi.TryEmit(g.EmitGeometry)

	if g.EmitHeldBead != nil {
		g.EmitHeldBead(gatecommon.NoValue)
	}
	heldCh := make(chan int64, 1)

	gatecommon.DriveHeld(ctx, g.Out, heldCh, func(h int64) int {
		if h == gatecommon.NoValue {
			return gatecommon.NoValue
		}
		return 1 - int(h)
	}, g.Clock, g.DriveSpeedCh)

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
		clock.SendLatestNonBlocking(heldCh, newHeld)
		if newHeld != lastDisplayed && g.EmitHeldBead != nil {
			g.EmitHeldBead(v)
		}
		lastDisplayed = newHeld
	}

	clk := g.Clock.Copy()

	for {
		if ctx.Err() != nil {
			return
		}
		consume()
		clock.ApplySpeedNonBlocking(clk, g.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

func init() {

	Wiring.RegisterBuilder("HoldFlip",
		[]portwiring.PortSpec{
			{Name: "In", Dir: portwiring.PortIn},
			{Name: "Out", Dir: portwiring.PortOut},
		},
		func(a Wiring.BuildArgs) (nodeapi.Node, error) {
			n := &Node{}
			n.Fire = a.Fire()
			n.EmitHeldBead = a.EmitHeldBead()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.DriveSpeedCh = a.SpeedCh()
			n.In = a.In("In")

			n.Out = a.DriveOut("Out", 0)

			return n, nil
		})
}
