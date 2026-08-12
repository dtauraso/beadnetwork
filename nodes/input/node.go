// node.go holds the Node type, its clock()/broadcastPlace() primitives, the Update
// dispatcher, and RegisterBuilder (this kind's registration must live in this file per
// CLAUDE.md's primitive landing rule). The feedback-ring emit path
// (updateFeedbackRing/feedbackRingSend/feedbackRingReact) lives in feedback_ring.go, and
// the plain periodic-source emit path (runPeriodicEmit/inputCadenceTicks) lives in
// periodic_emit.go — Update dispatches to whichever one applies depending on whether
// FeedbackIn is wired.
package input

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	Wiring "github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
)

type Node struct {
	Fire         func()
	EmitGeometry func()
	// EmitNodeBeads streams the live interior buffer (2x2 grid) as node-bead
	// events — one per present bead. Assigned by this kind's own builder (captures
	// this node's geometry). Called whenever working/backup change so the emitted set
	// always reflects the live arrays. Discrete positions only this phase.
	EmitNodeBeads func(working, backup []int)
	// EmitRefillSlide runs the clock-paced animated refill: the OLD backup (top
	// row) slides DOWN into the working (bottom) row at human speed. Assigned by this
	// kind's own builder; the caller supplies the CLOCK and SPEED CHANNEL at call
	// time (its own already-Copy()'d clock and its own n.SpeedCh — see
	// updateFeedbackRing's n.EmitRefillSlide(clk, n.SpeedCh, *backup) call), so
	// this closure captures only this node's id + geometry, never a clock — see
	// per-goroutine-clock.md's note on the old shape (a captured shared clock read
	// on every call) being a residual to close, not keep. It blocks for the slide
	// duration (pause-aware) and polls its own speed channel each cycle so a speed
	// change mid-slide takes effect immediately rather than waiting for the slide
	// to finish (the slide runs its own blocking loop separate from this node's
	// main loop). nil on test builds without injection — the caller then falls
	// back to the instant refill. beads is the OLD backup contents that become the
	// new working row.
	EmitRefillSlide func(clk clock.Clock, speedCh <-chan float64, beads []int)
	// Clock is this node's OWN clock storage, assigned by this kind's own builder
	// directly from the loader's origin (not derived from any specific wired
	// output port — deriving it from OutCadence/ToExcitatory was fragile:
	// whichever port happened to be wired first controlled pacing, and
	// per-goroutine-clock.md's API demolition removed port-derived clocks
	// entirely anyway). A bare interface field like this is an unguarded
	// nil-interface trap on any construction path that does not set it (a test
	// building &Node{} directly, say): an unguarded `clk.Tick()` panics with no
	// recover over the node goroutine, taking down every other node and the
	// buffer stream with it. The builder below defaults it to clock.NewRealClock()
	// so it is NEVER nil even on a test build with no loader, and overwrites it
	// with the loader's origin clock when there is one; clock() re-guards on
	// every read as a second line of defense in case some future construction
	// path bypasses the builder.
	Clock clock.Clock
	// SpeedCh delivers a speed change to THIS goroutine's own clk copy
	// (per-goroutine-clock.md "Delivery"), assigned by this kind's own builder
	// (injectSpeedChans) with a fresh buffered-1 channel. nil on a test build
	// with no loader — ApplySpeedNonBlocking is then always a no-op.
	SpeedCh <-chan float64
	Init    []int `wire:"data.init"`
	Repeat  bool  `wire:"data.repeat"`
	// OutCadence is a broadcast output like the other two below (n.broadcastPlace
	// gates all three identically on Wired() — none is more "required" than the
	// others; an earlier version of this comment claimed Primary/OutCadence was
	// required and the other two optional, which was FALSE). What sets it apart is
	// narrower: inputCadenceTicks reads OutCadence.Geom() UNCONDITIONALLY (no
	// Wired() check), so this is the one output whose edge length sets this node's
	// own emission cadence — hence the name. Was "ToTime", then "Primary", both
	// kind-leak/false-hierarchy names — the destination does not have to be a
	// Time/TimeStart kind, and this port is not privileged over the other two.
	OutCadence *wire.Out
	// ToExcitatory fans the emitted value out to whatever node samples and holds it
	// (sample-and-hold). It is optional: when unwired (Wired()==false) the emit is
	// skipped so existing topologies without that partner are unaffected.
	ToExcitatory *wire.Out
	FeedbackIn   *wire.In
}

// clock returns n.Clock, guarded against nil (belt-and-suspenders: the
// Register factory below already seeds a real default, but this is the single
// read path every call site goes through so no future construction path can
// reintroduce the bare-nil panic hazard described on the Clock field).
func (n *Node) clock() clock.Clock {
	if n.Clock == nil {
		return clock.NewRealClock()
	}
	return n.Clock
}

