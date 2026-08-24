package selectleft

import (
	"context"

)

func runGateLoop(ctx context.Context, g *GateNode, captureLeftFn, captureRightFn func(*GateNode) bool, fireResult func(*GateNode) int) {
	if g.EmitGeometry != nil {
		g.EmitGeometry()
	}
	g.Self.EmitGeometryOnce()

	if g.Clock == nil {
		panic("runGateLoop: gate node has no clock — a self-driven node steps its own geometry from its own clock, so there is no wall-clock fallback to fall back to")
	}

	clk := g.Clock.Copy()
	clk.SpeedFrom(g.SpeedCh)
	g.Self.StartRule(ctx, clk)
	now := clk.Tick

	sleep := func(ctx context.Context) error {
		g.Self.Step(ctx, clk.Tick())
		return clk.SleepCycle(ctx)
	}

	var w gateWindow
	emitInputs(g)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if sleep(ctx) != nil {
			return
		}

		if captureLeftFn(g) {
			emitInputs(g)
		}
		if captureRightFn(g) {
			emitInputs(g)
		}

		openWindowIfNeeded(g, &w, now)

		fired := tryFireOnDwell(g, &w, now, fireResult)

		if !fired && w.t0Set && !(g.HasLeft && g.HasRight) && now()-w.t0 > windowTicks {
			clearWindow(g, &w)
			emitInputs(g)
		}
	}
}

func RunGate(ctx context.Context, g *GateNode, invertLeft bool) {
	runGateLoop(ctx, g,
		func(g *GateNode) bool { return captureLeft(g, invertLeft) },
		func(g *GateNode) bool { return captureRight(g, invertLeft) },
		func(g *GateNode) int {
			if g.Left == 1 && g.Right == 1 {
				return 1
			}
			return 0
		},
	)
}

func RunGateAccept(ctx context.Context, g *GateNode, expectLeft, expectRight int) {
	runGateLoop(ctx, g,
		captureRawLeft,
		captureRawRight,
		func(g *GateNode) int {
			if g.Left == expectLeft && g.Right == expectRight {
				return 1
			}
			return 0
		},
	)
}
