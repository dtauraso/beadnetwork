// Package gatecommon holds the shared constants and gate-loop body used by
// SelectRight and SelectLeft. Each of those node
// packages is its own package (primitive landing rule) but delegates its
// Update body here, parameterised by which side is NOT-inverted on capture.
package gatecommon

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
)

// WindowMs is the target coincidence window expressed in milliseconds. This is a
// design choice calibrated to exceed the same-cycle input skew (~69 ms measured)
// while staying under the input cadence (~3104 ms).
const WindowMs = 3000

// PollIntervalTicks bounds the busy-spin of the window loop. It is a free
// scheduling choice (not derivable from pulse speed or fire-dwell) that trades
// CPU burn against reaction latency between window polls. One tick is the finest
// grain of the human-speed clock (MsPerTick ms ≈ the old 5 ms poll rounds up to 1).
const PollIntervalTicks = 1

// FireDwellMs holds both inputs visible (interior beads present) for this long
// once both are held, before the gate fires + clears. Without it the
// second-arriving interior bead only flashes for ~1ms before the fire clears it.
const FireDwellMs = 800

// NoValue aliases interior.NoValue, the sentinel meaning "no value yet" / "no real
// bead". Defined in package interior (not here) because gatecommon imports interior, not
// the reverse — interior.NoValue is the one definition; this is just gatecommon's name
// for it.
const NoValue = interior.NoValue

// GateNode holds all the fields shared between the two gate node kinds.
// Each kind embeds GateNode so its init/Update can delegate here.
type GateNode struct {
	Fire           func()
	EmitGeometry   func()
	EmitInputBeads func(left, right int)
	// Tick is a fallback "now" used ONLY when this node has no Clock copy at all
	// (a test build with no loader — Clock is nil in that case). It reads the
	// loader's ORIGIN clock, which per-goroutine-clock.md nothing ever applies a
	// speed change to (only per-goroutine copies receive speed sinks), so it is
	// deaf to the slider. RunGate must NOT fall back to this whenever a Clock
	// copy is available, even if the gate's output happens to be unwired in this
	// topology — that was the bug (a gate with no out-wire ran its window/dwell
	// timing, and therefore its interior-bead flicker, at a frozen speed
	// regardless of the slider).
	Tick func() int64
	// Clock is this node's OWN clock storage, assigned by the kind's own builder from the
	// loader's origin (builders.go injectClosures, bare-field injection matched by
	// exact type clock.Clock — see input.Node.Clock for the model this mirrors).
	// RunGate Copies it exactly ONCE at its own goroutine's start; ports no
	// longer carry or hand out a clock (API demolition item 1), so this is the
	// only path in.
	// nil on a test build with no loader — RunGate falls back to Tick/wall-clock
	// sleep in that case, exactly as before.
	Clock clock.Clock
	// SpeedCh delivers a speed change to RunGate's own clock copy
	// (per-goroutine-clock.md "Delivery"), assigned by this kind's own builder
	// (injectSpeedChans). nil on a test build with no loader / chan mode.
	SpeedCh   <-chan float64
	Left      int
	HasLeft   bool
	Right     int
	HasRight  bool
	FromLeft  *wire.In
	FromRight *wire.In
	ToPassed  *wire.Out
}

// windowTicks is the fixed coincidence window as a tick count (WindowMs / MsPerTick).
const windowTicks = int64(WindowMs / clock.MsPerTick)

// fireDwellTicks is FireDwellMs converted to a tick count.
const fireDwellTicks = int64(FireDwellMs / clock.MsPerTick)

// gateWindow holds the window/dwell timing state for one RunGate loop instance.
// It is local to a single call (not part of GateNode) since it is pure loop-scoped
// bookkeeping, not node-shared state.
type gateWindow struct {
	t0         int64
	t0Set      bool
	dwellStart int64
	dwellSet   bool
}

// emitInputs reports the currently-held interior bead values (NoValue where a side
// isn't held yet).
func emitInputs(g *GateNode) {
	l, r := NoValue, NoValue
	if g.HasLeft {
		l = g.Left
	}
	if g.HasRight {
		r = g.Right
	}
	if g.EmitInputBeads != nil {
		g.EmitInputBeads(l, r)
	}
}
