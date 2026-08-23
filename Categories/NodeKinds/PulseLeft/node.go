package pulseleft

import (
	"context"
	interior "github.com/dtauraso/wirefold/Categories/Node/Interior"

	clock "github.com/dtauraso/wirefold/Categories/Clock"
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/portwiring"
	Speed "github.com/dtauraso/wirefold/Categories/Speed"
)

type PulseLeft struct {
	Fire         func()
	EmitGeometry func()

	EmitHeldBead func(held int)

	Self *Self

	Clock clock.Clock

	SpeedCh <-chan float64

	In  *beadanimation.Receiver
	Out DrivenOut
}

func driveOutput(out DrivenOut) *HeldDriver {
	return newHeldDriver(out, func(h int64) int { return int(h) })
}

func (g *PulseLeft) Update(ctx context.Context) {
	tryEmit(g.EmitGeometry)
	g.Self.EmitGeometryOnce()

	var cur int64 = interior.NoValue
	if g.EmitHeldBead != nil {
		g.EmitHeldBead(interior.NoValue)
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

var Builder = BuilderFor("PulseLeft",
	func(a BuildArgs) (portwiring.Node, error) {
		n := &PulseLeft{}
		n.Fire = a.Fire()
		n.EmitHeldBead = a.EmitHeldBead()
		n.Clock = a.Clock()
		n.SpeedCh = a.SpeedCh()
		n.In = a.In("In")
		n.Self = claimSelfDrive(a)

		n.Out = a.DriveOut("Out")

		return n, nil
	})
