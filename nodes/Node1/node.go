// Package Node1 is the "Node1" kind: one half of the straightening-loop pair
// (Node1/Node2). It is REACTIVE, not periodic: every cycle it drains its own In
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
package Node1

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

type Node struct {
	Fire         func()
	EmitGeometry func()
	// Clock is this node's OWN clock storage, assigned by this kind's own
	// builder directly from the loader's origin (per-goroutine-clock.md; see
	// input.Node.Clock for the fuller rationale). Update() Copies it once for
	// its own loop — the sole clock-owning goroutine this node has.
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

// applyTiltEdit applies one panel-driven edit (TiltVectorAnglePanel's ±1 click, or the
// RESET button's TiltResetButton.tsx) directly to this node's OWN indices — same
// no-mover-round-trip shape as stepFromVector. Reports whether the caller should place "THE
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
// nothing behind that could restart the straightening exchange. A reset is not "set the
// indices to 0"; it is "there is no message anywhere in the pair", and zeroed indices are
// just what that looks like from outside. Everything the pair holds between clicks is
// cleared here, each piece by the goroutine that owns it:
//
//   - this node's own tilt and derived coplanar-normal indices (owned here);
//   - this node's record of the last received direction, the third drawn arrow (owned here);
//   - this node's VectorIn, drained non-blocking — the receive end is owned here, and a
//     direction already sitting in it would arrive on the next cycle and step the tilt
//     straight back off zero. Depth-1 latest-wins, so one receive empties it;
//   - this node's already-DELIVERED beads, drained off In the same way and for the same
//     reason — the bead edge paces the exchange that turns
//     these tilts (the bead paces each round trip of the vector exchange), so a reset that skips it
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
// Node1 ADDS the half turn (its mirror package does the opposite), the same
// opposite-senses convention outgoingVector already uses. Both signs land in the SAME drawn
// direction — ±180° in θ is the same place — so this is index bookkeeping, not geometry:
// each kind's indices keep walking in its own direction instead of one kind's jumping the
// other way at the turn.
func (n *Node) bottomTilt() Wiring.TiltVectorMsg {
	return Wiring.TiltVectorMsg{
		ThetaIdx: n.TopTiltThetaIdx + Wiring.HalfTurnThetaIdx,
		PhiIdx:   n.TopTiltPhiIdx,
	}
}

// coplanarNormal is THIS node's own coplanar normal: a quarter turn (90°, 6 steps of
// Wiring.CurveParamTiltVectorAngleStep — Wiring.PerpendicularThetaIdx names the same
// 90°-worth-of-steps magnitude) from THIS node's OWN tilt vector, so the normal stays
// perpendicular to the tilt as the tilt turns — index arithmetic only, never trig
// (memory/feedback_abc_times_constant_not_rederive.md). φ is left unchanged: the turn
// is entirely in θ, the same in-ring-plane assumption Wiring.PerpendicularThetaIdx's own
// doc comment spells out.
//
// theta is measured from world +y, so index 0 is the +y pole and each
// Wiring.HalfTurnThetaIdx (180°) crossed by TopTiltThetaIdx is a pole crossing. Every time
// the tilt has crossed an ODD number of poles, the coplanar normal itself has to gain a half
// turn to keep pointing the same drawn way, so this ADDS Wiring.HalfTurnThetaIdx on top of
// the usual quarter turn whenever floorDiv(TopTiltThetaIdx, HalfTurnThetaIdx) is odd. This is
// written as a PURE function of TopTiltThetaIdx alone — no stored "did we just cross" flag,
// no comparison against a previous value, no crossing event — so there is no missed-edge or
// double-flip bug class to have (memory/feedback_make_bug_class_unrepresentable.md). floorDiv
// is FLOOR division, not Go's truncating `/`: Node1's base direction subtracts one step, so
// TopTiltThetaIdx is negative for most of this node's life, and truncating division would
// flip parity on the wrong side of zero.
func (n *Node) coplanarNormal() Wiring.TiltVectorMsg {
	poles := floorDiv(n.TopTiltThetaIdx, Wiring.HalfTurnThetaIdx)
	thetaIdx := n.TopTiltThetaIdx + Wiring.PerpendicularThetaIdx
	if poles%2 != 0 {
		thetaIdx += Wiring.HalfTurnThetaIdx
	}
	return Wiring.TiltVectorMsg{
		ThetaIdx: thetaIdx,
		PhiIdx:   n.TopTiltPhiIdx,
	}
}

// floorDiv is FLOOR integer division (rounds toward negative infinity), unlike Go's `/`
// which truncates toward zero. coplanarNormal above is the one call site that needs this:
// pole-crossing parity must be correct for negative TopTiltThetaIdx, which is the common
// case since Node1's base direction subtracts.
func floorDiv(a, b int32) int32 {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
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
// 180° in θ. Node1 turns −180° (−12 steps of π/12); Node2 (its mirror package) turns
// +180° (+12 steps) — index arithmetic, never radians. φ is untouched.
func (n *Node) outgoingVector() Wiring.TiltVectorMsg {
	norm := n.coplanarNormal()
	norm.ThetaIdx -= 2 * Wiring.PerpendicularThetaIdx
	return norm
}

// stepFromVector ALWAYS steps this node's own TopTiltThetaIdx by exactly one step on every
// arrived vector — the exchange no longer has a halt condition. TWO DOT PRODUCTS decide only
// the SIGN of that one step:
//
//   - arrived vector ACUTE with this node's own TOP tilt vector    -> step -1 (Node1's base
//     direction)
//   - arrived vector ACUTE with this node's own BOTTOM tilt vector -> step +1, the REVERSE
//   - neither acute (exactly perpendicular)                        -> step -1, falling to
//     Node1's base direction — this used to be the case that halted the exchange (no step,
//     caller sends nothing); now it is just another sign choice, and the exchange keeps
//     sweeping through the perpendicular index instead of stopping there.
//
// The two acute cases are mutually exclusive: the bottom is the top's exact antipode, so the
// two dots are always exact negatives and at most one of them can be positive. There is no
// both-acute case for the ordering of these two ifs to arbitrate, and no free sign knob —
// which end the arrived vector leans toward IS the direction, except at the perpendicular
// index itself, which has no lean and so falls to the base direction.
//
// Node1's base direction (the top-acute and perpendicular cases) is subtracts one step; its
// mirror package's is the opposite, so a pair still turns symmetrically when both are leaning
// the same way.
//
// The return value is now always true — every caller that still checks it always proceeds.
// It stays a bool return (rather than being dropped) so callers read as "this always steps"
// rather than silently assuming it.
func (n *Node) stepFromVector(received Wiring.TiltVectorMsg) bool {
	switch {
	case Wiring.TiltVectorIsAcute(received, n.bottomTilt()):
		n.TopTiltThetaIdx += 1
	case Wiring.TiltVectorIsAcute(received, n.topTilt()):
		fallthrough
	default:
		// Acute with the top, or exactly perpendicular to both — both fall to Node1's base
		// direction.
		n.TopTiltThetaIdx -= 1
	}
	return true
}

// topTilt is THIS node's own top tilt vector as a direction pair — the same indices held on
// the struct, named so the dot products above read as vector-against-vector rather than as
// two loose ints.
func (n *Node) topTilt() Wiring.TiltVectorMsg {
	return Wiring.TiltVectorMsg{ThetaIdx: n.TopTiltThetaIdx, PhiIdx: n.TopTiltPhiIdx}
}

// handleVectorCycle is Node1's WHOLE per-cycle vector-channel loop body: read
// VectorIn non-blocking; if something arrived, step (stepFromVector always steps now — the
// two dots only pick which way); and send outgoingVector back out on VectorOut, also
// non-blocking. There is no longer a perpendicular halt: every arrival steps and replies, so
// the exchange keeps sweeping through the perpendicular index instead of stopping there. The
// only thing that still stops the exchange is a RESET marker (below), not an index value.
// This never touches In/Out or beads — the vector channel is a separate, additive exchange.
func (n *Node) handleVectorCycle(tick int64) {
	received, ok := Wiring.PollRecvVector(n.VectorIn)
	if !ok {
		return
	}
	// A RESET marker is not a direction to act on: run this node's FULL clear (indices,
	// third arrow, vector end, delivered beads, and the beads still crossing this node's
	// own outgoing wires) and REPLY WITH NOTHING. Replying would bounce the reset back and
	// forth forever; the marker's job is to stop the exchange, so it ends here. This is
	// the clear that actually makes the pair quiescent — see clear's own doc comment on
	// why the marker-driven one, not the button-driven one, is the one that lands last.
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
	if !n.stepFromVector(received) {
		return
	}
	n.syncTiltIndex()
	Wiring.SendVectorLatestNonBlocking(n.VectorOut, n.outgoingVector())
	// The bead rides along with the vector: one message, one visible bead, so the bead
	// loop ends exactly when the exchange does. THIS goroutine is still the sole placer on
	// this Out (wire.Out.PlaceDrivenAt's one-goroutine-per-Out invariant) — the placement
	// only moved between two branches of this same loop.
	if n.Out != nil {
		n.Out.PlaceDrivenAt(1, tick)
	}
}

func (n *Node) Update(ctx context.Context) {
	wire.TryEmit(n.EmitGeometry)

	// Report THIS node's OPENING tilt/normal pair once, before the loop. The mover is a
	// passive mirror of these (moveMsgKindTiltIndexSync) and has no way to derive the
	// normal itself, so without this its normal indices sit at their zero value until the
	// first arrival or panel click — and since the tilt index opens at 0 too, both
	// directions decode to world +y and the two drawn arrows superimpose, which reads as
	// the coplanar normal being missing entirely.
	n.syncTiltIndex()

	// Copy taken ONCE at this goroutine's start (Update IS the goroutine).
	clk := n.clock().Copy()

	for {
		if ctx.Err() != nil {
			return
		}

		// Drain In non-blocking. A bead arrival PACES the exchange and marks the round
		// trip; it DECIDES nothing. It used to step this node's tilt one click in this
		// kind's own fixed direction, with no reference to anything that arrived — so
		// every bead round trip turned this node the same way forever, independently of
		// (and on top of) the dot rule that is supposed to own that decision. Two rules
		// moved one index: when they agreed the node double-stepped, when they disagreed
		// they cancelled and it froze. The dots are now the only thing that turns a tilt
		// on an arrival, and the bead is what makes that turn visible and timed.
		//
		// It does not place a bead onward either: the bead now travels WITH the vector,
		// placed by handleVectorCycle when the dots actually move this node, so the bead
		// loop lives and dies with the exchange it is pacing instead of circulating on
		// its own.
		if _, ok := n.In.PollRecv(); ok {
			if n.Fire != nil {
				n.Fire()
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

		// Vector-channel exchange: the ONE place an arrival turns this node's tilt, and
		// the place the outgoing bead is now placed from — see handleVectorCycle's own
		// doc comment.
		n.handleVectorCycle(clk.Tick())

		wire.ApplySpeedNonBlocking(clk, n.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

func init() {
	// Node1 CONSTRUCTS ITSELF (Wiring.RegisterBuilder), same self-construction
	// shape as every other kind — see Pacer/Input for the general note on why
	// this replaced reflectBuild.
	Wiring.RegisterBuilder("Node1",
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