// broadcastPlace places v on every wired broadcast output (same cycle — preserves
// concurrent broadcast) without driving them. Returns false only on a
// structural, TERMINAL failure (DriveItem.Failed() — a nil Out), mirroring
// EmitOneDriven's false-return-stops-the-goroutine convention. A momentarily
// full paced-wire buffer (DriveItem.BufferFull()) is TRANSIENT — the wire's own
// driver (its source node's mover) drains it every cycle — so it must NOT stop
// this node's goroutine; that bead is simply dropped from this cycle's
// broadcast (a breadcrumb was already emitted by PacedWire.Send) and the next
// Fire cycle tries again. tick is THIS goroutine's own clock reading, read
// ONCE by the caller and stamped identically on both placements below — that
// single shared reading is what makes whatever is wired to OutCadence and
// whatever is wired to ToExcitatory traverse in lockstep, not a per-Out
// clock read here.
func (n *Node) broadcastPlace(v int, tick int64) bool {
	if n.OutCadence.Wired() && n.OutCadence.PlaceDrivenAt(v, tick).Failed() {
		return false
	}
	if n.ToExcitatory.Wired() && n.ToExcitatory.PlaceDrivenAt(v, tick).Failed() {
		return false
	}
	return true
}

// popEnd, cadenceTicks: see emit_helpers.go — the pure double-buffer/tick-count
// arithmetic this node's loop bodies below call.

func (n *Node) Update(ctx context.Context) {
	wire.TryEmit(n.EmitGeometry)
	if len(n.Init) == 0 {
		return
	}

	// Double-buffer derived from the spec init: working (bottom row) and backup
	// (top row), each a fresh copy of init. The working array IS the emission
	// state — no persistent index. End-popping is the read: end of working is
	// the next value out.
	init := append([]int(nil), n.Init...)
	working := append([]int(nil), init...)
	backup := append([]int(nil), init...)

	// emitBeads streams the live interior buffer as a discrete node-bead snapshot
	// (present beads only). Called on the initial full state and after every array
	// mutation (each pop, each refill) so the emitted set tracks working/backup.
	emitBeads := func() {
		if n.EmitNodeBeads != nil {
			n.EmitNodeBeads(working, backup)
		}
	}
	emitBeads() // initial full(4) state

	// Copy taken ONCE at this goroutine's start (Update IS the goroutine, run
	// once per Input node). Passed down to both branches below instead of each
	// independently calling n.clock() again.
	clk := n.clock().Copy()

	if n.FeedbackIn.Wired() {
		n.updateFeedbackRing(ctx, &working, &backup, init, emitBeads, clk)
		return
	}

	n.runPeriodicEmit(ctx, &working, &backup, init, emitBeads, clk)
}

func init() {
	// Input CONSTRUCTS ITSELF. Every assignment below was previously performed by
	// Wiring.reflectBuild via reflection (matching field NAMES/TYPES, and
	// wire:"data.*" tags for Init/Repeat) — a renamed field now fails to compile
	// instead of silently staying nil/zero.
	Wiring.RegisterBuilder("Input",
		[]portwiring.PortSpec{
			{Name: "OutCadence", Dir: portwiring.PortOut},
			{Name: "ToExcitatory", Dir: portwiring.PortOut},
			{Name: "FeedbackIn", Dir: portwiring.PortIn},
		},
		func(a Wiring.BuildArgs) (wire.Node, error) {
			n := &Node{
				// Clock defaults to a real, live-ticking clock (never nil) so this
				// node is safe even on a construction path with no loader (test
				// builds): a.Clock() below returns nil in that case (BuildArgs.Clock
				// doc: "nil on a test build with no loader"), and this default is
				// what previously came from the Register factory's zero-value seed
				// (`&Node{Clock: clock.NewRealClock()}`) the old registration supplied.
				Clock: clock.NewRealClock(),
			}
			n.Fire = a.Fire()
			n.EmitNodeBeads = a.EmitNodeBeads()
			n.EmitRefillSlide = a.EmitRefillSlide()
			// Only overwrite the constructor default when the loader supplies a
			// real origin clock (a.Clock() is nil on a no-loader test build) —
			// the retired injectClosures only injected Clock when pb.clock !=
			// nil, leaving the bare-field/zero-value default untouched otherwise.
			if clk := a.Clock(); clk != nil {
				n.Clock = clk
			}
			n.SpeedCh = a.SpeedCh()

			// Init/Repeat: wire:"data.init" / wire:"data.repeat" in the old
			// reflection tags. populateData copied slice fields (never aliased)
			// and left the field untouched when the spec supplied no data block
			// or a nil Init slice; reproduced explicitly here via a.Data().
			if data := a.Data(); data != nil {
				if data.Init != nil {
					n.Init = append([]int(nil), data.Init...)
				}
				n.Repeat = data.Repeat
			}

			n.OutCadence = a.Out("OutCadence")
			n.ToExcitatory = a.Out("ToExcitatory")
			n.FeedbackIn = a.In("FeedbackIn")
			// EmitGeometry stays nil deliberately — nodeMover/edgeMover emit the same
			// geometry from their own goroutine start.
			return n, nil
		})
}
