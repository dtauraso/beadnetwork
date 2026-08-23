package pulse

import (
	"context"

	clock "github.com/dtauraso/wirefold/Categories/Clock"
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/nodeapi"
	Speed "github.com/dtauraso/wirefold/Categories/Speed"

	"github.com/dtauraso/wirefold/Categories/NodeKinds/gatecommon"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/helddrive"
	Wiring "github.com/dtauraso/wirefold/Categories/NodeKinds/kindapi"
)

type Pulse struct {
	Fire         func()
	EmitGeometry func()

	EmitHeldBead func(held int)

	Self *Wiring.Self

	Clock clock.Clock

	SpeedCh <-chan float64

	In  *beadanimation.Receiver
	Out Wiring.DrivenOut

	OutFanout Wiring.DrivenOut
}

func driveOutput(out Wiring.DrivenOut) *helddrive.HeldDriver {
	return helddrive.New(out, func(h int64) int { return int(h) })
}

func (g *Pulse) Update(ctx context.Context) {
	nodeapi.TryEmit(g.EmitGeometry)
	g.Self.EmitGeometryOnce()

	var cur int64 = gatecommon.NoValue
	if g.EmitHeldBead != nil {
		g.EmitHeldBead(gatecommon.NoValue)
	}

	drivers := []*helddrive.HeldDriver{driveOutput(g.Out)}

	if g.OutFanout.HasRun() {
		drivers = append(drivers, driveOutput(g.OutFanout))
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
		for _, d := range drivers {
			d.Set(cur)
		}
	}

	clk := g.Clock.Copy()
	g.Self.StartRule(ctx, clk)

	for {
		if ctx.Err() != nil {
			return
		}
		consume()
		Speed.ApplySpeedNonBlocking(clk, g.SpeedCh)
		for _, d := range drivers {
			d.Step(clk.Tick())
		}
		g.Self.Step(ctx, clk.Tick())
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

var Builder = Wiring.BuilderFor("Pulse",
	func(a Wiring.BuildArgs) (nodeapi.Node, error) {
		n := &Pulse{}
		n.Fire = a.Fire()
		n.EmitHeldBead = a.EmitHeldBead()
		n.Clock = a.Clock()
		n.SpeedCh = a.SpeedCh()
		n.In = a.In("In")
		n.Self = a.ClaimSelfDrive()

		n.Out = a.DriveOut("Out")
		n.OutFanout = a.DriveOut("OutFanout")

		return n, nil
	})
