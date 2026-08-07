package Node1

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// The pair's whole reactive rule lives on THIS node kind's own goroutine — this asserts
// what that ONE goroutine's own methods decide, per docs/testing-shape.md: index arithmetic
// only, no mover/second goroutine involved.
//
// The old bead-path step tests are gone with that rule itself: a bead arrival used to step this
// node one click in the kind's own fixed direction, regardless of what arrived, which is
// what turned a pair one way forever. The bead now only paces; stepFromVector is the only
// rule that turns a tilt on an arrival, and it is asserted below.

// applyTiltEdit is what the RESET button, the START TILT button, and the tilt-angle panel
// (TiltVectorButtons.tsx / TiltVectorAnglePanel.tsx) each drive; this asserts THIS node
// kind's own decision for a reset, from EVERY starting index on the ring: both indices land
// on 0 (the start position), and — unlike Start — no bead is placed ("the kick" only fires
// for Start, never a stop-and-return).
func TestApplyTiltEditResetReturnsBothIndicesToZero(t *testing.T) {
	for start := int32(0); start < defaultRing.points; start++ {
		// Held on purpose: a reset must take it too. It marks whose angle is intent for an
		// exchange that has not run yet, and one surviving a reset would silently decide who
		// moves the next time one starts.
		n := &Node{Top: defaultRing.at(start), Held: true}
		placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
		if placeBead {
			t.Fatalf("reset from theta=%d must place no bead, got placeBead=true", start)
		}
		if n.topState().idx != 0 {
			t.Fatalf("reset from theta=%d: want index 0, got theta=%d", start, n.topState().idx)
		}
		if n.Held {
			t.Fatalf("reset from theta=%d must clear the hold; a reset leaves nothing in the pair", start)
		}
	}
}

// The drawn coplanar normal must sit exactly +6 steps (90°) in θ from the tilt — a link on
// the ring, reusing Wiring.PerpendicularThetaIdx, not a cross product. These indices are far
// enough from the ring's end that the link lands without coming round, so the raw sum is the
// answer; the version that covers every index, wrap included, is the next test.
func TestCoplanarNormalIsPlusSixStepsInTheta(t *testing.T) {
	for _, theta := range []int32{0, 5, 9} {
		n := &Node{Top: defaultRing.at(theta)}
		norm := n.coplanarNormal()
		if norm.ThetaIdx != theta+defaultRing.quarterTurn {
			t.Fatalf("coplanarNormal theta: want tilt+%d=%d, got %d", defaultRing.quarterTurn, theta+defaultRing.quarterTurn, norm.ThetaIdx)
		}
		// topTilt is the stored index NAMED as a direction — the acute tests read it as one
		// operand, so it must be exactly the index and never a derived value.
		if got := n.topTilt().ThetaIdx; got != theta {
			t.Fatalf("topTilt must be the stored index itself: want %d, got %d", theta, got)
		}
	}
}

// coplanarNormal is ONE QUARTER TURN past the tilt index, the same way round at every index —
// no bucket, no crossing count, no parity. Asserted as the SEPARATION between the two
// directions rather than as a formula, so it holds whether or not the addition wrapped: the
// wrap takes off a full turn, which is not a change of direction.
//
// EVERY state on the ring is checked, not a chosen few. A tilt is one of exactly
// FullTurnThetaIdx directions and cannot be anything else, so the whole domain is small
// enough to walk — including both sides of every multiple of a half turn, which is where a
// reintroduced parity term would flip the normal to the other side of the tilt.
//
// Indices off the ring are not in this list because they are not tilts: a state is the only
// thing a tilt can be, and the ring has no such member for the node to hold.
func TestCoplanarNormalIsOneQuarterTurnPastTheTiltEverywhere(t *testing.T) {
	full := defaultRing.points
	quarter := defaultRing.quarterTurn
	for theta := int32(0); theta < full; theta++ {
		n := &Node{Top: defaultRing.at(theta)}
		norm := n.coplanarNormal()
		// The separation must be the quarter turn AHEAD (6), never the quarter turn behind
		// (18) — the latter is what the removed parity term produced for half of the range.
		if d := ((norm.ThetaIdx-theta)%full + full) % full; d != quarter {
			t.Fatalf("theta=%d: normal must sit %d steps past the tilt, got %d (normal=%d)",
				theta, quarter, d, norm.ThetaIdx)
		}
	}
}

// EVERY INDEX STAYS ON THE RING, whatever a user or an exchange does to it. A tilt only ever
// moves by following a link, so this should be true by construction; it is asserted anyway,
// because "by construction" is a claim about every site that moves the tilt, and this checks
// them rather than asking a reader to.
//
// Walks the index all the way round in both directions with the panel's ±1, then all the way
// round again with the acute-test step, checking the range after every single move. A missed
// wrap at either boundary shows up on the first lap.
//
// The step half's arrival is no longer built with the deleted onCircle helper: an arrival ON
// this node's own top is n.topState().idx itself, and an arrival on its bottom is
// n.topState().opposite.idx — both are already ring-valued states, read straight off the
// ring rather than reduced by hand.
func TestTheTiltIndexNeverLeavesTheCircle(t *testing.T) {
	full := defaultRing.points

	for _, up := range []bool{true, false} {
		n := &Node{Top: defaultRing.at(0)}
		for i := 0; i < 3*int(full); i++ {
			n.applyTiltEdit(Wiring.TiltEditMsg{Up: up})
			if n.topState().idx < 0 || n.topState().idx >= full {
				t.Fatalf("panel up=%v, move %d: index left the circle: %d", up, i, n.topState().idx)
			}
		}
	}

	// The step: an arrival LEANING toward this node's own top turns it one way, an arrival
	// leaning toward its bottom the other, so these two drive the index round in opposite
	// directions. Exactly on the top or exactly on the bottom is now a halt (gap 0), not a
	// lean, so each arrival is one step PAST the relevant state — genuinely acute — and
	// recomputed relative to the CURRENT top every iteration so it keeps leaning the same way
	// as the index walks.
	for _, towardTop := range []bool{true, false} {
		n := &Node{Top: defaultRing.at(0)}
		for i := 0; i < 3*int(full); i++ {
			var arrivalIdx int32
			if towardTop {
				arrivalIdx = n.topState().next.idx
			} else {
				arrivalIdx = n.topState().opposite.next.idx
			}
			v := Wiring.TiltVectorMsg{ThetaIdx: arrivalIdx}
			if !n.stepFromVector(v) {
				t.Fatalf("arrival towardTop=%v, move %d: expected a step, got the halt", towardTop, i)
			}
			if n.topState().idx < 0 || n.topState().idx >= full {
				t.Fatalf("arrival towardTop=%v, move %d: index left the circle: %d",
					towardTop, i, n.topState().idx)
			}
		}
	}
}

