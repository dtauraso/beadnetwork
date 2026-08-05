// Package Node2 is the "Node2" kind — the mirror of Node1: one half of the
// straightening-loop pair. It is REACTIVE, not periodic: every cycle it drains its own In
// non-blockingly and, on an arrival, runs the straightening rule ITSELF — step this
// node's OWN tilt vector one click toward perpendicular to its coplanar normal and, if it
// moved, place a bead back out on its OWN Out; if it is already perpendicular, do nothing
// and send nothing, which is how the exchange terminates. This all runs on THIS
// goroutine: there is no round trip to the mover to decide (see tiltVectorThetaIdx below
// for who else the index is reported to and why).
//
// Emission is otherwise silent: with no In arrival there is nothing to react to, and the
// loop is kicked off by a USER tilting a node via the TiltVectorAnglePanel — routed here
// via its own dedicated TiltEditIn channel (BuildArgs.TiltEditIn), also drained
// non-blockingly every cycle. A panel edit unconditionally applies its ±1 click and always
// places a bead on Out ("THE KICK"): with both nodes of a pair perpendicular nothing
// circulates on In, correctly, since there is nothing left to straighten, so the loop has
// no way to start on its own — it is kicked off by the thing that actually moves a tilt
// away from perpendicular. Pairing a Node1 and a Node2 with one edge each direction
// (Node1.Out → Node2.In, Node2.Out → Node1.In) needs no seed/bootstrap node: nothing ever
// sends until a user tilt starts it, so there is no deadlock to bootstrap out of at t=0.
//
// Kept as a distinct package/kind rather than parametrizing Node1, per the check-dep-rules
// guard: a node-kind package may import only the shared spine, never a sibling kind, so
// Node1 and Node2 cannot share a struct across packages without a third shared package
// neither currently needs.
package Node2

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

type Node struct {
	Fire         func()
	EmitGeometry func()
	// Clock is this node's OWN clock storage, assigned by this kind's own
	// builder directly from the loader's origin (per-goroutine-clock.md).
	// Update() Copies it once for its own loop.
	Clock wire.Clock
	// SpeedCh delivers a speed change to this goroutine's own clk copy.
	// Assigned by this kind's own builder; nil on a test build with no loader.
	SpeedCh <-chan float64
	// In is one of two triggers that drive this node — see the package doc comment.
	In *wire.In
	// Out is the sole output. THIS goroutine is now the SOLE placer on it (below),
	// preserving wire.Out.PlaceDrivenAt's one-goroutine-per-Out invariant — the mover no
	// longer places on this Out at all.
	Out *wire.Out
	// TiltThetaIdx/TiltPhiIdx are THIS node's OWN vector-direction indices — the ONE
	// writer, full stop (memory/feedback_abc_times_constant_not_rederive.md: index ×
	// step-constant, trig only at the cartesian/polar boundary). Seeded once at build
	// time from the persisted value (BuildArgs.TiltVectorAngleSeed) and mutated ONLY by
	// this goroutine's own Update loop, below. Every change is reported one-way to this
	// node's own mover (SyncTiltIndex) so the mover — which still owns streaming this
	// node's geometry and persisting it to this node's own position.json — stays in
	// sync; the mover never decides or mutates these itself for this kind.
	TiltThetaIdx, TiltPhiIdx int32
	// TiltEditIn is this node's dedicated channel for a panel-driven tilt-angle click
	// (TiltVectorAnglePanel), claimed at build time via BuildArgs.TiltEditIn — see the
	// package doc comment's "THE KICK".
	TiltEditIn <-chan Wiring.TiltEditMsg
	// SyncTiltIndex notifies this node's own mover of the current TiltThetaIdx/TiltPhiIdx
	// — one-way, fire-and-forget, never an ack (BuildArgs.SyncTiltIndex).
	SyncTiltIndex func(theta, phi int32)
}

func (n *Node) clock() wire.Clock {
	if n.Clock == nil {
		return wire.NewRealClock()
	}
	return n.Clock
}

