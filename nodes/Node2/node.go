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
// loop is kicked off by a USER — routed here via its own dedicated TiltEditIn channel
// (BuildArgs.TiltEditIn), also drained non-blockingly every cycle. TiltEditIn carries THREE
// distinct edits (task/pair-node-owns-itself split), never conflated:
//
//   - TiltVectorAnglePanel's ▲/▼ click: applies exactly one ±1 step to the named axis and
//     stops — no send, no bead. It used to ALSO open the vector exchange as a side effect
//     ("the kick"), so one click moved the tilt by many π/12 steps once the exchange
//     settled instead of exactly one; that side effect is now Start's alone.
//   - the START TILT button (TiltVectorButtons.tsx, TiltEditMsg.Start): opens the vector
//     exchange from whatever angles are CURRENTLY set — sends this node's own outgoing
//     vector alongside a bead ("THE KICK"), which is what gives handleVectorCycle something
//     to reply to; a channel whose only sends are replies never carries anything at all. It
//     changes NO index of its own. With both nodes of a pair perpendicular nothing
//     circulates on In, correctly, since there is nothing left to straighten, so the loop
//     has no way to start on its own — Start is the thing a user clicks to start it.
//     Pairing a Node1 and a Node2 with one edge each direction (Node1.Out → Node2.In,
//     Node2.Out → Node1.In) needs no seed/bootstrap node: nothing ever sends until a user
//     starts it, so there is no deadlock to bootstrap out of at t=0.
//   - the RESET button (TiltResetButton.tsx, TiltEditMsg.Reset): the opposite of Start — it
//     places NO bead, a stop-and-return, not a nudge, so it never starts the straightening
//     exchange. It does more than zero the indices, because zeroed indices are not by
//     themselves a stopped exchange: it runs this node's full clear() (below), which also
//     empties the bead edge — the thing that has actually been turning these tilts — so
//     nothing is left in the pair that could land a moment later and step it back off zero.
//
// Kept as a distinct package/kind rather than parametrizing Node1, per the check-dep-rules
// guard: a node-kind package may import only the shared spine, never a sibling kind, so
// Node1 and Node2 cannot share a struct across packages without a third shared package
// neither currently needs.
package Node2