// The wrap is a COMPARISON — one test, one subtraction, no division — so a tilt already on
// the circle gives a normal on the circle rather than an index of 24 or more.
func TestCoplanarNormalWrapsOntoTheCircle(t *testing.T) {
	full := defaultRing.points
	for theta := int32(0); theta < full; theta++ {
		n := &Node{Top: defaultRing.at(theta)}
		if got := n.coplanarNormal().ThetaIdx; got < 0 || got >= full {
			t.Fatalf("theta=%d: want a normal in 0…%d, got %d", theta, full-1, got)
		}
	}
}

// RESET must also drain any value already sitting on VectorIn (depth-1 latest-wins), so
// it cannot arrive on the NEXT cycle and immediately step the tilt again, undoing the
// reset. Verified by observing the channel is empty afterward.
func TestApplyTiltEditResetDrainsVectorIn(t *testing.T) {
	vectorIn := make(chan Wiring.TiltVectorMsg, 1)
	vectorIn <- Wiring.TiltVectorMsg{ThetaIdx: 99}
	n := &Node{Top: defaultRing.at(5), VectorIn: vectorIn}
	n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
	select {
	case v := <-vectorIn:
		t.Fatalf("VectorIn must be drained by reset; still holds %+v", v)
	default:
	}
}

// A plain adjust (the panel's ▲/▼ click, neither Reset nor Start) moves the named index by
// exactly one step, marks this end HELD, and does NOTHING ELSE: no send, no bead.
//
// SETTING AN ANGLE AND RUNNING THE EXCHANGE ARE SEPARATE ACTS, and this test is what keeps
// them separate. Folding the send into a click puts "the kick" back inside an adjust, which
// is the side effect task/pair-node-owns-itself removed: opening the exchange from the click
// feeds arrivals straight back to the node just clicked, and one click moved that tilt by
// many steps. Held happens to mask that today, which is the reason to pin the split here
// rather than lean on it — the behaviour should not depend on a flag being set correctly.
//
// The hold is still set, because it says whose angle is intent when the exchange DOES run.
func TestApplyTiltEditAdjustMovesOneStepHoldsAndSendsNothing(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{Top: defaultRing.at(3), VectorOut: out}
	placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Up: true})
	if placeBead {
		t.Fatal("a plain adjust must place NO bead — START is what runs the exchange")
	}
	if n.topState().idx != 4 {
		t.Fatalf("adjust theta up: want theta=4, got theta=%d", n.topState().idx)
	}
	if !n.Held {
		t.Fatal("a plain adjust must mark this end Held: its angle is intent when the exchange runs")
	}
	select {
	case v := <-out:
		t.Fatalf("a plain adjust must send NOTHING on VectorOut; got %+v", v)
	default:
	}
}

// outgoingVector sends the coplanar normal — the direction this node computed and draws is
// the one that goes on the channel, unrotated. THIS node's own arithmetic, no channel
// involved.
//
// The equality against coplanarNormal is what catches a rotation added to this path; the
// separation check below is a second, independent statement of WHICH direction that is. It
// reduces onto the circle and demands exactly the quarter turn AHEAD — 18 (the quarter turn
// behind) and 30 (a reintroduced half-turn reversal) both fail it.
func TestOutgoingVectorIsTheCoplanarNormalUnchanged(t *testing.T) {
	// EVERY index on the ring is checked: a tilt is only ever one of the ring's own states
	// (a ▼ click follows n.top.prev, which cannot leave the ring), so there is no negative or
	// out-of-range stored index left to probe — the whole domain is walked instead.
	for theta := int32(0); theta < defaultRing.points; theta++ {
		n := &Node{Top: defaultRing.at(theta)}
		norm := n.coplanarNormal()
		out := n.outgoingVector()
		if out.ThetaIdx != norm.ThetaIdx {
			t.Fatalf("theta=%d: outgoingVector want %d (the normal itself), got %d", theta, norm.ThetaIdx, out.ThetaIdx)
		}
		full := defaultRing.points
		if d := ((out.ThetaIdx-n.topTilt().ThetaIdx)%full + full) % full; d != defaultRing.quarterTurn {
			t.Fatalf("theta=%d: what is sent must sit %d steps past the top, got %d",
				theta, defaultRing.quarterTurn, d)
		}
		// And the bottom tilt is the TOP's exact antipode, which is what makes the two acute
		// tests exact opposites of each other — the property the whole step rule rests on.
		// Both sides are now ring-valued (stateFor reduces theta onto the circle before this
		// node ever stores it), so the separation is asserted mod a full turn rather than as
		// a bare subtraction — the same wrap every other test in this file already applies.
		bottom := n.bottomTilt()
		if d := ((bottom.ThetaIdx-n.topTilt().ThetaIdx)%full + full) % full; d != defaultRing.halfTurn {
			t.Fatalf("theta=%d: bottom must sit a half turn from the top, got %d", theta, d)
		}
	}
}

// stepFromVector's DIRECTION RULE is distance, not a cone: from a given top, an arrival a
// few steps AHEAD (toward next) steps next, and one a few steps BEHIND (toward prev) steps
// prev — the shorter way round always wins. These assert one goroutine's own arithmetic, no
// channel involved (docs/testing-shape.md).
func TestStepFromVectorTakesBaseDirectionWhenAcuteWithTop(t *testing.T) {
	// Tilt at index 0 points at world +y; an arrival at index 1 is one step AHEAD of the top,
	// nearer via next than via prev, so this steps next.
	n := &Node{Top: defaultRing.at(0)}
	if moved := n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: 1}); !moved || n.topState().idx != 1 {
		t.Fatalf("arrival one step ahead: want moved=true thetaIdx=1, got moved=%v thetaIdx=%d",
			moved, n.topState().idx)
	}
}

// stepFromVector's direction the OTHER way: an arrival closer to the top going BACKWARD
// (via prev) steps prev, the reverse of the case above.
func TestStepFromVectorStepsPrevWhenArrivalIsBehind(t *testing.T) {
	// Two steps off the top going BACKWARD (index points-2): nearer via prev (gap 2) than via
	// next (gap points-2), so the step must go the OTHER way from the case above.
	//
	// Stepping down from 0 lands on points-1, not −1: the index is kept ON THE CIRCLE at every
	// site that moves it, so it never runs negative. Same drawn direction, one step back from +y.
	n := &Node{Top: defaultRing.at(0)}
	want := defaultRing.points - 1
	if moved := n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: defaultRing.points - 2}); !moved || n.topState().idx != want {
		t.Fatalf("arrival behind: want moved=true thetaIdx=%d, got moved=%v thetaIdx=%d",
			want, moved, n.topState().idx)
	}
}

