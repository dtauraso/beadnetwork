// Package Node2 is the "Node2" kind — the mirror of Node1: one half of the
// straightening-loop pair. It is REACTIVE, not periodic: every cycle it drains its own In
// non-blockingly and, on an arrival, runs the straightening rule ITSELF — step this
// node's OWN tilt vector one click toward perpendicular to its coplanar normal and, if it
// moved, place a bead back out on its OWN Out; if it is already perpendicular, do nothing
// and send nothing, which is how the exchange terminates. This all runs on THIS
// goroutine: there is no round trip to the mover to decide (see topTiltVectorThetaIdx below
// for who else the index is reported to and why).
//
// Emission is otherwise silent: with no In arrival there is nothing to react to, and the
// loop is kicked off by a USER tilting a node via the TiltVectorAnglePanel — routed here
// via its own dedicated TiltEditIn channel (BuildArgs.TiltEditIn), also drained
// non-blockingly every cycle. A panel edit unconditionally applies its ±1 click and always
// places a bead on Out ("THE KICK"): with both nodes of a pair perpendicular nothing
// circulates on In, correctly, since there is nothing left to straighten, so the loop has
// no way to start on its own — it is kicked off by the thing that actually moves a tilt
// away from perpendicular. The same click is also the vector channel's opening move — it
// sends this node's own outgoing vector alongside the bead (applyTiltEdit), which is what
// gives handleVectorCycle something to reply to; a channel whose only sends are replies
// never carries anything at all. Pairing a Node1 and a Node2 with one edge each direction
// (Node1.Out → Node2.In, Node2.Out → Node1.In) needs no seed/bootstrap node: nothing ever
// sends until a user tilt starts it, so there is no deadlock to bootstrap out of at t=0.
//
// The RESET button (TiltResetButton.tsx) also arrives on TiltEditIn (TiltEditMsg.Reset),
// but is the opposite of a panel click: it places NO bead — a stop-and-return, not a nudge,
// so it never starts the straightening exchange the way a panel click does. It does more
// than zero the indices, because zeroed indices are not by themselves a stopped exchange:
// it runs this node's full clear() (below), which also empties the bead edge — the thing
// that has actually been turning these tilts — so nothing is left in the pair that could
// land a moment later and step it back off zero.
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
	// TopTiltThetaIdx/TopTiltPhiIdx are THIS node's OWN vector-direction indices — the ONE
	// writer, full stop (memory/feedback_abc_times_constant_not_rederive.md: index ×
	// step-constant, trig only at the cartesian/polar boundary). Seeded once at build
	// time from the persisted value (BuildArgs.TiltVectorAngleSeed) and mutated ONLY by
	// this goroutine's own Update loop, below. Every change is reported one-way to this
	// node's own mover (SyncTiltIndex) so the mover — which still owns streaming this
	// node's geometry and persisting it to this node's own position.json — stays in
	// sync; the mover never decides or mutates these itself for this kind.
	TopTiltThetaIdx, TopTiltPhiIdx int32
	// TiltEditIn is this node's dedicated channel for a panel-driven tilt-angle click
	// (TiltVectorAnglePanel), claimed at build time via BuildArgs.TiltEditIn — see the
	// package doc comment's "THE KICK".
	TiltEditIn <-chan Wiring.TiltEditMsg
	// SyncTiltIndex notifies this node's own mover of the current TopTiltThetaIdx/TopTiltPhiIdx
	// AND the current coplanar-normal indices (coplanarNormal, below) — one-way,
	// fire-and-forget, never an ack (BuildArgs.SyncTiltIndex).
	SyncTiltIndex func(theta, phi, normalTheta, normalPhi, bottomTheta, bottomPhi int32)
	// VectorOut/VectorIn are THIS node's own ends of its dedicated tilt-vector channel
	// (Wiring.TiltVectorMsg — an integer θ/φ index pair, never floats on a channel),
	// claimed at build time via BuildArgs.VectorOut/VectorIn. It travels ALONGSIDE the
	// ordinary bead edge (In/Out above), never replacing it — beads are unaffected.
	// Buffered depth 1, latest-wins, non-blocking on both ends
	// (Wiring.SendVectorLatestNonBlocking / Wiring.PollRecvVector). nil when this
	// node's edge partner did not also ask for one, or on a bare test build with no
	// loader — both helpers already treat nil as "nothing wired".
	VectorOut chan<- Wiring.TiltVectorMsg
	VectorIn  <-chan Wiring.TiltVectorMsg
	// ReceivedThetaIdx/ReceivedPhiIdx/ReceivedSet are THIS node's own record of the LAST
	// direction that ARRIVED on VectorIn — the third drawn arrow (user request: "show a
	// 3rd vector...the last iteration of it as a different color in the node that
	// received it"). Written ONLY by this goroutine, in handleVectorCycle below: an
	// arrival REPLACES whatever was here before (never accumulates), and it persists
	// indefinitely otherwise — it is NOT cleared when the straightening exchange settles.
	// It IS cleared by a RESET, both this node's own (applyTiltEdit's Reset branch) and a
	// Reset marker arriving on VectorIn (handleVectorCycle's Reset branch): a reset is a
	// stop-and-return, and a stale received arrow left hanging would contradict that.
	// Reported one-way to this node's own mover via SyncReceivedVector, same shape as
	// TopTiltThetaIdx/SyncTiltIndex above.
	ReceivedThetaIdx, ReceivedPhiIdx int32
	ReceivedSet                      bool
	// SyncReceivedVector notifies this node's own mover of the current
	// ReceivedThetaIdx/PhiIdx/Set — one-way, fire-and-forget, never an ack
	// (BuildArgs.SyncReceivedVector).
	SyncReceivedVector func(theta, phi int32, set bool)
	// ClearOutBeads asks THIS node's own mover to drop every bead still crossing this
	// node's outgoing wires — one-way, fire-and-forget, never an ack
	// (BuildArgs.ClearOutBeads). Called only from clear(), below: those beads are owned
	// by the mover (it drives the wires), so this node asks rather than reaching in.
	ClearOutBeads func()
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
// on why Node1/Node2 stay distinct packages). Steps TopTiltThetaIdx ONE click toward
// Wiring.PerpendicularThetaIdx and reports true when it moved; a false return with no
// mutation is the loop's termination, not a missed case.
func (n *Node) stepTilt() bool {
	// Already perpendicular: the exchange ends here — no step, and the caller sends
	// nothing, which is what stops the beads circulating.
	if n.TopTiltThetaIdx == Wiring.PerpendicularThetaIdx {
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
	n.TopTiltThetaIdx += 1
	return true
}

// applyTiltEdit applies one panel-driven edit (TiltVectorAnglePanel's ±1 click, or the
// RESET button's TiltResetButton.tsx) directly to this node's OWN indices — same
// no-mover-round-trip shape as stepTilt. Reports whether the caller should place "THE
// KICK" bead: true for an adjust (unconditional, whichever side of perpendicular it lands
// on), false for a reset — a reset is a stop-and-return, not a nudge, so nothing should
// start circulating from it (package doc comment's "THE KICK").
//
// An adjust ALSO opens the vector exchange, by sending this node's own outgoingVector on
// VectorOut. This is the vector channel's whole starting move, and without it the channel
// never carries a direction at all: handleVectorCycle only ever sends in REPLY to an
// arrival, so with nothing to reply to, no node ever received one, no node ever set its
// received-direction record, and the third arrow could not be drawn anywhere. The panel
// click is the right place for it and not an extra decision: it is already the one thing
// that moves a tilt away from perpendicular, and already the one place that starts the
// bead half of the exchange. One user click now opens both channels at once, which is why
// they stay in step rather than one running a cycle ahead of the other.
func (n *Node) applyTiltEdit(edit Wiring.TiltEditMsg) (placeBead bool) {
	if edit.Reset {
		n.clear()
		// Tell the partner, so it clears too — see clear's own doc comment for why the
		// partner's clear, not this one, is what actually ends the exchange.
		Wiring.SendVectorLatestNonBlocking(n.VectorOut, Wiring.TiltVectorMsg{Reset: true})
		return false
	}
	delta := int32(-1)
	if edit.Up {
		delta = 1
	}
	if edit.Axis == "phi" {
		n.TopTiltPhiIdx += delta
	} else {
		n.TopTiltThetaIdx += delta
	}
	// Open the vector exchange — see this function's own doc comment. Sent AFTER the
	// index moved, so the partner gets the direction this click produced, not the one it
	// replaced.
	Wiring.SendVectorLatestNonBlocking(n.VectorOut, n.outgoingVector())
	return true
}

// clear returns THIS node to its opening state and — the part that matters — leaves
// nothing behind that could restart the straightening exchange (shared shape with Node1's
// copy, kept duplicated per-package for the same reason stepTilt is). A reset is not "set
// the indices to 0"; it is "there is no message anywhere in the pair", and zeroed indices
// are just what that looks like from outside. Everything the pair holds between clicks is
// cleared here, each piece by the goroutine that owns it:
//
//   - this node's own tilt and derived coplanar-normal indices (owned here);
//   - this node's record of the last received direction, the third drawn arrow (owned here);
//   - this node's VectorIn, drained non-blocking — the receive end is owned here, and a
//     direction already sitting in it would arrive on the next cycle and step the tilt
//     straight back off zero. Depth-1 latest-wins, so one receive empties it;
//   - this node's already-DELIVERED beads, drained off In the same way and for the same
//     reason — the bead edge, not the vector channel, is what has actually been turning
//     these tilts (In.PollRecv -> stepTilt -> Out.PlaceDrivenAt), so a reset that skips it
//     visibly does not take;
//   - this node's OUTGOING beads, still crossing. Those are NOT owned here: a PacedWire is
//     driven by its source node's own MOVER, so this asks the mover to drop them
//     (ClearOutBeads / Wiring.moveMsgKindBeadClear) rather than reaching into the wire.
//
// WHY BOTH CALLERS EXIST. The RESET button sends one record per node, but the two nodes
// act on their own goroutines at their own moments, so a single clear each is racy: the
// partner can place one more bead in the window after this node cleared, and it lands
// afterwards and restarts everything. What closes that window is the Reset MARKER — this
// node clears, sends the marker, and places nothing ever again from that path; the partner
// runs this same clear when the marker arrives, which is therefore ordered after the last
// thing this node could have placed. So each node clears twice, and the second one is the
// one that provably lands last. The marker gets no reply (handleVectorCycle), so it stops
// there instead of bouncing.
func (n *Node) clear() {
	n.TopTiltThetaIdx = 0
	n.TopTiltPhiIdx = 0
	n.syncTiltIndex()
	n.ReceivedThetaIdx = 0
	n.ReceivedPhiIdx = 0
	n.ReceivedSet = false
	n.syncReceivedVector()
	Wiring.PollRecvVector(n.VectorIn)
	n.drainIn()
	if n.ClearOutBeads != nil {
		n.ClearOutBeads()
	}
}

// drainIn empties this node's own In of every bead already delivered to it, on this
// node's own goroutine (In.PollRecv is non-blocking, so this terminates as soon as the
// queue is empty). Bounded by what the partner placed before it stopped placing, which
// is the pair's own bead-per-cycle traffic — the same drain-until-empty shape as
// PacedWire.drainPlacements, whose doc comment carries the full reasoning for why these
// loops need no cap.
func (n *Node) drainIn() {
	if n.In == nil {
		return
	}
	for {
		if _, ok := n.In.PollRecv(); !ok {
			return
		}
	}
}

// bottomTilt is THIS node's own BOTTOM TILT VECTOR: a half turn (180°,
// Wiring.HalfTurnThetaIdx steps) in θ from its OWN top tilt vector, so it points out of the
// node's other side and turns with the top as the top turns — index arithmetic only, never
// trig (memory/feedback_abc_times_constant_not_rederive.md). φ is left unchanged: a half
// turn in θ alone already negates the direction exactly, so no companion φ+π is needed (see
// Wiring.HalfTurnThetaIdx's own doc comment).
//
// Node2 SUBTRACTS the half turn (its mirror package does the opposite), the same
// opposite-senses convention outgoingVector already uses. Both signs land in the SAME drawn
// direction — ±180° in θ is the same place — so this is index bookkeeping, not geometry:
// each kind's indices keep walking in its own direction instead of one kind's jumping the
// other way at the turn.
func (n *Node) bottomTilt() Wiring.TiltVectorMsg {
	return Wiring.TiltVectorMsg{
		ThetaIdx: n.TopTiltThetaIdx - Wiring.HalfTurnThetaIdx,
		PhiIdx:   n.TopTiltPhiIdx,
	}
}

// coplanarNormal is THIS node's own coplanar normal: a quarter turn (90°, 6 steps of
// Wiring.CurveParamTiltVectorAngleStep — Wiring.PerpendicularThetaIdx names the same
// 90°-worth-of-steps magnitude) from THIS node's OWN tilt vector, so the normal stays
// perpendicular to the tilt as the tilt turns — index arithmetic only, never trig
// (memory/feedback_abc_times_constant_not_rederive.md). φ is left unchanged: the turn
// is entirely in θ, same in-ring-plane assumption this scene's stepTilt already relies
// on (see Wiring.PerpendicularThetaIdx's doc comment).
func (n *Node) coplanarNormal() Wiring.TiltVectorMsg {
	// Node2 SUBTRACTS the quarter turn (Node1's mirror package adds it) — the tilt
	// vector and the coplanar normal sit −90° apart in θ for Node2, +90° for Node1,
	// keeping the sign with the KIND, same as stepTilt's own add/subtract split.
	return Wiring.TiltVectorMsg{
		ThetaIdx: n.TopTiltThetaIdx - Wiring.PerpendicularThetaIdx,
		PhiIdx:   n.TopTiltPhiIdx,
	}
}

// syncTiltIndex reports THIS node's current tilt index AND its current coplanar-normal
// index (coplanarNormal above) to this node's own mover in one call — every call site
// that changes TopTiltThetaIdx/TopTiltPhiIdx must also report the normal, since the normal is
// derived from the tilt and the mover no longer derives it itself (see
// Wiring.moveMsgKindTiltIndexSync's doc comment). nil-safe, same as every other closure
// call here.
func (n *Node) syncTiltIndex() {
	if n.SyncTiltIndex == nil {
		return
	}
	norm := n.coplanarNormal()
	bottom := n.bottomTilt()
	n.SyncTiltIndex(n.TopTiltThetaIdx, n.TopTiltPhiIdx, norm.ThetaIdx, norm.PhiIdx, bottom.ThetaIdx, bottom.PhiIdx)
}

// syncReceivedVector reports THIS node's current received-vector state (ReceivedThetaIdx/
// PhiIdx/Set) to this node's own mover — the third-arrow twin of syncTiltIndex. Called by
// every site that changes those fields, below. nil-safe, same as syncTiltIndex.
func (n *Node) syncReceivedVector() {
	if n.SyncReceivedVector == nil {
		return
	}
	n.SyncReceivedVector(n.ReceivedThetaIdx, n.ReceivedPhiIdx, n.ReceivedSet)
}

// outgoingVector is what THIS node SENDS on VectorOut: its own coplanarNormal rotated
// 180° in θ. Node2 turns +180° (+12 steps of π/12); Node1 (its mirror package) turns
// −180° (−12 steps) — index arithmetic, never radians. φ is untouched.
func (n *Node) outgoingVector() Wiring.TiltVectorMsg {
	norm := n.coplanarNormal()
	norm.ThetaIdx += 2 * Wiring.PerpendicularThetaIdx
	return norm
}

// stepTowardPerpendicularFromVector is the vector-channel twin of stepTilt, and the ONE
// place the arrived direction's own value is consulted rather than just its arrival. The
// rule is two DOT PRODUCTS: the arrived vector against this node's own TOP tilt vector, and
// against its own BOTTOM tilt vector. If AT LEAST ONE of them is the acute-angle case, this
// node moves its own tilt by its own fixed one-step "little bit" — Node2 adds,
// its mirror package does the opposite, so a pair turns symmetrically rather than both
// chasing the same way. If NEITHER is acute, nothing steps and the caller sends nothing,
// which is how the exchange stops.
//
// Note what this means with the bottom tilt defined as a half turn from the top: the bottom
// is the top's exact antipode, so the two dots are always exact negatives of each other and
// "at least one is acute" is false ONLY when the arrived vector is exactly perpendicular to
// the tilt axis. The two-dot form is still what is written here, because it is the rule as
// stated and it keeps saying the right thing if the bottom ever stops being a straight
// half turn — but it is not currently a wider gate than a single dot would be.
func (n *Node) stepTowardPerpendicularFromVector(received Wiring.TiltVectorMsg) bool {
	if !Wiring.TiltVectorIsAcute(received, n.topTilt()) && !Wiring.TiltVectorIsAcute(received, n.bottomTilt()) {
		return false
	}
	return n.stepTilt()
}

// topTilt is THIS node's own top tilt vector as a direction pair — the same indices held on
// the struct, named so the dot products above read as vector-against-vector rather than as
// two loose ints.
func (n *Node) topTilt() Wiring.TiltVectorMsg {
	return Wiring.TiltVectorMsg{ThetaIdx: n.TopTiltThetaIdx, PhiIdx: n.TopTiltPhiIdx}
}

// handleVectorCycle is Node2's WHOLE per-cycle vector-channel loop body: read
// VectorIn non-blocking; if something arrived, decide (stepTowardPerpendicularFromVector);
// and if that moved this node's own tilt, send outgoingVector back out on VectorOut —
// also non-blocking. At the perpendicular index nothing steps and nothing sends,
// which is how the exchange stops. This never touches In/Out or beads — the vector
// channel is a separate, additive exchange.
func (n *Node) handleVectorCycle() {
	received, ok := Wiring.PollRecvVector(n.VectorIn)
	if !ok {
		return
	}
	// A RESET marker is not a direction to act on: zero this node's own tilt to match its
	// partner and REPLY WITH NOTHING. Replying would bounce the reset back and forth
	// forever; the marker's job is to stop the exchange, so it ends here. Same
	// stop-and-return reasoning as applyTiltEdit's Reset branch, and this is the clear
	// that actually makes the pair quiescent — see clear's own doc comment on why the
	// marker-driven one, not the button-driven one, is the one that lands last.
	if received.Reset {
		n.clear()
		return
	}
	// A real direction. It is recorded UNCONDITIONALLY — before, and independently of, the
	// step decision below — and then STAYS until the next arrival replaces it. It does not
	// vanish when the exchange settles: the last direction a node was sent is what it is
	// still holding, and blanking the arrow the moment the pair stops turning would erase
	// the very state the pair came to rest in. The only thing that removes it is a RESET
	// (clear, above — this node's own or its partner's marker), which removes it because a
	// reset means there is nothing in the pair at all any more.
	n.ReceivedThetaIdx = received.ThetaIdx
	n.ReceivedPhiIdx = received.PhiIdx
	n.ReceivedSet = true
	n.syncReceivedVector()
	if !n.stepTowardPerpendicularFromVector(received) {
		return
	}
	n.syncTiltIndex()
	Wiring.SendVectorLatestNonBlocking(n.VectorOut, n.outgoingVector())
}

func (n *Node) Update(ctx context.Context) {
	wire.TryEmit(n.EmitGeometry)

	// Report THIS node's OPENING tilt/normal pair once, before the loop — same reason as
	// Node1's: the mover mirrors these and cannot derive the normal itself, so without
	// this its normal indices stay at zero until the first arrival or panel click, the two
	// drawn arrows superimpose on world +y, and the coplanar normal reads as missing.
	n.syncTiltIndex()

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
				n.syncTiltIndex()
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
				placeBead := n.applyTiltEdit(edit)
				n.syncTiltIndex()
				if placeBead && n.Out != nil {
					n.Out.PlaceDrivenAt(1, clk.Tick())
				}
			default:
			}
		}

		// Vector-channel exchange: a separate, additive loop body from the bead
		// exchange above — see handleVectorCycle's own doc comment.
		n.handleVectorCycle()

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
			n.TopTiltThetaIdx, n.TopTiltPhiIdx = a.TiltVectorAngleSeed()
			n.TiltEditIn = a.TiltEditIn()
			n.SyncTiltIndex = a.SyncTiltIndex()
			n.SyncReceivedVector = a.SyncReceivedVector()
			n.ClearOutBeads = a.ClearOutBeads()
			n.VectorOut = a.VectorOut()
			n.VectorIn = a.VectorIn()
			// EmitGeometry stays nil deliberately — nodeMover/edgeMover emit the
			// same geometry from their own goroutine start.
			return n, nil
		})
}
