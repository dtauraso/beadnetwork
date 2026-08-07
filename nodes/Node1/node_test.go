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
// what turned a pair one way forever. The bead now only paces; stepFromVector's acute tests are
// the only rule that turns a tilt on an arrival, and they are asserted below.

// applyTiltEdit is what the RESET button, the START TILT button, and the tilt-angle panel
// (TiltVectorButtons.tsx / TiltVectorAnglePanel.tsx) each drive; this asserts THIS node
// kind's own decision for a reset, from EVERY starting index on the ring: both indices land
// on 0 (the start position), and — unlike Start — no bead is placed ("the kick" only fires
// for Start, never a stop-and-return).
func TestApplyTiltEditResetReturnsBothIndicesToZero(t *testing.T) {
	for start := int32(0); start < defaultRing.points; start++ {
		n := &Node{Top: defaultRing.at(start)}
		placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
		if placeBead {
			t.Fatalf("reset from theta=%d must place no bead, got placeBead=true", start)
		}
		if n.topState().idx != 0 {
			t.Fatalf("reset from theta=%d: want index 0, got theta=%d", start, n.topState().idx)
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

	// The step: an arrival ON this node's own top turns it one way, an arrival on its bottom
	// the other, so these two drive the index round in opposite directions.
	for _, onTop := range []bool{true, false} {
		n := &Node{Top: defaultRing.at(0)}
		for i := 0; i < 3*int(full); i++ {
			// The arrival is stated relative to this node's CURRENT tilt, so it keeps
			// leaning the same way as the index walks.
			var arrivalIdx int32
			if onTop {
				arrivalIdx = n.topState().idx
			} else {
				arrivalIdx = n.topState().opposite.idx
			}
			v := Wiring.TiltVectorMsg{ThetaIdx: arrivalIdx}
			if !n.stepFromVector(v) {
				t.Fatalf("arrival onTop=%v, move %d: expected a step, got the halt", onTop, i)
			}
			if n.topState().idx < 0 || n.topState().idx >= full {
				t.Fatalf("arrival onTop=%v, move %d: index left the circle: %d",
					onTop, i, n.topState().idx)
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
// exactly one step and does NOTHING else: no bead, no send on VectorOut
// (task/pair-node-owns-itself split — this used to also open the vector exchange as a side
// effect, "the kick", which made one click move the tilt by many π/12 steps once the
// exchange settled).
func TestApplyTiltEditAdjustMovesOneStepAndSendsNothing(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{Top: defaultRing.at(3), VectorOut: out}
	placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Up: true})
	if placeBead {
		t.Fatalf("a plain adjust must place NO bead, got placeBead=true")
	}
	if n.topState().idx != 4 {
		t.Fatalf("adjust theta up: want theta=4, got theta=%d", n.topState().idx)
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

// stepFromVector's TWO acute tests decide both whether to move and WHICH WAY. Leaning toward this
// node's own TOP tilt vector takes Node1's base direction; leaning toward its BOTTOM tilt
// vector takes the reverse. These assert one goroutine's own arithmetic, no channel
// involved (docs/testing-shape.md).
func TestStepFromVectorTakesBaseDirectionWhenAcuteWithTop(t *testing.T) {
	// Tilt at index 0 points at world +y; an arrival at index 0 is the same direction, so
	// it is 0 steps from the top (acute) and a half turn from the bottom (not acute).
	n := &Node{Top: defaultRing.at(0)}
	if moved := n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: 0}); !moved || n.topState().idx != 1 {
		t.Fatalf("acute with the TOP tilt: want moved=true thetaIdx=1, got moved=%v thetaIdx=%d",
			moved, n.topState().idx)
	}
}

func TestStepFromVectorReversesWhenAcuteWithBottom(t *testing.T) {
	// A half turn from the tilt: now it is 0 steps from the BOTTOM (acute) and a half turn
	// from the top (not acute), so the step must go the OTHER way from the case above.
	//
	// Stepping down from 0 lands on 23, not −1: the index is kept ON THE CIRCLE at every site
	// that moves it, so it never runs negative. Same drawn direction, one step back from +y.
	n := &Node{Top: defaultRing.at(0)}
	want := defaultRing.points - 1
	if moved := n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: defaultRing.halfTurn}); !moved || n.topState().idx != want {
		t.Fatalf("acute with the BOTTOM tilt: want moved=true thetaIdx=%d, got moved=%v thetaIdx=%d",
			want, moved, n.topState().idx)
	}
}

// Exactly perpendicular to both the top and bottom tilt is the HALT case: stepFromVector
// steps nothing and reports moved=false — this is how the vector exchange comes to rest.
func TestStepFromVectorHaltsWhenNeitherDotIsAcute(t *testing.T) {
	n := &Node{Top: defaultRing.at(defaultRing.quarterTurn)}
	before := n.topState().idx
	perp := Wiring.TiltVectorMsg{ThetaIdx: n.topState().idx + defaultRing.quarterTurn}
	if moved := n.stepFromVector(perp); moved {
		t.Fatal("stepFromVector must report moved=false on a perpendicular arrival, got true")
	}
	if n.topState().idx != before {
		t.Fatalf("a perpendicular arrival must step NOTHING; got %d, want unchanged %d",
			n.topState().idx, before)
	}
}

// stepFromVector's three cases: acute with top steps +1 and returns true, acute with
// bottom steps -1 and returns true, and exactly perpendicular to both steps nothing and
// returns false. Consolidates the three single-case tests above into one table asserting the
// full gate.
func TestStepFromVectorGatesOnBothDotsForAllThreeCases(t *testing.T) {
	cases := []struct {
		name        string
		arrivedIdx  int32
		wantMoved   bool
		wantDeltaTh int32
	}{
		{"acute with top", 0, true, 1},
		// Down from 0 lands on 23, not −1 — the index stays on the circle.
		{"acute with bottom", defaultRing.halfTurn, true, defaultRing.points - 1},
		{"exactly perpendicular", defaultRing.quarterTurn, false, 0},
	}
	for _, c := range cases {
		n := &Node{Top: defaultRing.at(0)}
		moved := n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: c.arrivedIdx})
		if moved != c.wantMoved {
			t.Fatalf("%s: want moved=%v, got %v", c.name, c.wantMoved, moved)
		}
		if n.topState().idx != c.wantDeltaTh {
			t.Fatalf("%s: want thetaIdx=%d, got %d", c.name, c.wantDeltaTh, n.topState().idx)
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
// decision steps anything. This is the case that matters: a perpendicular arrival must still
// be recorded as the third drawn vector even though stepFromVector halts and nothing is sent
// back — recording the arrival and halting the exchange are independent rules. Only a RESET
// removes the recorded direction (asserted separately, below).
func TestReceivedVectorRecordedButExchangeHaltsOnPerpendicularArrival(t *testing.T) {
	n := &Node{Top: defaultRing.at(0), ReceivedThetaIdx: 4, ReceivedSet: true}
	in := make(chan Wiring.TiltVectorMsg, 1)
	out := make(chan Wiring.TiltVectorMsg, 1)
	n.VectorIn, n.VectorOut = in, out
	arrived := Wiring.TiltVectorMsg{ThetaIdx: defaultRing.quarterTurn}
	in <- arrived

	n.handleVectorCycle(0)

	if !n.ReceivedSet || n.ReceivedThetaIdx != arrived.ThetaIdx {
		t.Fatalf("the arrived direction must be recorded even though it halts; got set=%v theta=%d",
			n.ReceivedSet, n.ReceivedThetaIdx)
	}
	if n.topState().idx != 0 {
		t.Fatalf("a perpendicular arrival must step NOTHING; got %d, want unchanged 0", n.topState().idx)
	}
	select {
	case <-out:
		t.Fatal("a perpendicular arrival must halt the exchange, so no reply may be sent")
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
// actually steps this node. A leaning arrival steps and places a bead; a perpendicular
// arrival halts (stepFromVector returns false) and places NO bead — this is how the bead
// exchange comes to rest alongside the vector exchange.
func TestBeadIsPlacedOnlyWhenVectorStepsNotOnPerpendicularHalt(t *testing.T) {
	ctx := context.Background()
	pw := wire.NewPacedWire(1, 1.0)
	out := wire.NewPacedOutNoGeom(pw, ctx, "Node1", "Out", nil, wire.RuleFireAndForget, 1, "")
	in := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{Top: defaultRing.at(0), VectorIn: in, Out: out}

	// An arrival that LEANS: the acute tests move this node, so a bead goes out with the reply.
	in <- Wiring.TiltVectorMsg{ThetaIdx: 0}
	n.handleVectorCycle(1)
	pw.DriveOneCycle(ctx, 2)
	if _, _, ok := pw.RecvTick(); !ok {
		t.Fatal("a vector step must place its own bead; nothing was placed")
	}

	// An arrival that is exactly PERPENDICULAR: this halts (steps nothing) and must place NO
	// bead across several subsequent cycles.
	n.Top = defaultRing.at(0)
	in <- Wiring.TiltVectorMsg{ThetaIdx: defaultRing.quarterTurn}
	n.handleVectorCycle(3)
	for tick := int64(4); tick < 10; tick++ {
		pw.DriveOneCycle(ctx, tick)
		if _, _, ok := pw.RecvTick(); ok {
			t.Fatal("a perpendicular arrival must halt and place no bead, but one was placed")
		}
	}
}

// acuteWith is the ring-walk reachability check that replaced Wiring.TiltVectorIsAcute's
// dot-product-sign arithmetic. This pins that the two agree at EVERY pair on the ring, not
// just the handful of cases exercised above: for every (a, b) in 0…FullTurnThetaIdx-1, the
// walk must report acute exactly when the arithmetic separation
// ((a-b) mod full + full) mod full is strictly less than a quarter turn (6) or strictly
// greater than three quarters of a turn (18) — the two ways of being "within a quarter turn
// either direction" of each other. A wrong hop count in acuteWith's loop bound, or an
// off-by-one in next/prev, shows up as a mismatch at the boundary pairs (separation exactly
// 6 or exactly 18) well before it would show up in the handful of indices the tests above
// exercise, since those never probe every boundary at once.
func TestAcuteWithAgreesWithTheArithmeticRuleAcrossTheWholeRing(t *testing.T) {
	full := defaultRing.points
	quarter := defaultRing.quarterTurn
	threeQuarters := full - quarter
	for a := int32(0); a < full; a++ {
		for b := int32(0); b < full; b++ {
			sep := ((a-b)%full + full) % full
			want := sep < quarter || sep > threeQuarters
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
			want := sep < quarter || sep > threeQuarters
			got := r.at(a).acuteWith(r.at(b))
			if got != want {
				t.Fatalf("points=12 a=%d b=%d separation=%d: want acuteWith=%v, got %v", a, b, sep, want, got)
			}
		}
	}
}