// THE ONE HALT: an arrival landing EXACTLY on this node's own top is square (see
// stepFromVector's own doc comment) — stepFromVector steps nothing and reports moved=false.
// This is the ONLY way the vector exchange comes to rest for a correcting (not held) end;
// an arrival on the bottom, or anywhere else, now steps.
func TestStepFromVectorHaltsOnlyWhenArrivalIsExactlyOnTop(t *testing.T) {
	n := &Node{Top: defaultRing.at(defaultRing.quarterTurn)}
	before := n.topState().idx
	onTop := Wiring.TiltVectorMsg{ThetaIdx: before}
	if moved := n.stepFromVector(onTop); moved {
		t.Fatal("stepFromVector must report moved=false when the arrival lands on this node's own top, got true")
	}
	if n.topState().idx != before {
		t.Fatalf("an arrival on top must step NOTHING; got %d, want unchanged %d",
			n.topState().idx, before)
	}

	// An arrival a quarter turn from the top (what used to be the "perpendicular halt") is no
	// longer a halt for an end that is not held: it steps, same as every other non-top gap.
	n2 := &Node{Top: defaultRing.at(0)}
	perp := Wiring.TiltVectorMsg{ThetaIdx: defaultRing.quarterTurn}
	if moved := n2.stepFromVector(perp); !moved {
		t.Fatal("an arrival a quarter turn from the top must now step (colinear/perpendicular is no longer a resting state)")
	}
}

// EVERY STEP STRICTLY CLOSES THE GAP to the arrival — this is what replaced picking a
// direction from an acute cone, which said "within a quarter turn" but not which side and so
// could step away from the arrival and oscillate. Checked at several offsets, both ahead of
// and behind the top, walking one step at a time until the arrival is reached (arrival ==
// topState halts, per stepFromVector's own rule).
func TestStepFromVectorEachStepStrictlyReducesDistanceToArrival(t *testing.T) {
	full := defaultRing.points
	for _, offset := range []int32{1, 3, 6, 9, 12, 15, 18, 23} {
		n := &Node{Top: defaultRing.at(0)}
		arrival := defaultRing.at(offset)
		msg := Wiring.TiltVectorMsg{ThetaIdx: offset}
		prevGap := ringGap(n.topState(), arrival)
		for step := int32(0); step < full; step++ {
			moved := n.stepFromVector(msg)
			if !moved {
				if n.topState() != arrival {
					t.Fatalf("offset=%d: halted before reaching the arrival: at %d, want %d",
						offset, n.topState().idx, arrival.idx)
				}
				break
			}
			gap := ringGap(n.topState(), arrival)
			if gap >= prevGap {
				t.Fatalf("offset=%d step=%d: gap must strictly shrink, was %d now %d",
					offset, step, prevGap, gap)
			}
			prevGap = gap
			if step == full-1 {
				t.Fatalf("offset=%d: never reached the arrival within %d steps", offset, full)
			}
		}
	}
}

// A RESET empties BOTH directions — but by each node draining the one receive end it owns,
// not by either of them evicting the other's queue. A send-only end cannot drain itself (in
// Go only a receiver empties a channel), so this asserts the pair-level property: reset both
// nodes and nothing is left anywhere.
func TestResettingBothNodesEmptiesBothDirections(t *testing.T) {
	oneToTwo := make(chan Wiring.TiltVectorMsg, 1)
	twoToOne := make(chan Wiring.TiltVectorMsg, 1)
	// one sends on oneToTwo and receives on twoToOne; the other Node1 instance in the pair
	// runs the same code with the ends swapped.
	one := &Node{Top: defaultRing.at(4), VectorOut: oneToTwo, VectorIn: twoToOne}
	partnerIn := oneToTwo // what the other node owns the receive end of

	// A stale direction is in flight BOTH ways when reset is pressed.
	oneToTwo <- Wiring.TiltVectorMsg{ThetaIdx: 9}
	twoToOne <- Wiring.TiltVectorMsg{ThetaIdx: 9}

	one.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
	if one.topState().idx != 0 {
		t.Fatalf("reset must zero the index; got theta=%d", one.topState().idx)
	}
	if _, ok := Wiring.PollRecvVector(twoToOne); ok {
		t.Fatal("reset left a value on the end this node owns")
	}
	// The other direction is still full — this node cannot clear it, and that is the point:
	// the partner's own reset is what empties it, which the button sends too.
	if _, ok := Wiring.PollRecvVector(partnerIn); !ok {
		t.Fatal("expected the outward direction to still hold its stale value before the partner resets")
	}
	if _, ok := Wiring.PollRecvVector(partnerIn); ok {
		t.Fatal("after the partner's own drain, both directions must be empty")
	}
}

// A received reset zeroes this node and REPLIES WITH NOTHING — a reply would bounce the
// reset back and forth forever instead of ending the exchange.
func TestReceivedResetZeroesAndDoesNotReply(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	in := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{Top: defaultRing.at(5), VectorOut: out, VectorIn: in}

	in <- Wiring.TiltVectorMsg{Reset: true}
	n.handleVectorCycle(0)

	if n.topState().idx != 0 {
		t.Fatalf("a received reset must zero the index; got theta=%d", n.topState().idx)
	}
	if v, ok := Wiring.PollRecvVector(out); ok {
		t.Fatalf("a received reset must not be replied to; got %+v", v)
	}
}

// handleVectorCycle stores the received direction as this node's own third drawn vector
// — asserted at the decision method itself (docs/testing-shape.md), never by observing a
// mover or a second goroutine. Perpendicular so the step itself does not fire; the record
// must still happen (the third arrow shows the last ARRIVAL, not the last arrival that
// moved something). It shows only while the exchange is RUNNING.
func TestHandleVectorCycleRecordsReceivedDirection(t *testing.T) {
	in := make(chan Wiring.TiltVectorMsg, 1)
	// MID-exchange: not at the perpendicular index, where an arrival would instead clear
	// the third arrow because the exchange has stopped.
	n := &Node{Top: defaultRing.at(2), VectorIn: in}

	in <- Wiring.TiltVectorMsg{ThetaIdx: 7}
	n.handleVectorCycle(0)

	if !n.ReceivedSet {
		t.Fatal("an arrival must set ReceivedSet, got false")
	}
	if n.ReceivedThetaIdx != 7 {
		t.Fatalf("want recorded theta=7, got theta=%d", n.ReceivedThetaIdx)
	}
}

