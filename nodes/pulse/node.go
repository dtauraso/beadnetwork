package pulse

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
	"github.com/dtauraso/wirefold/nodes/gatecommon"
)

// Pulse is the "Pulse" kind (registered as "Pulse" below — the name lives here in
// the comment, describing what its functions do). Its functions: it is a
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
// held is owned by the MAIN loop; each drive goroutine gets its OWN channel
// (Out1HeldCh/OutFanoutHeldCh) the main loop sends the latest held value on
// (wire.SendLatestNonBlocking) whenever it changes — the same per-goroutine-channel
// shape as SpeedCh/Out1SpeedCh/OutFanoutSpeedCh below, so two DriveHeld goroutines never
// steal values from each other. The output is NOT precondition-gated: it self-emits
// noValue from the start (like the Input bootstrap), never inert until fed.
type Pulse struct {
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
	// Out1SpeedCh/OutFanoutSpeedCh do the same for each DriveHeld goroutine's OWN
	// independent copy (per-goroutine-clock.md "Delivery") — three separate
	// clock-owning goroutines here need three separate channels, since sharing
	// one across goroutines would silently starve whichever one loses a given
	// receive. Assigned by this kind's own builder via a.SpeedCh(); nil on a test
	// build with no loader.
	SpeedCh          <-chan float64
	Out1SpeedCh      <-chan float64
	OutFanoutSpeedCh <-chan float64
	// In is the sole input: a sampled value that updates the held value (rule 4 —
	// with exactly one input, there is nothing to distinguish it from).
	In  *wire.In
	Out Wiring.DrivenOut
	// OutFanout is an optional SECOND continuous output driving the same held value, so a
	// Pulse can fan to two destinations (e.g. node 6 → node 5 via Out and → node 11
	// via OutFanout). Optional: when unwired (Wired()==false, e.g. node 7) its drive
	// goroutine is skipped, so single-output Pulse nodes are unaffected. Named for its
	// job, not "Out2" — a number says nothing about what distinguishes it from Out
	// (nothing does, functionally; only Out is driven unconditionally while this one
	// is optional — see Update).
	OutFanout Wiring.DrivenOut
}

// driveOutput runs a continuous-drive goroutine on out, always emitting the
// current value of held. Delegates to gatecommon.DriveHeld (shared with
// HoldFlip's identical-shaped drive goroutine) with an identity transform.
func driveOutput(ctx context.Context, out Wiring.DrivenOut, heldCh <-chan int64, clk wire.Clock, speedCh <-chan float64) {
	gatecommon.DriveHeld(ctx, out, heldCh, func(h int64) int { return int(h) }, clk, speedCh)
}

func (g *Pulse) Update(ctx context.Context) {
	wire.TryEmit(g.EmitGeometry)

	// held is owned by this main loop; cur is the main loop's OWN local copy
	// (seeded to gatecommon.NoValue, same as held).
	var cur int64 = gatecommon.NoValue
	if g.EmitHeldBead != nil {
		g.EmitHeldBead(gatecommon.NoValue) // startup: empty interior
	}

	// Each drive goroutine gets its OWN buffered-1, latest-wins channel — a
	// single channel cannot serve two receivers without one stealing values
	// from the other (see the doc comment on Out1SpeedCh/OutFanoutSpeedCh).
	out1HeldCh := make(chan int64, 1)
	outFanoutHeldCh := make(chan int64, 1)

	// DRIVE goroutine: continuously pulse the current held value to Out. g.Clock is
	// the ORIGIN clock; DriveHeld Copies it independently at its own goroutine's start
	// — never hand a copy to a second goroutine.
	driveOutput(ctx, g.Out, out1HeldCh, g.Clock, g.Out1SpeedCh)

	// Optional SECOND drive goroutine for OutFanout.
	if g.OutFanout.Wired() {
		driveOutput(ctx, g.OutFanout, outFanoutHeldCh, g.Clock, g.OutFanoutSpeedCh)
	}

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
		wire.SendLatestNonBlocking(outFanoutHeldCh, cur)
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
	// Pulse CONSTRUCTS ITSELF. Every assignment below was previously performed by
	// Wiring.reflectBuild via field-name/type reflection; a rename now fails to compile.
	Wiring.RegisterBuilder("Pulse",
		[]Wiring.PortSpec{
			{Name: "In", Dir: Wiring.PortIn},
			{Name: "Out", Dir: Wiring.PortOut},
			{Name: "OutFanout", Dir: Wiring.PortOut},
		},
		func(a Wiring.BuildArgs) (wire.Node, error) {
			n := &Pulse{}
			n.Fire = a.Fire()
			n.EmitHeldBead = a.EmitHeldBead()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.Out1SpeedCh = a.SpeedCh()
			n.OutFanoutSpeedCh = a.SpeedCh()
			n.In = a.In("In")
			// DriveOut, not Out: both Out and OutFanout are driven by their OWN
			// gatecommon.DriveHeld goroutine below (driveOutput calls), a SEPARATE
			// goroutine from this node's own Update loop — see DriveOut's doc
			// comment and docs/interior-stream-framing.md. Distinct slots (0, 1):
			// two DriveHeld goroutines on one node must never share a stream.
			n.Out = a.DriveOut("Out", 0)
			n.OutFanout = a.DriveOut("OutFanout", 1)
			// EmitGeometry stays nil deliberately — nodeMover/edgeMover emit the same
			// geometry from their own goroutine start (see builders.go's note).
			return n, nil
		})
}
