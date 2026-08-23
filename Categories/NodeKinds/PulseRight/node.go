package pulseright

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

type PulseRight struct {
	Fire         func()
	EmitGeometry func()

	EmitHeldBead func(held int)

	Self *nodeactor.PairNodeSelf

	Clock clock.Clock

	SpeedCh <-chan float64

	In  *beadanimation.Receiver
	Out Wiring.DrivenOut
}

func driveOutput(out Wiring.DrivenOut) *helddrive.HeldDriver {
	return helddrive.New(out, func(h int64) int { return int(h) })
}

func (g *PulseRight) Update(ctx context.Context) {
	nodeapi.TryEmit(g.EmitGeometry)
	g.Self.EmitGeometryOnce()

	var cur int64 = gatecommon.NoValue
	if g.EmitHeldBead != nil {
		g.EmitHeldBead(gatecommon.NoValue)
	}

	driver := driveOutput(g.Out)

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
		driver.Set(cur)
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

	Wiring.RegisterBuilder("PulseRight",
		func(a Wiring.BuildArgs) (nodeapi.Node, error) {
			n := &PulseRight{}
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