// A second arrival REPLACES the first, never accumulates.
func TestHandleVectorCycleReplacesPreviousReceivedDirection(t *testing.T) {
	in := make(chan Wiring.TiltVectorMsg, 1)
	// MID-exchange, so both arrivals are recorded rather than clearing (see above).
	n := &Node{Top: defaultRing.at(2), VectorIn: in}

	in <- Wiring.TiltVectorMsg{ThetaIdx: 7}
	n.handleVectorCycle(0)
	in <- Wiring.TiltVectorMsg{ThetaIdx: 15}
	n.handleVectorCycle(0)

	if n.ReceivedThetaIdx != 15 {
		t.Fatalf("want the LATEST arrival theta=15, got theta=%d", n.ReceivedThetaIdx)
	}
	if !n.ReceivedSet {
		t.Fatal("ReceivedSet must stay true across a replace")
	}
}

// The third arrow STAYS until the next arrival replaces it, and this is recorded
// UNCONDITIONALLY — before the step decision even runs, and independently of whether that
// decision steps anything. This is the case that matters: an arrival that lands on this
// node's own TOP (the one halt, square) must still be recorded as the third drawn vector
// even though stepFromVector halts and nothing is sent back — recording the arrival and
// halting the exchange are independent rules. Only a RESET removes the recorded direction
// (asserted separately, below).
func TestReceivedVectorRecordedButExchangeHaltsOnPerpendicularArrival(t *testing.T) {
	n := &Node{Top: defaultRing.at(0), ReceivedThetaIdx: 4, ReceivedSet: true}
	in := make(chan Wiring.TiltVectorMsg, 1)
	out := make(chan Wiring.TiltVectorMsg, 1)
	n.VectorIn, n.VectorOut = in, out
	// An arrival ON this node's own top: square, the one halt.
	arrived := Wiring.TiltVectorMsg{ThetaIdx: 0}
	in <- arrived

	n.handleVectorCycle(0)

	if !n.ReceivedSet || n.ReceivedThetaIdx != arrived.ThetaIdx {
		t.Fatalf("the arrived direction must be recorded even though it halts; got set=%v theta=%d",
			n.ReceivedSet, n.ReceivedThetaIdx)
	}
	if n.topState().idx != 0 {
		t.Fatalf("an arrival on top must step NOTHING; got %d, want unchanged 0", n.topState().idx)
	}
	select {
	case <-out:
		t.Fatal("an arrival on top must halt the exchange, so no reply may be sent")
	default:
	}
}

// This node's own LOCAL reset (applyTiltEdit's Reset branch) clears the received-vector
// record too — a stale received arrow left hanging would contradict the reset.
func TestApplyTiltEditResetClearsReceivedVector(t *testing.T) {
	n := &Node{Top: defaultRing.at(5), ReceivedThetaIdx: 9, ReceivedSet: true}
	n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})

	if n.ReceivedSet {
		t.Fatal("local reset must clear ReceivedSet, got true")
	}
	if n.ReceivedThetaIdx != 0 {
		t.Fatalf("local reset must zero the received index; got theta=%d", n.ReceivedThetaIdx)
	}
}

// A Reset marker ARRIVING on the channel clears the received-vector record too, same as
// the local reset — verified alongside TestReceivedResetZeroesAndDoesNotReply's existing
// tilt-index assertion.
func TestHandleVectorCycleReceivedResetClearsReceivedVector(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	in := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{Top: defaultRing.at(5), VectorOut: out, VectorIn: in,
		ReceivedThetaIdx: 9, ReceivedSet: true}

	in <- Wiring.TiltVectorMsg{Reset: true}
	n.handleVectorCycle(0)

	if n.ReceivedSet {
		t.Fatal("a received reset must clear ReceivedSet, got true")
	}
	if n.ReceivedThetaIdx != 0 {
		t.Fatalf("a received reset must zero the received index; got theta=%d", n.ReceivedThetaIdx)
	}
}

// Update must report this node's OPENING tilt/normal pair BEFORE its loop body runs. The
// mover mirrors these and cannot derive the normal itself, so a node that stays silent
// until its first arrival leaves the mover's normal at zero — where it decodes to world
// +y, the same direction the opening tilt decodes to, and the two drawn arrows
// superimpose. Single goroutine: Update is called directly on an already-cancelled
// context, so it emits the opening sync and returns without entering the loop.
func TestUpdateSyncsOpeningTiltIndexBeforeLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var gotTheta, gotNormalTheta, gotBottomTheta int32
	calls := 0
	n := &Node{Top: defaultRing.at(4)}
	n.SyncTiltIndex = func(theta, normalTheta, bottomTheta int32) {
		calls++
		gotTheta, gotNormalTheta = theta, normalTheta
		gotBottomTheta = bottomTheta
	}

	n.Update(ctx)

	if calls != 1 {
		t.Fatalf("want exactly one opening sync, got %d", calls)
	}
	if gotTheta != 4 {
		t.Fatalf("opening tilt: want 4, got %d", gotTheta)
	}
	// The BOTTOM tilt rides the same opening sync, for the same reason: the mover cannot
	// derive the half turn's sign either, and a bottom left at zero would draw at world +y
	// on top of the opening top tilt.
	wantBottomTheta := int32(4) + defaultRing.halfTurn
	if gotBottomTheta != wantBottomTheta {
		t.Fatalf("opening bottom tilt: want %d, got %d", wantBottomTheta, gotBottomTheta)
	}
	wantNormalTheta := int32(4) + defaultRing.quarterTurn
	if gotNormalTheta != wantNormalTheta {
		t.Fatalf("opening normal: want %d, got %d", wantNormalTheta, gotNormalTheta)
	}
}

// A reset is not "zero the indices" — it is "leave nothing that can restart the exchange".
// The bead edge is what has actually been turning these tilts, so clear must empty this
// node's own In of beads already delivered to it; anything left there arrives on the next
// cycle and steps the tilt straight back off zero, which is what a reset that visibly does
// not take looks like. One goroutine, its own port, per docs/testing-shape.md.
func TestClearDrainsDeliveredBeads(t *testing.T) {
	beads := make(chan int, 4)
	beads <- 1
	beads <- 1
	n := &Node{Top: defaultRing.at(5), In: wire.NewInChan(beads, "n1", "In", nil, nil)}

	n.clear()

	if _, ok := n.In.PollRecv(); ok {
		t.Fatal("clear must leave In empty; a bead survived and would restart the exchange")
	}
	if n.topState().idx != 0 {
		t.Fatalf("clear must zero the tilt index, got %d", n.topState().idx)
	}
}