import (
	"context"
	"fmt"

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
	// TopTiltThetaIdx is THIS node's OWN vector-direction index — the ONE writer, full stop
	// (memory/feedback_abc_times_constant_not_rederive.md: index × step-constant, trig only
	// at the cartesian/polar boundary). There is no companion φ: every tilt vector in this
	// exchange lives in the θ-only plane. Seeded once at build time from the persisted value
	// (BuildArgs.TiltVectorAngleSeed) and mutated ONLY by this goroutine's own Update loop,
	// below. Every change is reported one-way to this node's own mover (SyncTiltIndex) so the
	// mover — which still owns streaming this node's geometry and persisting it to this
	// node's own position.json — stays in sync; the mover never decides or mutates these
	// itself for this kind.
	TopTiltThetaIdx int32
	// TiltEditIn is this node's dedicated channel for a panel-driven tilt-angle click
	// (TiltVectorAnglePanel), claimed at build time via BuildArgs.TiltEditIn — see the
	// package doc comment's "THE KICK".
	TiltEditIn <-chan Wiring.TiltEditMsg
	// SyncTiltIndex notifies this node's own mover of the current TopTiltThetaIdx AND the
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
	// Reported one-way to this node's own mover via SyncReceivedVector, same shape as
	// TopTiltThetaIdx/SyncTiltIndex above.
	ReceivedThetaIdx int32
	ReceivedSet      bool
	// SyncReceivedVector notifies this node's own mover of the current
	// ReceivedThetaIdx/Set — one-way, fire-and-forget, never an ack
	// (BuildArgs.SyncReceivedVector).
	SyncReceivedVector func(theta int32, set bool)
	// ClearOutBeads asks THIS node's own mover to drop every bead still crossing this
	// node's outgoing wires — one-way, fire-and-forget, never an ack
	// (BuildArgs.ClearOutBeads). Called only from clear(), below: those beads are owned
	// by the mover (it drives the wires), so this node asks rather than reaching in.
	ClearOutBeads func()
	// Self is this node's own mover state (task/pair-node-owns-itself,
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
// caller should place "THE KICK" bead: true for Start, false for a plain adjust or a reset.
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
//   - a plain adjust (neither Reset nor Start): applies the ±1 click to the named axis and
//     STOPS — no send, no bead. It used to also open the vector exchange as a side effect,
//     so one click moved the tilt by many π/12 steps once the exchange settled instead of
//     exactly one; Start is what a user now clicks to begin the exchange separately.
func (n *Node) applyTiltEdit(edit Wiring.TiltEditMsg) (placeBead bool) {
	if edit.Reset {
		n.clear()
		// Tell the partner, so it clears too — see clear's own doc comment for why the
		// partner's clear, not this one, is what actually ends the exchange.
		Wiring.SendVectorLatestNonBlocking(n.VectorOut, Wiring.TiltVectorMsg{Reset: true})
		return false
	}
	if edit.Start {
		// START IS NODE1'S ALONE — this kind ignores it. The exchange is begun from ONE end
		// so there is exactly one opening direction to answer; started from both, each node
		// would also be replying to the other's opener in the same round, which is two
		// exchanges running through one pair of channels rather than the one a user asked
		// for. Node1's applyTiltEdit is where the send lives.
		//
		// The button still addresses every node the angles panel lists (TiltVectorButtons
		// .tsx sends one record per row, exactly as RESET does), because the WEBVIEW must
		// not know which node is node 1 — that is domain knowledge, and TS holds none. Go
		// decides, here, by kind.
		return false
	}
	delta := int32(-1)
	if edit.Up {
		delta = 1
	}
	n.TopTiltThetaIdx += delta
	return false
}

// clear returns THIS node to its opening state and — the part that matters — leaves
// nothing behind that could restart the straightening exchange (shared shape with Node1's
// copy, kept duplicated per-package for the same reason stepFromVector is). A reset is not "set
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

// bottomTilt is THIS node's own BOTTOM TILT VECTOR: a half turn (180°,
// Wiring.HalfTurnThetaIdx steps) in θ from its OWN top tilt vector, so it points out of the
// node's other side and turns with the top as the top turns — index arithmetic only, never
// trig (memory/feedback_abc_times_constant_not_rederive.md). There is no φ any more: a half
// turn in θ alone already negates the direction exactly (see Wiring.HalfTurnThetaIdx's own
// doc comment).
//
// Node2 SUBTRACTS the half turn (its mirror package does the opposite), the same
// opposite-senses convention outgoingVector already uses. Both signs land in the SAME drawn
// direction — ±180° in θ is the same place — so this is index bookkeeping, not geometry:
// each kind's indices keep walking in its own direction instead of one kind's jumping the
// other way at the turn.
func (n *Node) bottomTilt() Wiring.TiltVectorMsg {
	return Wiring.TiltVectorMsg{
		ThetaIdx: n.TopTiltThetaIdx - Wiring.HalfTurnThetaIdx,
	}
}

// coplanarNormal is THIS node's own coplanar normal: a quarter turn (90°, 6 steps of
// Wiring.CurveParamTiltVectorAngleStep — Wiring.PerpendicularThetaIdx names the same
// 90°-worth-of-steps magnitude) from THIS node's OWN tilt vector, so the normal stays
// perpendicular to the tilt as the tilt turns — index arithmetic only, never trig
// (memory/feedback_abc_times_constant_not_rederive.md). There is no φ: the turn
// is entirely in θ, the same in-ring-plane assumption Wiring.PerpendicularThetaIdx's own
// doc comment spells out.
func (n *Node) coplanarNormal() Wiring.TiltVectorMsg {
	// Node2 SUBTRACTS the quarter turn (Node1's mirror package adds it) — the tilt
	// vector and the coplanar normal sit −90° apart in θ for Node2, +90° for Node1,
	// keeping the sign with the KIND, same as every other per-kind sign here.
	return Wiring.TiltVectorMsg{
		ThetaIdx: n.TopTiltThetaIdx - Wiring.PerpendicularThetaIdx,
	}
}

// syncTiltIndex reports THIS node's current tilt index AND its current coplanar-normal
// index (coplanarNormal above) to this node's own mover in one call — every call site
// that changes TopTiltThetaIdx must also report the normal, since the normal is
// derived from the tilt and the mover no longer derives it itself (see
// Wiring.moveMsgKindTiltIndexSync's doc comment). nil-safe, same as every other closure
// call here.
func (n *Node) syncTiltIndex() {
	if n.SyncTiltIndex == nil {
		return
	}
	norm := n.coplanarNormal()
	bottom := n.bottomTilt()
	n.SyncTiltIndex(n.TopTiltThetaIdx, norm.ThetaIdx, bottom.ThetaIdx)
}

// syncReceivedVector reports THIS node's current received-vector state (ReceivedThetaIdx/
// Set) to this node's own mover — the third-arrow twin of syncTiltIndex. Called by
// every site that changes those fields, below. nil-safe, same as syncTiltIndex.
func (n *Node) syncReceivedVector() {
	if n.SyncReceivedVector == nil {
		return
	}
	n.SyncReceivedVector(n.ReceivedThetaIdx, n.ReceivedSet)
}

// outgoingVector is what THIS node SENDS on VectorOut: its own coplanarNormal rotated
// 180° in θ. Node2 turns +180° (+12 steps of π/12); Node1 (its mirror package) turns
// −180° (−12 steps) — index arithmetic, never radians.
func (n *Node) outgoingVector() Wiring.TiltVectorMsg {
	norm := n.coplanarNormal()
	norm.ThetaIdx += 2 * Wiring.PerpendicularThetaIdx
	return norm
}

// stepFromVector is the ONE place the arrived
// direction's own value is consulted rather than just its arrival. TWO DOT PRODUCTS decide,
// and between them they decide BOTH questions — whether to move, and which way:
//
//   - arrived vector ACUTE with this node's own TOP tilt vector    -> step +1 (adds one step)
//   - arrived vector ACUTE with this node's own BOTTOM tilt vector -> step -1, the REVERSE
//   - neither acute (exactly perpendicular)                        -> no step, and the caller
//     sends nothing, which is how the exchange stops
//
// The two cases are mutually exclusive and jointly exhaustive apart from the perpendicular
// one: the bottom is the top's exact antipode, so the two dots are always exact negatives
// and at most one of them can be positive. There is no both-acute case for the ordering of
// these two ifs to arbitrate, and no free sign knob — which end the arrived vector leans
// toward IS the direction.
//
// Node2's base direction (the top-acute case) is adds one step; its mirror package's is the
// opposite, so a pair still turns symmetrically when both are leaning the same way.
//
// This does NOT consult Wiring.PerpendicularThetaIdx: the dots are the whole gate, so a node
// sitting exactly at the perpendicular index still steps if what arrived leans one way or
// the other. (The retired bead-path rule did consult it — and turned
// this node in a fixed direction regardless of what arrived, which is what made a pair
// march one way forever.)
func (n *Node) stepFromVector(received Wiring.TiltVectorMsg) bool {
	switch {
	case Wiring.TiltVectorIsAcute(received, n.topTilt()):
		n.TopTiltThetaIdx += 1
	case Wiring.TiltVectorIsAcute(received, n.bottomTilt()):
		n.TopTiltThetaIdx -= 1
	default:
		return false
	}
	return true
}

// topTilt is THIS node's own top tilt vector as a direction pair — the same indices held on
// the struct, named so the dot products above read as vector-against-vector rather than as
// two loose ints.
func (n *Node) topTilt() Wiring.TiltVectorMsg {
	return Wiring.TiltVectorMsg{ThetaIdx: n.TopTiltThetaIdx}
}

// handleVectorCycle is Node2's WHOLE per-cycle vector-channel loop body: read
// VectorIn non-blocking; if something arrived, decide (stepFromVector);
// and if that moved this node's own tilt, send outgoingVector back out on VectorOut —
// also non-blocking. At the perpendicular index nothing steps and nothing sends,
// which is how the exchange stops. This never touches In/Out or beads — the vector
// channel is a separate, additive exchange.
func (n *Node) handleVectorCycle(tick int64) {
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
	n.ReceivedSet = true
	n.syncReceivedVector()
	// DIAGNOSTIC (task/log-pair-vector-exchange) — mirror of Node1's, see there for why the
	// operands are captured BEFORE the step.
	before := n.TopTiltThetaIdx
	acuteTop := Wiring.TiltVectorIsAcute(received, n.topTilt())
	acuteBottom := Wiring.TiltVectorIsAcute(received, n.bottomTilt())
	moved := n.stepFromVector(received)
	if n.Self != nil {
		n.Self.Breadcrumb("pair-vector", fmt.Sprintf(
			"tick=%d recv=%d top=%d bottom=%d acuteTop=%v acuteBottom=%v moved=%v idx %d->%d",
			tick, received.ThetaIdx, before, before+Wiring.HalfTurnThetaIdx,
			acuteTop, acuteBottom, moved, before, n.TopTiltThetaIdx))
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

		// Drain TiltEditIn non-blocking: a panel/RESET/START edit — see the package doc
		// comment for the three-way split. applyTiltEdit decides placeBead: true only for
		// Start ("THE KICK"), false for a plain adjust (index-only, no send) and for Reset.
		if n.TiltEditIn != nil {
			select {
			case edit := <-n.TiltEditIn:
				placeBead := n.applyTiltEdit(edit)
				n.syncTiltIndex()
				if placeBead && n.Out != nil {
					n.Out.PlaceDrivenAt(1, clk.Tick())
				}
				// DIAGNOSTIC — mirror of Node1's; see there.
				if n.Self != nil {
					n.Self.Breadcrumb("pair-tilt-edit", fmt.Sprintf(
						"tick=%d reset=%v start=%v up=%v placeBead=%v theta=%d",
						clk.Tick(), edit.Reset, edit.Start, edit.Up, placeBead,
						n.TopTiltThetaIdx))
				}
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
			n.TopTiltThetaIdx = a.TiltVectorAngleSeed()
			n.TiltEditIn = a.TiltEditIn()
			// Self replaces the old SyncTiltIndex/SyncReceivedVector/ClearOutBeads
			// messages-to-a-separate-mover-goroutine (task/pair-node-owns-itself):
			// this node's own goroutine now owns that mover state directly, so what
			// used to be a message is a plain method call on the same object below.
			self := a.ClaimSelfDrive()
			n.Self = self
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