// stepTowardPerpendicular is the straightening rule (shared shape with Node1's copy — kept
// duplicated per-package rather than factored into Wiring, since a node-kind package may
// import only the shared spine, never a sibling kind — see this package's own doc comment
// on why Node1/Node2 stay distinct packages). Steps TiltThetaIdx ONE click toward
// Wiring.PerpendicularThetaIdx and reports true when it moved; a false return with no
// mutation is the loop's termination, not a missed case.
func (n *Node) stepTilt() bool {
	// Already perpendicular: the exchange ends here — no step, and the caller sends
	// nothing, which is what stops the beads circulating.
	if n.TiltThetaIdx == Wiring.PerpendicularThetaIdx {
		return false
	}
	// Node2 adds. The two kinds move their tilt in OPPOSITE senses by the same one
	// step of π/12 — Node2 adds where the other does the reverse — so a pair turns
	// symmetrically rather than both chasing the same direction.
	//
	// The step is a fixed direction, NOT "toward perpendicular": this node subtracts (or
	// adds) whatever side of perpendicular it starts on. That is what was asked for, and it
	// means the exchange only terminates when the walk happens to land exactly on
	// PerpendicularThetaIdx — from the far side it walks away and keeps going. The bead
	// paces it, so "keeps going" is a slow visible turn rather than a spin.
	n.TiltThetaIdx += 1
	return true
}

func (n *Node) Update(ctx context.Context) {
	wire.TryEmit(n.EmitGeometry)

	// Copy taken ONCE at this goroutine's start (Update IS the goroutine).
	clk := n.clock().Copy()

	for {
		if ctx.Err() != nil {
			return
		}

		// Drain In non-blocking: an arrival is the straightening loop's reactive
		// trigger. Step ONE click toward perpendicular; if it moved, sync the mover
		// and place the outgoing bead ourselves — no round trip to any other
		// goroutine to decide.
		if _, ok := n.In.PollRecv(); ok {
			if n.Fire != nil {
				n.Fire()
			}
			if n.stepTilt() {
				if n.SyncTiltIndex != nil {
					n.SyncTiltIndex(n.TiltThetaIdx, n.TiltPhiIdx)
				}
				if n.Out != nil {
					n.Out.PlaceDrivenAt(1, clk.Tick())
				}
			}
		}

		// Drain TiltEditIn non-blocking: a panel-driven click. Unconditional ±1 on
		// the named axis (never a step-toward-perpendicular decision — the user asked
		// for exactly this move), sync, and unconditionally place "THE KICK" bead —
		// see the package doc comment for why this send is never conditional.
		if n.TiltEditIn != nil {
			select {
			case edit := <-n.TiltEditIn:
				delta := int32(-1)
				if edit.Up {
					delta = 1
				}
				if edit.Axis == "phi" {
					n.TiltPhiIdx += delta
				} else {
					n.TiltThetaIdx += delta
				}
				if n.SyncTiltIndex != nil {
					n.SyncTiltIndex(n.TiltThetaIdx, n.TiltPhiIdx)
				}
				if n.Out != nil {
					n.Out.PlaceDrivenAt(1, clk.Tick())
				}
			default:
			}
		}

		wire.ApplySpeedNonBlocking(clk, n.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

func init() {
	// Node2 CONSTRUCTS ITSELF (Wiring.RegisterBuilder), same self-construction
	// shape as every other kind.
	Wiring.RegisterBuilder("Node2",
		[]Wiring.PortSpec{
			{Name: "In", Dir: Wiring.PortIn},
			{Name: "Out", Dir: Wiring.PortOut},
		},
		func(a Wiring.BuildArgs) (wire.Node, error) {
			n := &Node{
				Clock: wire.NewRealClock(),
			}
			n.Fire = a.Fire()
			if clk := a.Clock(); clk != nil {
				n.Clock = clk
			}
			n.SpeedCh = a.SpeedCh()
			n.In = a.In("In")
			n.Out = a.Out("Out")
			n.TiltThetaIdx, n.TiltPhiIdx = a.TiltVectorAngleSeed()
			n.TiltEditIn = a.TiltEditIn()
			n.SyncTiltIndex = a.SyncTiltIndex()
			// EmitGeometry stays nil deliberately — nodeMover/edgeMover emit the
			// same geometry from their own goroutine start.
			return n, nil
		})
}