// The beads still CROSSING this node's outgoing wires are not this goroutine's to drop —
// a PacedWire is driven by its source node's own mover. So clear asks (ClearOutBeads),
// and this asserts the ask, which is the whole of what this goroutine decides here.
func TestClearAsksTheMoverToEmptyItsOutgoingWires(t *testing.T) {
	asked := 0
	n := &Node{Top: defaultRing.at(5), ClearOutBeads: func() { asked++ }}

	n.clear()

	if asked != 1 {
		t.Fatalf("clear must ask the mover exactly once to empty this node's outgoing wires, got %d asks", asked)
	}
}

// Both routes into a reset run the SAME clear: the button (applyTiltEdit) and the partner's
// Reset marker (handleVectorCycle). The marker-driven one is the one that lands after the
// partner stopped placing, so it is the one that actually makes the pair quiescent — it
// must do the full clear, not just the index zeroing it used to do.
func TestReceivedResetMarkerRunsTheFullClear(t *testing.T) {
	beads := make(chan int, 4)
	beads <- 1
	asked := 0
	in := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{Top: defaultRing.at(5), ReceivedThetaIdx: 9, ReceivedSet: true,
		In: wire.NewInChan(beads, "n1", "In", nil, nil), VectorIn: in,
		ClearOutBeads: func() { asked++ }}

	in <- Wiring.TiltVectorMsg{Reset: true}
	n.handleVectorCycle(0)

	if _, ok := n.In.PollRecv(); ok {
		t.Fatal("a received reset marker must drain this node's delivered beads too")
	}
	if asked != 1 {
		t.Fatalf("a received reset marker must ask the mover to empty the outgoing wires, got %d asks", asked)
	}
	if n.topState().idx != 0 || n.ReceivedSet {
		t.Fatalf("a received reset marker must zero the tilt and clear the third arrow; got theta=%d set=%v", n.topState().idx, n.ReceivedSet)
	}
}

// START is the vector exchange's OPENING MOVE (task/pair-node-owns-itself split — this used
// to be a panel adjust's side effect). Without it nothing is ever sent on the channel that
// is not a reply to an arrival, so no arrival ever happens, no node records a received
// direction, and the third arrow cannot be drawn anywhere. It sends from whatever angles are
// CURRENTLY set and changes NO index. Asserts what this ONE goroutine's own method emits,
// per docs/testing-shape.md.
func TestStartOpensTheVectorExchangeWithoutChangingAnyIndex(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	// PairID 1: START is addressed by id, and only the node numbered 1 opens the exchange.
	n := &Node{PairID: 1, Top: defaultRing.at(3), VectorOut: out}

	placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Start: true})

	if !placeBead {
		t.Fatal("Start must place a bead, got placeBead=false")
	}
	if n.topState().idx != 3 {
		t.Fatalf("Start must change NO index; got theta=%d, want unchanged theta=3", n.topState().idx)
	}
	select {
	case got := <-out:
		if got.Reset {
			t.Fatal("Start must send a DIRECTION, not a reset marker")
		}
		if want := n.outgoingVector(); got != want {
			t.Fatalf("sent %+v, want this node's current outgoing vector %+v", got, want)
		}
	default:
		t.Fatal("Start sent nothing on VectorOut; the exchange has no opening move and the third arrow can never appear")
	}
}

// A RESET is the opposite: it sends the Reset MARKER, never a direction. A direction here
// would restart the very exchange the reset exists to end.
func TestResetSendsAMarkerNotADirection(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{Top: defaultRing.at(3), VectorOut: out}

	n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})

	got := <-out
	if !got.Reset {
		t.Fatalf("reset must send the Reset marker, got a direction %+v", got)
	}
}

// THE ID CHANGES NOTHING EXCEPT WHO ANSWERS START. Both ends run this one implementation
// unmodified, so at the same tilt index they derive the same directions and step the same
// way — which is what makes the shared implementation VISIBLE in the editor: end 2's
// coplanar normal points where end 1's does, instead of opposite it. EVERY index on the
// ring is checked for the derive half of this claim.
func TestBothIDsDeriveAndStepIdentically(t *testing.T) {
	for theta := int32(0); theta < defaultRing.points; theta++ {
		one := &Node{PairID: 1, Top: defaultRing.at(theta)}
		two := &Node{PairID: 2, Top: defaultRing.at(theta)}

		if one.coplanarNormal().ThetaIdx != two.coplanarNormal().ThetaIdx {
			t.Fatalf("theta=%d: the two ids must derive the SAME normal, got %d and %d",
				theta, one.coplanarNormal().ThetaIdx, two.coplanarNormal().ThetaIdx)
		}
		if one.bottomTilt().ThetaIdx != two.bottomTilt().ThetaIdx {
			t.Fatalf("theta=%d: the two ids must derive the SAME bottom, got %d and %d",
				theta, one.bottomTilt().ThetaIdx, two.bottomTilt().ThetaIdx)
		}
	}

	for _, arrival := range []int32{0, defaultRing.halfTurn} {
		one := &Node{PairID: 1}
		two := &Node{PairID: 2}
		one.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: arrival})
		two.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: arrival})
		if one.topState().idx != two.topState().idx {
			t.Fatalf("arrival=%d: the two ids must step the SAME way, got id1=%d id2=%d",
				arrival, one.topState().idx, two.topState().idx)
		}
	}
}

// START OPENS THE EXCHANGE FROM ID 1 ONLY. Opened from both ends, each end also replies to
// the other's opener in the same round — two exchanges through one pair of channels, which
// the user sees as the pair reaching its rest state and being kicked off it again, forever.
// The panel cannot make this decision: it posts START to every node row, holding no domain
// knowledge about which end is which.
func TestStartOpensTheExchangeFromPairIDOneOnly(t *testing.T) {
	for _, c := range []struct {
		id       int32
		wantSend bool
	}{{1, true}, {2, false}, {7, false}, {0, false}} {
		out := make(chan Wiring.TiltVectorMsg, 1)
		n := &Node{PairID: c.id, Top: defaultRing.at(3), VectorOut: out}
		placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Start: true})

		if placeBead != c.wantSend {
			t.Fatalf("id %d: placeBead want %v, got %v", c.id, c.wantSend, placeBead)
		}
		select {
		case v := <-out:
			if !c.wantSend {
				t.Fatalf("id %d must not open the exchange, but sent %+v", c.id, v)
			}
		default:
			if c.wantSend {
				t.Fatalf("id %d must open the exchange, but sent nothing", c.id)
			}
		}
		if n.topState().idx != 3 {
			t.Fatalf("id %d: START must change no index, got %d", c.id, n.topState().idx)
		}
	}
}

