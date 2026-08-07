// Package Node1 is the "Node1" kind: a pair is two nodes of this one kind. It is
// REACTIVE, not periodic: every cycle it drains its own In and its own
// VectorIn non-blockingly, and runs the straightening rule ITSELF on what arrived. An In
// bead PACES the exchange and decides nothing; the rule lives on the VECTOR channel
// (handleVectorCycle below) — two acute tests against this node's own top and bottom tilt
// vectors decide whether this node turns and which way, and only if it turned does it reply
// with a vector and place a bead. There is ONE target — the two tilts a quarter turn apart
// ("square") — and reaching it is what halts the exchange, not either test alone. This
// all runs on THIS goroutine: there is no round trip to a second goroutine to decide (see
// TopTiltThetaIdx below for who else the index is reported to and why).
//
// Emission is otherwise silent: with no In arrival there is nothing to react to, and the
// loop is kicked off by a USER — routed here via its own dedicated TiltEditIn channel
// (BuildArgs.TiltEditIn), also drained non-blockingly every cycle. TiltEditIn carries THREE
// distinct edits (task/pair-node-owns-itself split), never conflated:
//
//   - TiltVectorAnglePanel's ▲/▼ click: applies exactly one ±1 step to the named axis, marks
//     this end HELD (a tilt a user set is intent, not error — the partner moves to restore
//     square instead of this end turning on an arrival), and ALSO opens the vector exchange
//     by sending this node's own outgoing vector alongside a bead — a click that only moved an
//     index would leave the partner with nothing to answer.
//   - the START TILT button (TiltVectorButtons.tsx, TiltEditMsg.Start): opens the vector
//     exchange from whatever angles are CURRENTLY set — sends this node's own outgoing
//     vector alongside a bead ("THE KICK"), which is what gives handleVectorCycle something
//     to reply to; a channel whose only sends are replies never carries anything at all. It
//     changes NO index of its own. With both nodes of a pair perpendicular nothing
//     circulates on In, correctly, since there is nothing left to straighten, so the loop
//     has no way to start on its own — Start is the thing a user clicks to start it.
//     Pairing two Node1 instances with one edge each direction (a.Out → b.In,
//     b.Out → a.In) needs no seed/bootstrap node: nothing ever sends until a user
//     starts it, so there is no deadlock to bootstrap out of at t=0.
//   - the RESET button (TiltVectorButtons.tsx, TiltEditMsg.Reset): the opposite of Start — it
//     places NO bead, a stop-and-return, not a nudge, so it never starts the straightening
//     exchange. It does more than zero the indices, because zeroed indices are not by
//     themselves a stopped exchange: it runs this node's full clear() (below), which also
//     empties the bead edge — the thing that has actually been turning these tilts — so
//     nothing is left in the pair that could land a moment later and step it back off zero.
package Node1

