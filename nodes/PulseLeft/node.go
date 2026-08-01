package pulseleft

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
	"github.com/dtauraso/wirefold/nodes/gatecommon"
)

// PulseLeft is the "PulseLeft" kind (registered as "PulseLeft" below — the name
// lives here in the comment, describing what its functions do). It is a clone
// of the Pulse kind (see nodes/pulse) — identical behavior for now, split out
// so node 3's future divergence has a home of its own. Its functions: it is a
// sample-and-hold that HOLDS one int value (initialized to noValue) and drives it
// out continuously.
//   - The MAIN loop (Update) runs one activity cycle per human clock tick: a
//     non-blocking input check (PollRecv) on In, and on a new value it
//     calls EmitHeldBead and stores the new held, then sleeps one cycle
//     (clk.SleepCycle) — parking the CPU instead of spinning while idle.
//   - A DRIVE goroutine continuously pulses the CURRENT held value to the output
//     via gatecommon.DriveHeld (PlaceDriven + per-cycle StepOnce, sleeping one
//     cycle between steps), self-pacing at the wire rate and re-reading held each
//     pulse — when held changes the next pulse carries the new value.
//   - EmitGeometry publishes this node's geometry.
//
// held is owned by the MAIN loop; the drive goroutine gets its OWN channel
// (Out1HeldCh) the main loop sends the latest held value on
// (wire.SendLatestNonBlocking) whenever it changes — the same per-goroutine-channel
// shape as SpeedCh/Out1SpeedCh below. The output is NOT precondition-gated: it
// self-emits noValue from the start (like the Input bootstrap), never inert
// until fed.
type PulseLeft struct {
	Fire         func()
	EmitGeometry func()
	// EmitHeldBead, assigned by this kind's own builder, streams the held value as a
	// SINGLE centered interior node-bead (present when held != noValue). Re-emitted at
	// startup (held = noValue, empty interior) and whenever the held value changes.
	EmitHeldBead func(held int)
	// Clock is this node's OWN clock storage, assigned by this kind's own builder
	// directly from the loader's origin (bare-field injection by exact type
	// wire.Clock — see input.Node.Clock; ports no longer hand out a clock,
	// per-goroutine-clock.md API demolition item 1). Update() Copies it once
	// for its own loop, and passes the ORIGIN (not that copy) to each DRIVE
	// goroutine below, which Copies independently at ITS OWN start.
	Clock wire.Clock
	// SpeedCh delivers a speed change to the MAIN loop's own clock copy;
	// Out1SpeedCh does the same for the DriveHeld goroutine's OWN independent
	// copy (per-goroutine-clock.md "Delivery") — two separate clock-owning
	// goroutines here need two separate channels, since sharing one across
	// goroutines would silently starve whichever one loses a given receive.
	// Assigned by this kind's own builder via a.SpeedCh(); nil on a test build
	// with no loader.
	SpeedCh     <-chan float64
	Out1SpeedCh <-chan float64
	// In is the sole input: a sampled value that updates the held value (rule 4 —
	// with exactly one input, there is nothing to distinguish it from).
	In  *wire.In
	Out Wiring.DrivenOut
}

// driveOutput runs a continuous-drive goroutine on out, always emitting the
// current value of held. Delegates to gatecommon.DriveHeld (shared with
// HoldFlip's identical-shaped drive goroutine) with an identity transform.
func driveOutput(ctx context.Context, out Wiring.DrivenOut, heldCh <-chan int64, clk wire.Clock, speedCh <-chan float64) {
	gatecommon.DriveHeld(ctx, out, heldCh, func(h int64) int { return int(h) }, clk, speedCh)
}

func (g *PulseLeft) Update(ctx context.Context) {
	wire.TryEmit(g.EmitGeometry)

	// held is owned by this main loop; cur is the main loop's OWN local copy
	// (seeded to gatecommon.NoValue, same as held).
	var cur int64 = gatecommon.NoValue
	if g.EmitHeldBead != nil {
		g.EmitHeldBead(gatecommon.NoValue) // startup: empty interior
	}

	// The drive goroutine gets its OWN buffered-1, latest-wins channel (see the
	// doc comment on Out1SpeedCh).
	out1HeldCh := make(chan int64, 1)

	// DRIVE goroutine: continuously pulse the current held value to Out. g.Clock is
	// the ORIGIN clock; DriveHeld Copies it independently at its own goroutine's start
	// — never hand a copy to a second goroutine.
	driveOutput(ctx, g.Out, out1HeldCh, g.Clock, g.Out1SpeedCh)

	// MAIN loop frame: do activities (non-blocking input check + update held),
	// then sleep one human clock cycle, repeat. The drive goroutine picks up the
	// new held on its next pulse. Sleeping one cycle per iteration (paced mode)
	// keeps the loop off the CPU 99% of the time instead of spinning millions of
	// times per human tick while there is nothing to receive.
	consume := func() {
		v, ok := g.In.PollRecv()
		if !ok {
			return
		}
		if g.Fire != nil {
			g.Fire()
		}
		if int64(v) != cur && g.EmitHeldBead != nil {
			g.EmitHeldBead(v) // show the new interior bead IMMEDIATELY
		}
		cur = int64(v)
		wire.SendLatestNonBlocking(out1HeldCh, cur)
	}

	// Copy taken ONCE at this goroutine's start (Update IS the goroutine); each
	// DRIVE goroutine above takes its own copy independently inside
	// gatecommon.DriveHeld.
	clk := g.Clock.Copy()

	// Paced mode: do activities, sleep one human clock cycle, repeat.
	for {
		if ctx.Err() != nil {
			return
		}
		consume()
		wire.ApplySpeedNonBlocking(clk, g.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

func init() {
	// PulseLeft CONSTRUCTS ITSELF. Every assignment below was previously performed by
	// Wiring.reflectBuild via field-name/type reflection; a rename now fails to compile.
	Wiring.RegisterBuilder("PulseLeft",
		[]Wiring.PortSpec{
			{Name: "In", Dir: Wiring.PortIn},
			{Name: "Out", Dir: Wiring.PortOut},
		},
		func(a Wiring.BuildArgs) (wire.Node, error) {
			n := &PulseLeft{}
			n.Fire = a.Fire()
			n.EmitHeldBead = a.EmitHeldBead()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.Out1SpeedCh = a.SpeedCh()
			n.In = a.In("In")
			// DriveOut, not Out: Out is driven by its own gatecommon.DriveHeld
			// goroutine, a SEPARATE goroutine from this node's own Update loop —
			// see BuildArgs.DriveOut's doc comment and docs/interior-stream-
			// framing.md.
			n.Out = a.DriveOut("Out", 0)
			// EmitGeometry stays nil deliberately — nodeMover/edgeMover emit the same
			// geometry from their own goroutine start (see builders.go's note).
			return n, nil
		})
}