// The bead travels WITH the vector: it is placed by the vector branch only when the vector
// actually steps this node. A stepping arrival places a bead; an arrival landing exactly on
// this node's own TOP (square, the one halt) places NO bead — this is how the bead exchange
// comes to rest alongside the vector exchange.
func TestBeadIsPlacedOnlyWhenVectorStepsNotOnPerpendicularHalt(t *testing.T) {
	ctx := context.Background()
	pw := wire.NewPacedWire(1, 1.0)
	out := wire.NewPacedOutNoGeom(pw, ctx, "Node1", "Out", nil, wire.RuleFireAndForget, 1, "")
	in := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{Top: defaultRing.at(0), VectorIn: in, Out: out}

	// An arrival that STEPS: one step ahead of the top, so stepFromVector moves this node and
	// a bead goes out with the reply.
	in <- Wiring.TiltVectorMsg{ThetaIdx: 1}
	n.handleVectorCycle(1)
	pw.DriveOneCycle(ctx, 2)
	if _, _, ok := pw.RecvTick(); !ok {
		t.Fatal("a vector step must place its own bead; nothing was placed")
	}

	// An arrival that lands EXACTLY on this node's own top: square, the one halt, and must
	// place NO bead across several subsequent cycles.
	n.Top = defaultRing.at(0)
	in <- Wiring.TiltVectorMsg{ThetaIdx: 0}
	n.handleVectorCycle(3)
	for tick := int64(4); tick < 10; tick++ {
		pw.DriveOneCycle(ctx, tick)
		if _, _, ok := pw.RecvTick(); ok {
			t.Fatal("an arrival on top must halt and place no bead, but one was placed")
		}
	}
}

// acuteWith is the ring-walk reachability check that replaced Wiring.TiltVectorIsAcute's
// dot-product-sign arithmetic. This pins that the two agree at EVERY pair on the ring, not
// just the handful of cases exercised above: for every (a, b) in 0…FullTurnThetaIdx-1, the
// walk must report acute exactly when the arithmetic separation
// ((a-b) mod full + full) mod full is strictly greater than 0 AND strictly less than a
// quarter turn (6), or strictly greater than three quarters of a turn (18) — BOTH bounds
// open at every boundary: a separation of exactly 0, exactly a quarter turn either way, or
// exactly a half turn is NOT acute. A wrong hop count in acuteWith's loop bound, or an
// off-by-one in next/prev, shows up as a mismatch at the boundary pairs (separation exactly
// 0, 6, or 18) well before it would show up in the handful of indices the tests above
// exercise, since those never probe every boundary at once.
func TestAcuteWithAgreesWithTheArithmeticRuleAcrossTheWholeRing(t *testing.T) {
	full := defaultRing.points
	quarter := defaultRing.quarterTurn
	threeQuarters := full - quarter
	for a := int32(0); a < full; a++ {
		for b := int32(0); b < full; b++ {
			sep := ((a-b)%full + full) % full
			want := (sep > 0 && sep < quarter) || sep > threeQuarters
			got := defaultRing.at(a).acuteWith(defaultRing.at(b))
			if got != want {
				t.Fatalf("a=%d b=%d separation=%d: want acuteWith=%v, got %v", a, b, sep, want, got)
			}
		}
	}
}

// arrivedState PANICS on an index off the ring, and returns the plain ring member for every
// in-range one — it is the boundary for a direction ARRIVING ON THE VECTOR CHANNEL, where the
// sender is this same kind sending one of its own states, so an off-ring value is a defect in
// this package (or a foreign writer on a channel only this kind holds) rather than something to
// fold. Silently folding it would turn that bug into a direction some number of steps from the
// one that was actually sent — plausible, drawable, and undetectable, which is exactly the
// failure the ring exists to make impossible; panicking with the bad index in the message is
// what makes that defect loud instead.
func TestArrivedStatePanicsOffRingAndNotOnRing(t *testing.T) {
	full := defaultRing.points

	for idx := int32(0); idx < full; idx++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("arrivedState(%d): must not panic on an in-range index, got %v", idx, r)
				}
			}()
			s := defaultRing.arrivedState(idx)
			if s.idx != idx {
				t.Fatalf("arrivedState(%d): want idx=%d, got %d", idx, idx, s.idx)
			}
		}()
	}

	for _, idx := range []int32{-1, -24, -25, full, full + 1, 2 * full} {
		func() {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("arrivedState(%d): want a panic, got none", idx)
				}
				msg := fmt.Sprintf("%v", r)
				if !strings.Contains(msg, fmt.Sprintf("%d", idx)) {
					t.Fatalf("arrivedState(%d): panic message must name the bad index, got %q", idx, msg)
				}
				// IT MUST BE THIS FUNCTION'S OWN PANIC, not Go's array-bounds panic from
				// at(). Checking only for the index cannot tell them apart — Go's message
				// for at(24) is "index out of range [24] with length 24", which contains
				// the index too, so a bound loosened from >= to > still passed this test.
				// The named phrase is what makes the check specific to the invariant
				// rather than to any panic that happens to mention the number.
				if !strings.Contains(msg, "vector channel") {
					t.Fatalf("arrivedState(%d): want THIS function's own panic naming the invariant, got %q",
						idx, msg)
				}
			}()
			defaultRing.arrivedState(idx)
		}()
	}
}

// seedState maps the PERSISTED SEED by asking the ring which state carries that index. A
// number the ring has names that state exactly; a number it does not have names nothing, and
// loads at the origin having said so.
//
// WHY THE ORIGIN AND NOT A FOLD. An index of 30 on a 24-state ring folds to 6 — a quarter
// turn from where the file said, arrived at by arithmetic, and once drawn indistinguishable
// from a direction somebody chose. The origin is wrong in a way that is obvious on sight, and
// the reported flag is what says so out loud. This matters more the moment the lattice size
// is editable: a file saved at one size and opened at a smaller one names indices this ring
// does not have, and that is a routine case rather than a legacy one.
//
// The in-range half of this is the stronger claim: EVERY index the ring has must come back as
// the identical ring member, not an equal-valued copy, so pointer identity keeps meaning
// direction equality.
func TestSeedStateLoadsAKnownIndexExactlyAndAnUnknownOneAtTheOrigin(t *testing.T) {
	full := defaultRing.points

	for idx := int32(0); idx < full; idx++ {
		s, unknown := defaultRing.seedState(idx)
		if unknown {
			t.Fatalf("seedState(%d): the ring has this index, so it must not report unknown", idx)
		}
		if s != defaultRing.at(idx) {
			t.Fatalf("seedState(%d): want the identical ring member at(%d), got a different state", idx, idx)
		}
	}

	for _, idx := range []int32{-1, -24, -25, full, full + 1, full + 23, 2 * full} {
		s, unknown := defaultRing.seedState(idx)
		if !unknown {
			t.Fatalf("seedState(%d): the ring has no such index, so it must report unknown", idx)
		}
		if s != defaultRing.at(0) {
			t.Fatalf("seedState(%d): an index the ring does not have must load at the origin, got idx=%d",
				idx, s.idx)
		}
	}
}