import (
	"context"
	"fmt"
	"strconv"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

type Node struct {
	// PairID is THIS NODE'S OWN SPEC ID, the number the editor draws on it — the builder
	// parses it from BuildArgs.Name(), which is the node's directory name under topology/
	// and is a number by construction (.claude/rules/persistence-ownership.md: node ids ARE
	// numbers, strings only because they are directory names).
	//
	// It decides ONE thing: START opens the exchange from id 1 alone (applyTiltEdit). The
	// panel posts START to every node row, because the webview holds no domain knowledge
	// about which node should open; Go decides here, by id.
	//
	// It decides NOTHING ELSE. There is one pair kind and one implementation, so both ends
	// of a pair derive the same directions and step the same way, and the id is the only
	// thing that distinguishes them at all. A pair therefore needs the node numbered 1 to be
	// one of its ends — with no id 1 present, nothing opens the exchange and START does
	// nothing.
	//
	// The zero value is what a bare test build in this package constructs; it is not id 1,
	// so such a node does not open an exchange unless the test says PairID: 1.
	PairID int32

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
	// preserving wire.Out.PlaceDrivenAt's one-goroutine-per-Out invariant — nothing else
	// places on this Out at all.
	Out *wire.Out
	// Top is THIS node's OWN tilt direction, held as a STATE ON THE RING (ring.go) rather
	// than as a number this goroutine does arithmetic to. Turning is following a link, so
	// there is no index to keep in range and no twenty-fifth direction to land on.
	//
	// The ONE writer, full stop: seeded once at build time from the persisted value
	// (BuildArgs.TiltVectorAngleSeed, mapped through stateFor) and moved ONLY by this
	// goroutine's own Update loop, below. Every change is reported one-way to this node's own
	// geometry (SyncTiltIndex, i.e. Self) so the geometry — which owns streaming this node's
	// scene columns and persisting them to its own position.json — stays in sync; the
	// geometry never decides or mutates this itself, it mirrors what it is told.
	//
	// nil means the ring's origin: a bare test build constructs this struct without a
	// builder, and topState below reads a nil Top as direction 0 rather than making every
	// such test say so. There is no companion φ — every tilt vector in this exchange lives in
	// the θ-only plane (memory/feedback_abc_times_constant_not_rederive.md: an index times a
	// step constant, trig only at the cartesian/polar boundary).
	Top *tiltState

	// Halt is WHICH RESTING STATE this node is holding — perpendicular or parallel (ring.go's
	// haltKind). The two are different states with different halts, and a node has to know
	// which of them it is in, because that is what says which way to turn when something
	// disturbs it: an arrival that leans takes the pair off whichever relationship it was in,
	// and the correction for one is the correction for the other run backwards.
	//
	// It is written when an arrival lands on either halt and by nothing else — no arrival in
	// between erases it, so a node disturbed mid-turn still knows what it is returning to. The
	// RESET button is the one thing that erases it (clear).
	Halt haltKind

	// Ring is THIS NODE'S OWN lattice — its states, and the counts every rule reads off them.
	// The point count is a scene setting a user can change, so this is not fixed for the life
	// of the process; a change means this goroutine building itself a new ring, never a
	// shared one being rewritten under other readers.
	//
	// nil means the default lattice, which is what a bare test build gets — see ringOf below,
	// the one read of this field.
	Ring *ring
	// TiltEditIn is this node's dedicated channel for a panel-driven tilt-angle click
	// (TiltVectorAnglePanel), claimed at build time via BuildArgs.TiltEditIn — see the
	// package doc comment's "THE KICK".
	TiltEditIn <-chan Wiring.TiltEditMsg
	// LatticeIn carries a new POINT COUNT for this node's own ring — the scene setting the
	// angles panel changes, delivered to every pair node on its own dedicated channel, the
	// same shape as TiltEditIn. Drained non-blocking every cycle; a value that matches the
	// count this node already runs is a no-op (adoptLattice). nil on a bare test build.
	LatticeIn <-chan int32
	// SyncLatticePoints notifies this node's own geometry of the current lattice point
	// count — one-way, fire-and-forget, never an ack, same shape as SyncTiltIndex below.
	// The geometry converts a tilt-vector INDEX to an angle every frame (2π / points per
	// step) but does not itself decide the count, so it has to be told whenever this
	// goroutine adopts a new one (BuildArgs.SyncLatticePoints, i.e. Self.SetLatticePoints).
	SyncLatticePoints func(points int32)
	// SyncTiltIndex notifies this node's own geometry of the current TopTiltThetaIdx AND the
	// current coplanar-normal index (coplanarNormal, below) — one-way, fire-and-forget,
	// never an ack (BuildArgs.SyncTiltIndex).
	SyncTiltIndex func(theta, normalTheta, bottomTheta int32)
	// VectorOut/VectorIn are THIS node's own ends of its dedicated tilt-vector channel
	// (Wiring.TiltVectorMsg — an integer θ index, never floats on a channel),
	// claimed at build time via BuildArgs.VectorOut/VectorIn. It travels ALONGSIDE the
	// ordinary bead edge (In/Out above), never replacing it — beads are unaffected.
	// Buffered depth 1, latest-wins, non-blocking on both ends
	// (Wiring.SendVectorLatestNonBlocking / Wiring.PollRecvVector). nil when this
	// node's edge partner did not also ask for one, or on a bare test build with no
	// loader — both helpers already treat nil as "nothing wired".
	VectorOut chan<- Wiring.TiltVectorMsg
	VectorIn  <-chan Wiring.TiltVectorMsg
	// ReceivedThetaIdx/ReceivedSet are THIS node's own record of the LAST
	// direction that ARRIVED on VectorIn — the third drawn arrow (user request: "show a
	// 3rd vector...the last iteration of it as a different color in the node that
	// received it"). Written ONLY by this goroutine, in handleVectorCycle below: an
	// arrival REPLACES whatever was here before (never accumulates), and it persists
	// indefinitely otherwise — it is NOT cleared when the straightening exchange settles.
	// It IS cleared by a RESET, both this node's own (applyTiltEdit's Reset branch) and a
	// Reset marker arriving on VectorIn (handleVectorCycle's Reset branch): a reset is a
	// stop-and-return, and a stale received arrow left hanging would contradict that.
	// Reported one-way to this node's own geometry via SyncReceivedVector, same shape as
	// TopTiltThetaIdx/SyncTiltIndex above.
	ReceivedThetaIdx int32
	ReceivedSet      bool
	// SyncReceivedVector notifies this node's own geometry of the current
	// ReceivedThetaIdx/Set — one-way, fire-and-forget, never an ack
	// (BuildArgs.SyncReceivedVector).
	SyncReceivedVector func(theta int32, set bool)
	// ClearOutBeads drops every bead still crossing this node's outgoing wires — a call on
	// this node's own Self (BuildArgs.ClearOutBeads), not a message to a second goroutine.
	// Called only from clear(), below: this goroutine drives those wires, so it clears them
	// through the same object that drives them rather than reaching into the wire.
	ClearOutBeads func()
	// Self is this node's own geometry/mover state (task/pair-node-owns-itself,
	// Wiring.PairNodeSelf), claimed at build time via BuildArgs.ClaimSelfDrive. THIS
	// goroutine (Update, below) is the sole driver of it — there is no separate
	// nodeMover goroutine for this node any more. nil on a bare test build with no
	// loader; every PairNodeSelf method is nil-safe.
	Self *Wiring.PairNodeSelf
}

func (n *Node) clock() wire.Clock {
	if n.Clock == nil {
		return wire.NewRealClock()
	}
	return n.Clock
}

// applyTiltEdit applies one panel-driven edit — TiltVectorAnglePanel's ±1 click, the START
// TILT button (TiltVectorButtons.tsx), or the RESET button (same file) — directly to this
// node's OWN indices, same no-mover-round-trip shape as stepFromVector. Reports whether the
// caller should place a bead: true for Start and for a plain adjust, false only for a reset.
//
// The three branches are now split (task/pair-node-owns-itself):
//
//   - Reset: a stop-and-return, not a nudge — see clear's own doc comment.
//   - Start: opens the vector exchange from whatever angles are CURRENTLY set, by sending
//     this node's own outgoingVector on VectorOut and placing a bead. It changes NO index.
//     This is the vector channel's whole starting move, and without it the channel never
//     carries a direction at all: handleVectorCycle only ever sends in REPLY to an arrival,
//     so with nothing to reply to, no node ever received one, no node ever set its
//     received-direction record, and the third arrow could not be drawn anywhere.
//   - a plain adjust (neither Reset nor Start): applies the ±1 click to the named axis, marks
//     this end Held (a tilt a user set is intent, so this end keeps its index and does not
//     turn on an arrival — the partner moves instead), and ALSO OPENS THE EXCHANGE by sending
//     this node's own outgoing vector and asking the caller to place a bead. Previously an
//     adjust sent nothing, which is why a tilted pair just sat there: the tilted end went deaf
//     and silent at once.
func (n *Node) applyTiltEdit(edit Wiring.TiltEditMsg) (placeBead bool) {
	if edit.Reset {
		n.clear()
		// Tell the partner, so it clears too — see clear's own doc comment for why the
		// partner's clear, not this one, is what actually ends the exchange.
		Wiring.SendVectorLatestNonBlocking(n.VectorOut, Wiring.TiltVectorMsg{Reset: true})
		return false
	}
	if edit.Start {
		// START BELONGS TO PAIR ID 1 ALONE. The exchange is begun from ONE end, so there is
		// exactly one opening direction to answer; opened from both, each end also replies to
		// the other's opener in the same round — two exchanges running through one pair of
		// channels rather than the one a user asked for, which shows up as the pair settling
		// and then being kicked off its rest state again, forever.
		//
		// The panel sends START to every node it lists (TiltVectorButtons.tsx posts one
		// record per row, exactly as RESET does), because the WEBVIEW must not know which end
		// is which — that is domain knowledge, and TS holds none. Go decides, here, by id.
		if n.PairID != 1 {
			return false
		}
		// Open the vector exchange from the current angles — see this function's own doc
		// comment. Sends exactly what the old adjust-side-effect kick sent, but changes no
		// index of its own.
		Wiring.SendVectorLatestNonBlocking(n.VectorOut, n.outgoingVector())
		return true
	}
	if edit.Up {
		n.Top = n.topState().next
	} else {
		n.Top = n.topState().prev
	}
	// AND IT STOPS THERE: no send, no bead. Setting an angle and running the exchange are
	// separate acts — a click moves this node's own tilt and nothing else happens until START.
	return false
}

// adoptLattice rebuilds THIS node's own ring at a new point count, on THIS node's own
// goroutine. Nothing else touches the ring, so there is nothing to coordinate: the old one is
// simply dropped and a new one takes its place.
//
// WHAT SURVIVES THE CHANGE IS THE INDEX, not the angle. A tilt at 6 stays at 6 — which is a
// quarter turn on a 24-point lattice and a half turn on a 12-point one, so the drawn arrow
// moves. That is the honest reading of "the lattice changed underneath a direction": the
// number a user set is kept, and what it means follows the new lattice. An index the new ring
// does not have names nothing there, so that node opens at the origin and says so
// (ring.seedState).
//
// TWO THINGS ARE DISCARDED, both because they are indices on the lattice being left:
//
//   - the received direction, the third drawn arrow. It was picked on the old lattice, so
//     redrawing it at the same index would point it somewhere the partner never sent.
//   - whatever is queued on VectorIn. Same reason, and it would otherwise be read as a
//     direction on the new ring the moment the next cycle polls.
//
// The beads in flight are untouched: a bead carries no direction, only pacing.
func (n *Node) adoptLattice(points int32) {
	if points == n.ringOf().points {
		return
	}
	keptIdx := n.topState().idx
	n.Ring = newRing(points)
	top, unknown := n.Ring.seedState(keptIdx)
	n.Top = top
	if unknown && n.Self != nil {
		n.Self.Breadcrumb("pair-lattice-adopt", fmt.Sprintf(
			"points=%d keptIdx=%d unknown=true loaded=%d", points, keptIdx, top.idx))
	}
	n.ReceivedThetaIdx = 0
	n.ReceivedSet = false
	n.syncReceivedVector()
	Wiring.PollRecvVector(n.VectorIn)
	if n.SyncLatticePoints != nil {
		n.SyncLatticePoints(points)
	}
	n.syncTiltIndex()
}

// ringOf is this node's own lattice, with the default standing in for a Ring that was never
// set — a bare test build. Every read of the lattice goes through here.
func (n *Node) ringOf() *ring {
	if n.Ring == nil {
		return defaultRing
	}
	return n.Ring
}

// topState is this node's own tilt direction, with its ring's origin standing in for a Top
// that was never set — see the field's own doc comment. Every read of the tilt goes through
// here, so nothing else in this file has to care about that case.
func (n *Node) topState() *tiltState {
	if n.Top == nil {
		return n.ringOf().at(0)
	}
	return n.Top
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
//   - this node's OUTGOING beads, still crossing. A PacedWire is driven by its source
//     node's own mover state, which for this kind is this goroutine's own Self, so it drops
//     them through that (ClearOutBeads / PairNodeSelf.ClearOutBeads) rather than reaching
//     into the wire.
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
	n.Top = n.ringOf().at(0)
	// The held resting state goes too: RESET is the one thing that erases it.
	n.Halt = haltNone
	n.syncTiltIndex()
	n.ReceivedThetaIdx = 0
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

// bottomTilt is THIS node's own BOTTOM TILT VECTOR: the state a half turn from its own top,
// so it points out of the node's other side and turns with the top as the top turns. A LINK,
// resolved when the ring was built (ring.go) — no arithmetic, and no trig
// (memory/feedback_abc_times_constant_not_rederive.md). There is no φ any more: a half turn
// in θ alone already negates the direction exactly (see Wiring.HalfTurnThetaIdx's own doc
// comment).
func (n *Node) bottomTilt() Wiring.TiltVectorMsg {
	return Wiring.TiltVectorMsg{ThetaIdx: n.topState().opposite.idx}
}

// coplanarNormal is THIS node's own coplanar normal: ONE QUARTER TURN past this node's own
// tilt vector, always the same way round — Wiring.PerpendicularThetaIdx (6 steps of
// Wiring.CurveParamTiltVectorAngleStep, 90°) ADDED to the tilt index, and nothing else. Both
// ends of a pair run this same unmodified addition — there is no sign difference between
// them. Index arithmetic, never trig (memory/feedback_abc_times_constant_not_rederive.md).
//
// The wrap is a COMPARISON, not a division: a quarter turn added to an index can carry it at
// most one full turn past the top of the circle, so one test against Wiring.FullTurnThetaIdx
// and one subtraction bring it back. No `/`, no `%`, and therefore no sign convention to get
// wrong at negative indices — which is what the arithmetic this replaced needed floor
// division for.
//
// DO NOT ADD A POLE-CROSSING TERM HERE. One was here, adding a further half turn on odd
// multiples of HalfTurnThetaIdx, on the reasoning that θ measured from world +y makes each
// half turn a pole crossing, and that crossing a pole flips φ by 180° so the normal needs its
// own half turn to keep pointing the same drawn way. That reasoning needs a φ to flip, and
// this model has none: task/drop-tilt-vector-phi removed the φ column, and every consumer
// treats an index as a position on a plain circle — the renderer decodes it as
// (sinθ, cosθ, 0), which has no pole and no fold anywhere. What the term actually did was
// point the normal the OTHER way, t + 18 rather than t + 6, over half the index range,
// including the negative indices a ▼ click reaches first.
func (n *Node) coplanarNormal() Wiring.TiltVectorMsg {
	return Wiring.TiltVectorMsg{ThetaIdx: n.topState().quarter.idx}
}

// syncTiltIndex reports THIS node's current tilt index AND its current coplanar-normal
// index (coplanarNormal above) to this node's own geometry in one call — every call site
// that changes TopTiltThetaIdx must also report the normal, since the normal is
// derived from the tilt and the geometry does not derive it itself (see
// Wiring.PairNodeSelf.SetTiltIndex, and moveMsgKindTiltIndexSync's retirement note in
// move_msg.go for what this used to be). nil-safe, same as every other closure call here.
func (n *Node) syncTiltIndex() {
	if n.SyncTiltIndex == nil {
		return
	}
	norm := n.coplanarNormal()
	bottom := n.bottomTilt()
	n.SyncTiltIndex(n.topState().idx, norm.ThetaIdx, bottom.ThetaIdx)
}

// syncReceivedVector reports THIS node's current received-vector state (ReceivedThetaIdx/
// ReceivedThetaIdx/Set) to this node's own geometry — the third-arrow twin of syncTiltIndex.
// Called by every site that changes those fields, below. nil-safe, same as syncTiltIndex.
func (n *Node) syncReceivedVector() {
	if n.SyncReceivedVector == nil {
		return
	}
	n.SyncReceivedVector(n.ReceivedThetaIdx, n.ReceivedSet)
}

// outgoingVector is what THIS node SENDS on VectorOut: its own coplanarNormal. The message
// on the channel IS the direction this node computed and draws, so the arrow the partner
// draws as its received direction coincides with this node's own normal on screen.
//
// Do not rotate it on the way out. A rotation here has to be undone by the receiver's step
// signs (stepFromVector, below) to leave behaviour unchanged, which would spread one
// convention across two call sites of the same rule instead of leaving it in the one place;
// and rotating by a half turn in particular changes nothing at all about where the pair
// comes to rest, since the bottom tilt is the top plus that same half turn — it only swaps
// which of the receiver's two acute tests fires.
func (n *Node) outgoingVector() Wiring.TiltVectorMsg {
	v := n.coplanarNormal()
	// The lattice the index is on travels with it — see TiltVectorMsg.Points. Without it the
	// partner cannot tell a direction picked before a point-count change from one picked
	// after, and the two mean different angles.
	v.Points = n.ringOf().points
	return v
}

// stepFromVector decides whether an arrived direction turns this node at all, and if so which
// way. THERE IS ONE TARGET: an arrival landing exactly on this node's own TOP means the two
// tilts of the pair are a quarter turn apart ("square") — that is the only resting state.
// Colinear is no longer one.
//
//   - arrival == this node's own TOP                 -> SQUARE: clear Held, return false (halt).
//   - Held && arrival == this node's own BOTTOM       -> the partner has finished (its normal
//     lands a half turn away when IT was the one that moved): clear Held, return false.
//   - Held                                            -> return true without moving. A held end
//     still replies, so the partner has something to correct against.
//   - otherwise                                       -> step TOWARD the arrival by the shorter
//     way round (stepsToward/ringGap), and return true.
//
// This node no longer picks a direction from an acute-cone test against its top/bottom
// (tiltState.acuteWith): a cone says the arrival is within a quarter turn but not which SIDE,
// so half the time the step went away from the arrival and the pair oscillated one step either
// side of a boundary. Distance strictly closes the gap every step instead, so the walk ends.
//
// Both ends of a pair run this same unmodified rule, and both directions of travel are links
// rather than ±1, so a step cannot leave the ring. The pairing that matters is with what
// outgoingVector sends: this reads an arrival that is the partner's coplanar normal as-is.
// Worked run: docs/pair-node/vectors.html.
func (n *Node) stepFromVector(received Wiring.TiltVectorMsg) bool {
	arrival := n.ringOf().arrivedState(received.ThetaIdx)
	before := n.topState()

	// EACH RESTING STATE HAS ITS OWN HALT. The arrival names one of them or neither
	// (haltAgainst). Landing on one records WHICH — nothing else writes this, so an arrival
	// that lands between the two leaves the node still knowing what it was holding — and the
	// node stands still, because it is already where it belongs.
	if halt := before.haltAgainst(arrival); halt != haltNone {
		n.Halt = halt
		return true
	}

	// Disturbed. The acute tests say which side the arrival leans, and the halt this node is
	// holding says what to do about that: closing on the arrival walks the pair toward
	// parallel, so a node holding parallel takes that step and a node holding perpendicular
	// takes it backwards. A node holding neither yet has nothing to return to and closes.
	switch {
	case before.acuteWith(arrival):
		if n.Halt == haltPerpendicular {
			n.Top = before.prev
		} else {
			n.Top = before.next
		}
	case before.opposite.acuteWith(arrival):
		if n.Halt == haltPerpendicular {
			n.Top = before.next
		} else {
			n.Top = before.prev
		}
	}
	return true
}

// handleVectorCycle is Node1's WHOLE per-cycle vector-channel loop body: read
// VectorIn non-blocking; if something arrived, step (stepFromVector decides whether this node
// turns at all, and which way — see its own doc comment for the one target, square); and if it
// stepped, send outgoingVector back out on VectorOut, also non-blocking, and place the paired
// bead. On a false return from stepFromVector (the arrival landed exactly on this node's own
// top) this sends nothing and places no bead — that is the ONE way the vector exchange comes to
// rest. A RESET marker (below) is the other way the exchange stops. This never touches In/Out
// or beads on the halt path — the vector channel is a separate, additive exchange.
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
	// A DIRECTION FROM ANOTHER LATTICE IS NOT A DIRECTION HERE. The two ends of a pair adopt
	// a new point count at their own moments, each on its own goroutine, so between those
	// moments an index picked on the old lattice can land here — where it names a different
	// angle, or no state at all. Dropping it is the definite answer: the partner adopts the
	// same count within its own next cycle and the exchange resumes from directions both
	// ends can read. Zero is a bare test build that stated nothing, and is taken as this
	// node's own lattice.
	if received.Points != 0 && received.Points != n.ringOf().points {
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
	n.ReceivedSet = true
	n.syncReceivedVector()
	// DIAGNOSTIC (task/log-pair-vector-exchange): everything the two acute tests read, and
	// what this node decided from them, in one row per arrival. Recorded BEFORE the step so
	// the `top`/`bottom` here are the operands the tests actually used, not the post-step ones.
	// A halted pair re-reads the same arrival every tick, so a row per arrival was 1749 rows of
	// `hold` in 1768 — the per-tick firehose .claude/rules/go-debugging.md warns about. Only a
	// row where something CHANGED is written: the index moved, or the held resting state did.
	before := n.topState()
	arrival := n.ringOf().arrivedState(received.ThetaIdx)
	acuteTop := before.acuteWith(arrival)
	acuteBottom := before.opposite.acuteWith(arrival)
	heldBefore := n.Halt
	moved := n.stepFromVector(received)
	if n.Self != nil && (n.topState() != before || n.Halt != heldBefore) {
		kind, dir := "none", "hold"
		switch {
		case acuteTop:
			kind = "top"
		case acuteBottom:
			kind = "bottom"
		}
		switch {
		case n.topState() == before.next:
			dir = "next"
		case n.topState() == before.prev:
			dir = "prev"
		}
		n.Self.Breadcrumb("pair-vector", fmt.Sprintf(
			"id=%d recv=%2d top=%2d bottom=%2d sep=%2d kind=%-6s held=%-13s -> %-4s idx %2d->%2d sent=%2d moved=%v tick=%d",
			n.PairID, received.ThetaIdx, before.idx, before.opposite.idx,
			before.separation(arrival), kind, n.Halt, dir, before.idx, n.topState().idx,
			n.outgoingVector().ThetaIdx, moved, tick))
	}
	if !moved {
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
	// This node's own mover-owned startup geometry emit — see Self's own doc comment.
	// There is no separate nodeMover goroutine to make this emit any more.
	n.Self.EmitGeometryOnce()

	// Report THIS node's OPENING tilt/normal pair once, before the loop. Self is a
	// passive mirror of these (PairNodeSelf.SetTiltIndex) and has no way to derive the
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
		// (and on top of) the acute-test rule that is supposed to own that decision. Two
		// rules moved one index: when they agreed the node double-stepped, when they
		// disagreed they cancelled and it froze. The tests are now the only thing that turns a tilt
		// on an arrival, and the bead is what makes that turn visible and timed.
		//
		// It does not place a bead onward either: the bead now travels WITH the vector,
		// placed by handleVectorCycle when the tests actually move this node, so the bead
		// loop lives and dies with the exchange it is pacing instead of circulating on
		// its own.
		if _, ok := n.In.PollRecv(); ok {
			if n.Fire != nil {
				n.Fire()
			}
		}

		// Drain TiltEditIn non-blocking: a panel/RESET/START edit — see the package doc
		// comment for the three-way split. applyTiltEdit decides placeBead: true for Start
		// and for a plain adjust (both open the exchange), false only for Reset.
		if n.TiltEditIn != nil {
			select {
			case edit := <-n.TiltEditIn:
				placeBead := n.applyTiltEdit(edit)
				n.syncTiltIndex()
				if placeBead && n.Out != nil {
					n.Out.PlaceDrivenAt(1, clk.Tick())
				}
				// DIAGNOSTIC: the BOUNDARY of an exchange — which edit a user sent and the
				// indices it left behind. Reading the log, every "pair-vector" row between
				// two of these belongs to the run this one started.
				if n.Self != nil {
					n.Self.Breadcrumb("pair-tilt-edit", fmt.Sprintf(
						"tick=%d reset=%v start=%v up=%v placeBead=%v theta=%d",
						clk.Tick(), edit.Reset, edit.Start, edit.Up, placeBead,
						n.topState().idx))
				}
			default:
			}
		}

		// Drain LatticeIn non-blocking: a new point count for this node's own ring. Drained
		// BEFORE the vector cycle below so that anything already queued on VectorIn is
		// discarded by the adopt rather than read one last time against the lattice it was
		// not picked on.
		if n.LatticeIn != nil {
			select {
			case points := <-n.LatticeIn:
				n.adoptLattice(points)
			default:
			}
		}

		// Vector-channel exchange: the ONE place an arrival turns this node's tilt, and
		// the place the outgoing bead is now placed from — see handleVectorCycle's own
		// doc comment.
		n.handleVectorCycle(clk.Tick())

		// This node's own mover work — drain its own dedicated inbound channels
		// (drag/select/hover/center/neighborCenter/etc.), drive its own outgoing
		// wires one cycle, retry pending sends, write its own dedicated stream
		// frame. Run on THIS goroutine, on THIS node's own clock tick: there is no
		// separate nodeMover goroutine for this node any more (task/pair-node-owns-
		// itself) — see Self's own doc comment.
		n.Self.Step(ctx, clk.Tick())

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
			// This node's own spec id, which is what START is addressed by — see PairID's
			// own doc comment. A name that is not a number leaves PairID at 0, so such a
			// node simply never opens an exchange rather than silently becoming id 1.
			if id, err := strconv.Atoi(a.Name()); err == nil {
				n.PairID = int32(id)
			}
			n.Fire = a.Fire()
			if clk := a.Clock(); clk != nil {
				n.Clock = clk
			}
			n.SpeedCh = a.SpeedCh()
			n.In = a.In("In")
			n.Out = a.Out("Out")
			// The persisted seed is a NUMBER from outside this kind — an old position.json
			// can hold anything, including a running count from before the tilt became a
			// state — so it comes in through seedState, which asks the ring which state
			// carries that index. After this line the tilt is a state and stays one.
			// This node's own lattice, opened at the scene's currently-persisted point
			// count (view/lattice.json via BuildArgs.LatticePointsSeed) rather than the
			// compile-time default.
			latticeSeed := a.LatticePointsSeed()
			n.Ring = newRing(latticeSeed)
			seed, seedUnknown := n.Ring.seedState(a.TiltVectorAngleSeed())
			n.Top = seed
			n.TiltEditIn = a.TiltEditIn()
			n.LatticeIn = a.LatticeIn()
			// Self replaces the old SyncTiltIndex/SyncReceivedVector/ClearOutBeads
			// messages-to-a-separate-mover-goroutine (task/pair-node-owns-itself):
			// this node's own goroutine now owns that mover state directly, so what
			// used to be a message is a plain method call on the same object below.
			self := a.ClaimSelfDrive()
			n.Self = self
			n.SyncLatticePoints = func(points int32) {
				self.SetLatticePoints(points)
			}
			n.SyncLatticePoints(latticeSeed)
			if seedUnknown {
				// The persisted index is not one this ring has — a position.json written
				// before the tilt became a state, or by a build with a different lattice.
				// The node opens at the origin and says which number it refused, rather
				// than computing some other direction and drawing it as if chosen.
				self.Breadcrumb("pair-seed-unknown", fmt.Sprintf(
					"node=%s persisted=%d loaded=%d", a.Name(), a.TiltVectorAngleSeed(), seed.idx))
			}
			n.SyncTiltIndex = func(theta, normalTheta, bottomTheta int32) {
				self.SetTiltIndex(theta, normalTheta, bottomTheta)
			}
			n.SyncReceivedVector = func(theta int32, set bool) {
				self.SetReceivedVector(theta, set)
			}
			n.ClearOutBeads = func() { self.ClearOutBeads() }
			n.VectorOut = a.VectorOut()
			n.VectorIn = a.VectorIn()
			// EmitGeometry stays nil deliberately — n.Self.EmitGeometryOnce (Update)
			// makes this node's own startup geometry emit instead.
			return n, nil
		})
}
