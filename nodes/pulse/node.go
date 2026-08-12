package pulse

import (
	"context"
	"github.com/dtauraso/wirefold/nodes/clock"
	"github.com/dtauraso/wirefold/nodes/nodeapi"
	"github.com/dtauraso/wirefold/nodes/wire/inport"

	Wiring "github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/gatecommon"
)

type Pulse struct {
	Fire         func()
	EmitGeometry func()

	EmitHeldBead func(held int)

	Clock clock.Clock

	SpeedCh          <-chan float64
	Out1SpeedCh      <-chan float64
	OutFanoutSpeedCh <-chan float64

	In  *inport.In
	Out Wiring.DrivenOut

	OutFanout Wiring.DrivenOut
}

func driveOutput(ctx context.Context, out Wiring.DrivenOut, heldCh <-chan int64, clk clock.Clock, speedCh <-chan float64) {
	gatecommon.DriveHeld(ctx, out, heldCh, func(h int64) int { return int(h) }, clk, speedCh)
}

func (g *Pulse) Update(ctx context.Context) {
	nodeapi.TryEmit(g.EmitGeometry)

	var cur int64 = gatecommon.NoValue
	if g.EmitHeldBead != nil {
		g.EmitHeldBead(gatecommon.NoValue)
	}

	out1HeldCh := make(chan int64, 1)
	outFanoutHeldCh := make(chan int64, 1)

	driveOutput(ctx, g.Out, out1HeldCh, g.Clock, g.Out1SpeedCh)

	if g.OutFanout.Wired() {
		driveOutput(ctx, g.OutFanout, outFanoutHeldCh, g.Clock, g.OutFanoutSpeedCh)
	}

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
		clock.SendLatestNonBlocking(outFanoutHeldCh, cur)
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

	Wiring.RegisterBuilder("Pulse",
		[]portwiring.PortSpec{
			{Name: "In", Dir: portwiring.PortIn},
			{Name: "Out", Dir: portwiring.PortOut},
			{Name: "OutFanout", Dir: portwiring.PortOut},
		},
		func(a Wiring.BuildArgs) (nodeapi.Node, error) {
			n := &Pulse{}
			n.Fire = a.Fire()
			n.EmitHeldBead = a.EmitHeldBead()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.Out1SpeedCh = a.SpeedCh()
			n.OutFanoutSpeedCh = a.SpeedCh()
			n.In = a.In("In")

			n.Out = a.DriveOut("Out", 0)
			n.OutFanout = a.DriveOut("OutFanout", 1)

			return n, nil
		})
}