// newRing PANICS on any count that is not a positive multiple of four, and does NOT panic
// on one that is — the quarter turn and half turn must be whole numbers of states, or the
// coplanar normal and the perpendicular halt name nothing (see newRing's own doc comment).
// The panic message is checked for a DISTINCTIVE PHRASE, not just the offending count: Go's
// own out-of-range panics also happen to mention a number, so a bound loosened elsewhere
// could still pass a check that only grepped for the digits (the same trap
// TestArrivedStatePanicsOffRingAndNotOnRing above already guards against for arrivedState).
func TestNewRingPanicsOffMultipleOfFourAndNotOn(t *testing.T) {
	for _, points := range []int32{0, 1, 5, 25, 26, -4} {
		func() {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("newRing(%d): want a panic, got none", points)
				}
				msg := fmt.Sprintf("%v", r)
				if !strings.Contains(msg, fmt.Sprintf("%d", points)) {
					t.Fatalf("newRing(%d): panic message must name the offending count, got %q", points, msg)
				}
				if !strings.Contains(msg, "multiple of four") {
					t.Fatalf("newRing(%d): panic message must name the multiple-of-four requirement, got %q", points, msg)
				}
			}()
			newRing(points)
		}()
	}

	for _, points := range []int32{4, 8, 12, 24, 36} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("newRing(%d): must not panic on a valid count, got %v", points, r)
				}
			}()
			newRing(points)
		}()
	}
}

// A RING OF ANY VALID COUNT IS INTERNALLY CONSISTENT, not just the default 24-point one:
// walking every state, next/prev must be exact inverses of each other, opposite must be
// exactly halfTurn steps FORWARD (not backward — the two only coincide by accident at a
// half turn), quarter must be exactly quarterTurn steps forward, and every state must point
// back at the ring that built it, which is what lets acuteWith read its own bounds off
// s.ring rather than needing them passed alongside. Checked at several counts because a
// per-node ring is a NEW capability this commit adds — a formula that happens to work only
// at 24 would not be caught by the default-ring tests above, which never build another size.
func TestRingIsInternallyConsistentAtEveryValidCount(t *testing.T) {
	for _, points := range []int32{8, 12, 24, 36} {
		r := newRing(points)
		for i := int32(0); i < points; i++ {
			s := r.at(i)
			if s.ring != r {
				t.Fatalf("points=%d idx=%d: state's ring must be the ring that built it", points, i)
			}
			if s.next.prev != s {
				t.Fatalf("points=%d idx=%d: next.prev must be s itself, got a different state", points, i)
			}
			if s.prev.next != s {
				t.Fatalf("points=%d idx=%d: prev.next must be s itself, got a different state", points, i)
			}
			wantOpposite := (i + r.halfTurn) % points
			if s.opposite != r.at(wantOpposite) {
				t.Fatalf("points=%d idx=%d: opposite must be %d steps forward (idx=%d), got a different state",
					points, i, r.halfTurn, wantOpposite)
			}
			wantQuarter := (i + r.quarterTurn) % points
			if s.quarter != r.at(wantQuarter) {
				t.Fatalf("points=%d idx=%d: quarter must be %d steps forward (idx=%d), got a different state",
					points, i, r.quarterTurn, wantQuarter)
			}
		}
	}
}

// acuteWith's rule must hold at a ring size OTHER than the default 24, because acuteWith
// reads its bounds off s.ring rather than a package constant — a version that secretly
// still read Wiring.PerpendicularThetaIdx (6, the default's quarter turn) would pass every
// test above (which never leaves the default ring) yet be wrong here, where a 12-point
// ring's quarter turn is 3, not 6. The expectation is computed with the same explicit
// wrapped-separation arithmetic TestAcuteWithAgreesWithTheArithmeticRuleAcrossTheWholeRing
// uses for the default ring, so both tests are pinning the identical rule at two counts.
func TestAcuteWithHoldsAtANonDefaultRingSize(t *testing.T) {
	r := newRing(12)
	quarter := r.quarterTurn // 3
	full := r.points         // 12
	threeQuarters := full - quarter
	for a := int32(0); a < full; a++ {
		for b := int32(0); b < full; b++ {
			sep := ((a-b)%full + full) % full
			want := (sep > 0 && sep < quarter) || sep > threeQuarters
			got := r.at(a).acuteWith(r.at(b))
			if got != want {
				t.Fatalf("points=12 a=%d b=%d separation=%d: want acuteWith=%v, got %v", a, b, sep, want, got)
			}
		}
	}
}

// ADOPTING A NEW POINT COUNT KEEPS THE INDEX, so a tilt at 6 is still at 6 and the drawn
// angle moves — 90° on a 24-point lattice, 180° on a 12-point one. That is what "the lattice
// changed underneath a direction" means here: the number a user set is kept, and what it
// means follows the new lattice.
//
// An index the new lattice does not have names nothing there, so that node opens at its
// origin rather than at some computed neighbour.
func TestAdoptLatticeKeepsTheIndexOrOpensAtTheOrigin(t *testing.T) {
	// Kept: 6 exists on both lattices.
	n := &Node{Ring: newRing(24)}
	n.Top = n.Ring.at(6)
	n.adoptLattice(12)
	if got := n.ringOf().points; got != 12 {
		t.Fatalf("adoptLattice(12): want a 12-point ring, got %d", got)
	}
	if got := n.topState().idx; got != 6 {
		t.Fatalf("adoptLattice(12): index 6 exists on the new lattice and must be kept, got %d", got)
	}
	// The state must belong to the NEW ring, not be a survivor of the old one — pointer
	// identity is direction equality only within one lattice.
	if n.topState().ring != n.Ring {
		t.Fatal("adoptLattice: the tilt must be a state of the node's own new ring")
	}

	// Not kept: 20 exists on 24 points and not on 12.
	m := &Node{Ring: newRing(24)}
	m.Top = m.Ring.at(20)
	m.adoptLattice(12)
	if got := m.topState().idx; got != 0 {
		t.Fatalf("adoptLattice(12) from index 20: the new lattice has no such index, so it must open at the origin, got %d", got)
	}
}

