package gatecommon

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// runGateLoop is the shared window/dwell/clock loop body used by both RunGate
// (invert + AND-11) and RunGateAccept (raw capture + direct pattern match). The
// ONLY things that vary between callers are how each side is captured (inversion
// or not) and what result the dwell fires — everything else (window-open,
// window-timeout clear, dwell timing, breadcrumbs, clock/speed handling) is
// identical and lives here exactly once.
func runGateLoop(ctx context.Context, g *GateNode, captureLeftFn, captureRightFn func(*GateNode) bool, fireResult func(*GateNode) int) {
	if g.EmitGeometry != nil {
		g.EmitGeometry()
	}

	// Copy taken ONCE at this goroutine's start (RunGate IS the goroutine, run
	// once per gate node). This copy backs both now() and sleep() whenever the loader provided one
	// (g.Clock != nil). g.Clock is this node's own clock storage (seeded by
	// the kind's own builder from the loader's origin); ports no longer hand out a clock
	// (API demolition item 1), so this replaces the old g.ToPassed.Clock().Copy().
	// g.Tick/defaultTick are kept only as the no-loader fallback for now() (unit
	// tests with no loader), matching prior behavior there. The window/dwell
	// timing that governs the gate's own interior-bead animation is speed-aware
	// regardless of whether this gate happens to have a live out-wire in this
	// topology — a gate with an unconnected ToPassed still owns a real Clock
	// copy and SpeedCh (requested unconditionally by the kind's builder whenever a
	// loader is present).
	var now func() int64
	sleep := defaultSleep()
	if g.Clock != nil {
		clk := g.Clock.Copy()
		now = clk.Tick
		// Fold the speed-delivery poll into the one blocking point this loop
		// has (per-goroutine-clock.md "Delivery" — DriveHeld's sibling note
		// applies equally here: RunGate's only blocking point is this sleep).
		sleep = func(ctx context.Context) error {
			clock.ApplySpeedNonBlocking(clk, g.SpeedCh)
			return clk.SleepCycle(ctx)
		}
	} else if g.Tick != nil {
		now = g.Tick
	} else {
		now = defaultTick()
	}

	var w gateWindow
	emitInputs(g) // initial empty interior

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Sleep BEFORE this cycle's observe/fire work.
		if sleep(ctx) != nil {
			return
		}

		// Each side tracks the MOST-RECENT real bead: drain to the latest value
		// (discarding NoValue placeholders) and update the slot even if already
		// held. NoValue never fills a slot.
		if captureLeftFn(g) {
			emitInputs(g)
		}
		if captureRightFn(g) {
			emitInputs(g)
		}

		openWindowIfNeeded(g, &w, now)

		fired := tryFireOnDwell(g, &w, now, fireResult)

		// A partial combination has been open longer than W → clear it. Only
		// time out while still waiting for the second input; once both are held
		// we are committed to firing after the dwell. Skipped on the cycle we
		// just fired (mirrors the old `continue`).
		if !fired && w.t0Set && !(g.HasLeft && g.HasRight) && now()-w.t0 > windowTicks {
			clearWindow(g, &w)
			emitInputs(g)
		}
	}
}

// RunGate runs the shared SelectRight/SelectLeft gate loop.
// invertLeft=true  → the LEFT input is NOT-inverted on capture  (SelectRight).
// invertLeft=false → the RIGHT input is NOT-inverted on capture (SelectLeft).
// Fires 1 iff the STORED (post-inversion) values are Left==1 AND Right==1.
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

// RunGateAccept runs the shared SelectRight/SelectLeft gate loop with NO inversion on
// capture (raw FromLeft/FromRight values are stored as-is — no NOT gates) and
// fires 1 iff the raw stored values DIRECTLY match the given pattern
// (Left==expectLeft && Right==expectRight). This expresses an acceptance pattern
// (e.g. "01") without going through an invert-then-AND-11 detour.
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
