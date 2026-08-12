package pulseright

import (
	"context"
	"github.com/dtauraso/wirefold/nodes/nodeapi"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
	"github.com/dtauraso/wirefold/nodes/wire/inport"

	Wiring "github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/gatecommon"
)

type PulseRight struct {
	Fire         func()
	EmitGeometry func()

	EmitHeldBead func(held int)

	Clock clock.Clock

	SpeedCh     <-chan float64
	Out1SpeedCh <-chan float64

	In  *inport.In
	Out Wiring.DrivenOut
}

func driveOutput(ctx context.Context, out Wiring.DrivenOut, heldCh <-chan int64, clk clock.Clock, speedCh <-chan float64) {
	gatecommon.DriveHeld(ctx, out, heldCh, func(h int64) int { return int(h) }, clk, speedCh)
}

func (g *PulseRight) Update(ctx context.Context) {
	nodeapi.TryEmit(g.EmitGeometry)

	var cur int64 = gatecommon.NoValue
	if g.EmitHeldBead != nil {
		g.EmitHeldBead(gatecommon.NoValue)
	}

	out1HeldCh := make(chan int64, 1)

	driveOutput(ctx, g.Out, out1HeldCh, g.Clock, g.Out1SpeedCh)

	consume := func() {
		v, ok := g.In.PollRecv()
		if !ok {
			return
		}
		if g.Fire != nil {
			g.Fire()
		}
		if int64(v) != cur && g.EmitHeldBead != nil {
			g.EmitHeldBead(v)
		}
		cur = int64(v)
		clock.SendLatestNonBlocking(out1HeldCh, cur)
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

	Wiring.RegisterBuilder("PulseRight",
		[]portwiring.PortSpec{
			{Name: "In", Dir: portwiring.PortIn},
			{Name: "Out", Dir: portwiring.PortOut},
		},
		func(a Wiring.BuildArgs) (nodeapi.Node, error) {
			n := &PulseRight{}
			n.Fire = a.Fire()
			n.EmitHeldBead = a.EmitHeldBead()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.Out1SpeedCh = a.SpeedCh()
			n.In = a.In("In")

			n.Out = a.DriveOut("Out", 0)

			return n, nil
		})
}