// A DIRECTION FROM THE OLD LATTICE IS DROPPED, not acted on and not fatal. The two ends of a
// pair adopt a new count at their own moments, so between them an index picked on the old
// lattice arrives here — where it names a different angle, or no state at all.
//
// Dropping it is what keeps arrivedState's panic meaning "this cannot happen" rather than
// "this happens whenever a user changes the count". Asserted through handleVectorCycle, the
// real path, rather than by calling the check directly.
func TestAnArrivalFromAnotherLatticeIsDropped(t *testing.T) {
	in := make(chan Wiring.TiltVectorMsg, 1)
	out := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{Ring: newRing(12), VectorIn: in, VectorOut: out}
	n.Top = n.Ring.at(3)

	// An index this node's own 12-point ring does not even have, stated as being from a
	// 24-point lattice. Acting on it would panic; folding it would turn the node toward an
	// angle nobody sent.
	in <- Wiring.TiltVectorMsg{ThetaIdx: 20, Points: 24}
	n.handleVectorCycle(1)

	if got := n.topState().idx; got != 3 {
		t.Fatalf("an arrival from another lattice must not turn this node: want index 3, got %d", got)
	}
	if n.ReceivedSet {
		t.Fatal("an arrival from another lattice must not be recorded as the received direction")
	}
	select {
	case v := <-out:
		t.Fatalf("an arrival from another lattice must not be replied to; sent %+v", v)
	default:
	}
}

// What a node SENDS names the lattice it was picked on, so the partner can tell an old
// direction from a current one. Without it the drop above cannot be decided at all.
func TestOutgoingVectorNamesItsOwnLattice(t *testing.T) {
	for _, points := range []int32{8, 12, 24, 36} {
		n := &Node{Ring: newRing(points)}
		n.Top = n.Ring.at(1)
		if got := n.outgoingVector().Points; got != points {
			t.Fatalf("a %d-point node must send Points=%d, got %d", points, points, got)
		}
	}
}

// The GEOMETRY converts a tilt-vector INDEX to a drawn angle every frame (2π/points per
// step, nodes/Wiring/node_geometry.go's writeStreamFrame) but does not itself decide the
// point count — that is this node's own scene setting (adoptLattice). So adopting a count
// must report it onward exactly once, the same "report, don't derive" shape SyncTiltIndex
// already uses (TestUpdateSyncsOpeningTiltIndexBeforeLoop, above): a mover that missed this
// call would keep converting every index against whatever count it last heard, drawing a
// scene that no longer matches the ring this node is actually running.
func TestAdoptLatticeReportsTheNewCountExactlyOnce(t *testing.T) {
	var got int32
	calls := 0
	n := &Node{Ring: newRing(24)}
	n.Top = n.Ring.at(6)
	n.SyncLatticePoints = func(points int32) {
		calls++
		got = points
	}

	n.adoptLattice(12)

	if calls != 1 {
		t.Fatalf("adoptLattice(12): want exactly one SyncLatticePoints call, got %d", calls)
	}
	if got != 12 {
		t.Fatalf("adoptLattice(12): reported %d, want 12", got)
	}
}

// A node built with no count set at all (this package's test-literal shape — every field
// left zero) runs the model's documented default, Wiring.FullTurnThetaIdx (24) — the same
// default nodeGeometry.latticePoints falls back to when SetLatticePoints is never called.
// Only adoptLattice reports a NEW count (above); a node's OPENING count is stated once at
// build time (BuildArgs, node.go), which a bare test-literal Node never runs — so this test
// pins the default via ringOf(), the one read every other test in this file already uses
// for "no Ring set", rather than asserting a build-time call this construction never makes.
func TestNodeWithNoRingSetRunsTheDefaultLatticeCount(t *testing.T) {
	n := &Node{}
	if got := n.ringOf().points; got != Wiring.FullTurnThetaIdx {
		t.Fatalf("a Node with no Ring set must run the default lattice count %d, got %d", Wiring.FullTurnThetaIdx, got)
	}
}

// A TILT A USER SET IS KEPT, and the pair goes back to square by moving the OTHER end.
//
// This is what the hold is for: a correction that undid the click would make the panel
// useless. The tilted end holds its own index and keeps answering; its partner walks until
// the arrival lands on its own top, which is the two tilts a quarter turn apart. The user's
// number survives, the relationship comes back, and the hold releases itself.
//
// Both ends are the SAME KIND running the SAME rule — the asymmetry is which end the user
// touched, which each end knows locally, not which end it is.
//
// Driven through the real methods, over every starting position, both click directions, and
// several click counts: a rule that holds for one index and one click is not a rule.
func TestAUserTiltIsKeptAndThePartnerRestoresSquare(t *testing.T) {
	full := defaultRing.points
	quarter := defaultRing.quarterTurn

	for base := int32(0); base < full; base++ {
		for _, clicks := range []int{1, 2, 3, 5} {
			for _, up := range []bool{true, false} {
				a := &Node{PairID: 1, Ring: newRing(full), VectorOut: make(chan Wiring.TiltVectorMsg, 1)}
				b := &Node{PairID: 2, Ring: newRing(full)}
				a.Top = a.Ring.at(base)
				b.Top = b.Ring.at((base + quarter) % full) // square to start

				for i := 0; i < clicks; i++ {
					a.applyTiltEdit(Wiring.TiltEditMsg{Up: up})
				}
				want := a.topState().idx
				if !a.Held {
					t.Fatalf("base=%d: a panel tilt must hold this end's own tilt", base)
				}

				msg := a.outgoingVector()
				to, from := b, a
				settled := false
				for round := 0; round < 400; round++ {
					if !to.stepFromVector(msg) {
						settled = true
						break
					}
					msg = to.outgoingVector()
					to, from = from, to
				}
				if !settled {
					t.Fatalf("base=%d clicks=%d up=%v: never came to rest", base, clicks, up)
				}
				if got := a.topState().idx; got != want {
					t.Fatalf("base=%d clicks=%d up=%v: the tilt a user set must survive; want %d, got %d",
						base, clicks, up, want, got)
				}
				sep := ((b.topState().idx-a.topState().idx)%full + full) % full
				if sep != quarter && sep != full-quarter {
					t.Fatalf("base=%d clicks=%d up=%v: want the tilts a quarter turn apart, got %d apart (a=%d b=%d)",
						base, clicks, up, sep, a.topState().idx, b.topState().idx)
				}
			}
		}
	}
}
