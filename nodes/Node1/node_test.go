package Node1

import (
	"context"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// The straightening loop's whole reactive rule now lives on THIS node kind's own
// goroutine (stepTilt) — this asserts what that ONE goroutine's own
// method decides, per docs/testing-shape.md: index arithmetic only, no mover/second
// goroutine involved.

// Node1 always SUBTRACTS one step of π/12, whichever side of perpendicular it is on — the
// direction is a property of the KIND, not of where the tilt currently sits. Its partner
// moves the opposite way by the same step, so a pair turns symmetrically.
func TestStepAlwaysMovesTheSameDirection(t *testing.T) {
	below := &Node{TopTiltThetaIdx: 3}
	if moved := below.stepTilt(); !moved || below.TopTiltThetaIdx != 2 {
		t.Fatalf("from below perpendicular, want moved=true thetaIdx=2, got moved=%v thetaIdx=%d", moved, below.TopTiltThetaIdx)
	}

	above := &Node{TopTiltThetaIdx: 9}
	if moved := above.stepTilt(); !moved || above.TopTiltThetaIdx != 8 {
		t.Fatalf("from above perpendicular, want moved=true thetaIdx=8, got moved=%v thetaIdx=%d", moved, above.TopTiltThetaIdx)
	}
}

// AT perpendicular, a call changes nothing and reports no move — this is the loop's
// termination, not a missed case.
func TestStepStopsAtPerpendicular(t *testing.T) {
	n := &Node{TopTiltThetaIdx: Wiring.PerpendicularThetaIdx}
	if moved := n.stepTilt(); moved {
		t.Fatalf("at perpendicular, want moved=false, got true (index now %d)", n.TopTiltThetaIdx)
	}
	if n.TopTiltThetaIdx != Wiring.PerpendicularThetaIdx {
		t.Fatalf("at perpendicular, index must not move; got %d, want %d", n.TopTiltThetaIdx, Wiring.PerpendicularThetaIdx)
	}
}

// applyTiltEdit is what the RESET button (TiltResetButton.tsx) and the tilt-angle panel
// both drive; this asserts THIS node kind's own decision for a reset, from any starting
// indices: both indices land on 0 (the start position), and — unlike an adjust — no bead
// is placed ("the kick" only fires for a panel nudge, never a stop-and-return).
func TestApplyTiltEditResetReturnsBothIndicesToZero(t *testing.T) {
	for _, start := range []struct{ theta, phi int32 }{
		{0, 0},
		{5, -3},
		{-9, 12},
		{Wiring.PerpendicularThetaIdx, 4},
	} {
		n := &Node{TopTiltThetaIdx: start.theta, TopTiltPhiIdx: start.phi}
		placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
		if placeBead {
			t.Fatalf("reset from theta=%d phi=%d must place no bead, got placeBead=true", start.theta, start.phi)
		}
		if n.TopTiltThetaIdx != 0 || n.TopTiltPhiIdx != 0 {
			t.Fatalf("reset from theta=%d phi=%d: want both indices 0, got theta=%d phi=%d", start.theta, start.phi, n.TopTiltThetaIdx, n.TopTiltPhiIdx)
		}
	}
}

// The drawn coplanar normal must sit exactly +6 steps (90°) in θ from the tilt, φ
// unchanged, for Node1 — pure index arithmetic reusing Wiring.PerpendicularThetaIdx, not
// a cross product (Part 1 of this task).
func TestCoplanarNormalIsPlusSixStepsInTheta(t *testing.T) {
	for _, start := range []struct{ theta, phi int32 }{
		{0, 0}, {5, -3}, {-9, 12},
	} {
		n := &Node{TopTiltThetaIdx: start.theta, TopTiltPhiIdx: start.phi}
		norm := n.coplanarNormal()
		if norm.ThetaIdx != start.theta+Wiring.PerpendicularThetaIdx {
			t.Fatalf("coplanarNormal theta: want tilt+%d=%d, got %d", Wiring.PerpendicularThetaIdx, start.theta+Wiring.PerpendicularThetaIdx, norm.ThetaIdx)
		}
		if norm.PhiIdx != start.phi {
			t.Fatalf("coplanarNormal phi must equal tilt phi unchanged: want %d, got %d", start.phi, norm.PhiIdx)
		}
	}
}

// RESET must also drain any value already sitting on VectorIn (depth-1 latest-wins), so
// it cannot arrive on the NEXT cycle and immediately step the tilt again, undoing the
// reset. Verified by observing the channel is empty afterward.
func TestApplyTiltEditResetDrainsVectorIn(t *testing.T) {
	vectorIn := make(chan Wiring.TiltVectorMsg, 1)
	vectorIn <- Wiring.TiltVectorMsg{ThetaIdx: 99, PhiIdx: -1}
	n := &Node{TopTiltThetaIdx: 5, TopTiltPhiIdx: -2, VectorIn: vectorIn}
	n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
	select {
	case v := <-vectorIn:
		t.Fatalf("VectorIn must be drained by reset; still holds %+v", v)
	default:
	}
}

// A non-reset edit (the panel's ±1 click) still unconditionally applies its delta and
// reports a bead should be placed — the RESET addition must not change this existing path.
func TestApplyTiltEditAdjustStillPlacesBead(t *testing.T) {
	n := &Node{TopTiltThetaIdx: 3, TopTiltPhiIdx: 1}
	placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Axis: "theta", Up: true})
	if !placeBead {
		t.Fatalf("adjust must place a bead, got placeBead=false")
	}
	if n.TopTiltThetaIdx != 4 || n.TopTiltPhiIdx != 1 {
		t.Fatalf("adjust theta up: want theta=4 phi=1, got theta=%d phi=%d", n.TopTiltThetaIdx, n.TopTiltPhiIdx)
	}
}

// outgoingVector reverses the coplanar normal by 180° in θ — Node1's own direction is
// −12 index steps (−180°, two of the 6-step quarter turns) — and leaves φ untouched.
// This is THIS node's own arithmetic, asserted without any channel involved.
func TestOutgoingVectorIsMinus180StepsInThetaOnly(t *testing.T) {
	n := &Node{TopTiltThetaIdx: 3, TopTiltPhiIdx: -7}
	norm := n.coplanarNormal()
	if norm.ThetaIdx != 3+Wiring.PerpendicularThetaIdx || norm.PhiIdx != -7 {
		t.Fatalf("coplanarNormal: want theta=%d phi=-7, got theta=%d phi=%d",
			3+Wiring.PerpendicularThetaIdx, norm.ThetaIdx, norm.PhiIdx)
	}
	out := n.outgoingVector()
	wantTheta := norm.ThetaIdx - 2*Wiring.PerpendicularThetaIdx
	if out.ThetaIdx != wantTheta {
		t.Fatalf("outgoingVector theta: want %d (norm - 12 steps), got %d", wantTheta, out.ThetaIdx)
	}
	if out.PhiIdx != norm.PhiIdx {
		t.Fatalf("outgoingVector must leave phi unchanged: norm phi=%d, out phi=%d", norm.PhiIdx, out.PhiIdx)
	}
}

// stepTowardPerpendicularFromVector is Node1's own step decision on a vector-channel
// arrival: away from perpendicular it moves one π/12 step (same direction stepTilt
// always takes for this kind), and reports the move.
func TestStepFromVectorMovesOneStepAwayFromPerpendicular(t *testing.T) {
	n := &Node{TopTiltThetaIdx: 3}
	received := Wiring.TiltVectorMsg{ThetaIdx: 99, PhiIdx: -5}
	if moved := n.stepTowardPerpendicularFromVector(received); !moved || n.TopTiltThetaIdx != 2 {
		t.Fatalf("want moved=true thetaIdx=2, got moved=%v thetaIdx=%d", moved, n.TopTiltThetaIdx)
	}
}

// AT perpendicular, an arriving vector still steps nothing and the caller must send
// nothing — this is the exchange's stop condition, asserted at the decision method
// itself, not by observing two goroutines communicate.
func TestStepFromVectorStopsAtPerpendicular(t *testing.T) {
	n := &Node{TopTiltThetaIdx: Wiring.PerpendicularThetaIdx}
	received := Wiring.TiltVectorMsg{ThetaIdx: 1, PhiIdx: 1}
	if moved := n.stepTowardPerpendicularFromVector(received); moved {
		t.Fatalf("at perpendicular, want moved=false, got true (index now %d)", n.TopTiltThetaIdx)
	}
	if n.TopTiltThetaIdx != Wiring.PerpendicularThetaIdx {
		t.Fatalf("at perpendicular, index must not move; got %d, want %d", n.TopTiltThetaIdx, Wiring.PerpendicularThetaIdx)
	}
}

// A RESET empties BOTH directions — but by each node draining the one receive end it owns,
// not by either of them evicting the other's queue. A send-only end cannot drain itself (in
// Go only a receiver empties a channel), so this asserts the pair-level property: reset both
// nodes and nothing is left anywhere.
func TestResettingBothNodesEmptiesBothDirections(t *testing.T) {
	oneToTwo := make(chan Wiring.TiltVectorMsg, 1)
	twoToOne := make(chan Wiring.TiltVectorMsg, 1)
	// Node1 sends on oneToTwo and receives on twoToOne; its partner is the mirror image.
	one := &Node{TopTiltThetaIdx: 4, TopTiltPhiIdx: 2, VectorOut: oneToTwo, VectorIn: twoToOne}
	partnerIn := oneToTwo // what the other node owns the receive end of

	// A stale direction is in flight BOTH ways when reset is pressed.
	oneToTwo <- Wiring.TiltVectorMsg{ThetaIdx: 9}
	twoToOne <- Wiring.TiltVectorMsg{ThetaIdx: 9}

	one.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
	if one.TopTiltThetaIdx != 0 || one.TopTiltPhiIdx != 0 {
		t.Fatalf("reset must zero both indices; got theta=%d phi=%d", one.TopTiltThetaIdx, one.TopTiltPhiIdx)
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
	n := &Node{TopTiltThetaIdx: 5, TopTiltPhiIdx: 3, VectorOut: out, VectorIn: in}

	in <- Wiring.TiltVectorMsg{Reset: true}
	n.handleVectorCycle()

	if n.TopTiltThetaIdx != 0 || n.TopTiltPhiIdx != 0 {
		t.Fatalf("a received reset must zero both indices; got theta=%d phi=%d", n.TopTiltThetaIdx, n.TopTiltPhiIdx)
	}
	if v, ok := Wiring.PollRecvVector(out); ok {
		t.Fatalf("a received reset must not be replied to; got %+v", v)
	}
}

// handleVectorCycle stores the received direction as this node's own third drawn vector
// — asserted at the decision method itself (docs/testing-shape.md), never by observing a
// mover or a second goroutine. Perpendicular so stepTilt itself does not fire; the record
// must still happen (the third arrow shows the last ARRIVAL, not the last arrival that
// moved something). It shows only while the exchange is RUNNING.
func TestHandleVectorCycleRecordsReceivedDirection(t *testing.T) {
	in := make(chan Wiring.TiltVectorMsg, 1)
	// MID-exchange: not at the perpendicular index, where an arrival would instead clear
	// the third arrow because the exchange has stopped.
	n := &Node{TopTiltThetaIdx: 2, VectorIn: in}

	in <- Wiring.TiltVectorMsg{ThetaIdx: 7, PhiIdx: -4}
	n.handleVectorCycle()

	if !n.ReceivedSet {
		t.Fatal("an arrival must set ReceivedSet, got false")
	}
	if n.ReceivedThetaIdx != 7 || n.ReceivedPhiIdx != -4 {
		t.Fatalf("want recorded theta=7 phi=-4, got theta=%d phi=%d", n.ReceivedThetaIdx, n.ReceivedPhiIdx)
	}
}

// A second arrival REPLACES the first, never accumulates.
func TestHandleVectorCycleReplacesPreviousReceivedDirection(t *testing.T) {
	in := make(chan Wiring.TiltVectorMsg, 1)
	// MID-exchange, so both arrivals are recorded rather than clearing (see above).
	n := &Node{TopTiltThetaIdx: 2, VectorIn: in}

	in <- Wiring.TiltVectorMsg{ThetaIdx: 7, PhiIdx: -4}
	n.handleVectorCycle()
	in <- Wiring.TiltVectorMsg{ThetaIdx: -2, PhiIdx: 11}
	n.handleVectorCycle()

	if n.ReceivedThetaIdx != -2 || n.ReceivedPhiIdx != 11 {
		t.Fatalf("want the LATEST arrival theta=-2 phi=11, got theta=%d phi=%d", n.ReceivedThetaIdx, n.ReceivedPhiIdx)
	}
	if !n.ReceivedSet {
		t.Fatal("ReceivedSet must stay true across a replace")
	}
}

// The third arrow STAYS until the next arrival replaces it. An arrival is recorded even
// when it does not move this node — including the case where the step decision declines and
// the exchange comes to rest — because the last direction a node was sent is what it is
// still holding. Only a RESET removes it (asserted separately, below).
func TestReceivedVectorRecordedEvenWhenNothingSteps(t *testing.T) {
	// Perpendicular to this node's own tilt axis: neither dot is acute, so nothing steps
	// and nothing is sent — the exchange comes to rest on this arrival.
	n := &Node{TopTiltThetaIdx: 0, ReceivedThetaIdx: 4, ReceivedPhiIdx: 1, ReceivedSet: true}
	in := make(chan Wiring.TiltVectorMsg, 1)
	out := make(chan Wiring.TiltVectorMsg, 1)
	n.VectorIn, n.VectorOut = in, out
	arrived := Wiring.TiltVectorMsg{ThetaIdx: Wiring.PerpendicularThetaIdx, PhiIdx: 0}
	in <- arrived

	n.handleVectorCycle()

	if !n.ReceivedSet || n.ReceivedThetaIdx != arrived.ThetaIdx || n.ReceivedPhiIdx != arrived.PhiIdx {
		t.Fatalf("the arrived direction must be recorded even when nothing steps; got set=%v theta=%d phi=%d",
			n.ReceivedSet, n.ReceivedThetaIdx, n.ReceivedPhiIdx)
	}
	if n.TopTiltThetaIdx != 0 {
		t.Fatalf("neither dot is acute here, so the tilt must not move; got %d", n.TopTiltThetaIdx)
	}
	select {
	case v := <-out:
		t.Fatalf("nothing must be sent when nothing steps; got %+v", v)
	default:
	}
}

// This node's own LOCAL reset (applyTiltEdit's Reset branch) clears the received-vector
// record too — a stale received arrow left hanging would contradict the reset.
func TestApplyTiltEditResetClearsReceivedVector(t *testing.T) {
	n := &Node{TopTiltThetaIdx: 5, ReceivedThetaIdx: 9, ReceivedPhiIdx: -1, ReceivedSet: true}
	n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})

	if n.ReceivedSet {
		t.Fatal("local reset must clear ReceivedSet, got true")
	}
	if n.ReceivedThetaIdx != 0 || n.ReceivedPhiIdx != 0 {
		t.Fatalf("local reset must zero the received indices; got theta=%d phi=%d", n.ReceivedThetaIdx, n.ReceivedPhiIdx)
	}
}

// A Reset marker ARRIVING on the channel clears the received-vector record too, same as
// the local reset — verified alongside TestReceivedResetZeroesAndDoesNotReply's existing
// tilt-index assertion.
func TestHandleVectorCycleReceivedResetClearsReceivedVector(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	in := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{TopTiltThetaIdx: 5, VectorOut: out, VectorIn: in,
		ReceivedThetaIdx: 9, ReceivedPhiIdx: -1, ReceivedSet: true}

	in <- Wiring.TiltVectorMsg{Reset: true}
	n.handleVectorCycle()

	if n.ReceivedSet {
		t.Fatal("a received reset must clear ReceivedSet, got true")
	}
	if n.ReceivedThetaIdx != 0 || n.ReceivedPhiIdx != 0 {
		t.Fatalf("a received reset must zero the received indices; got theta=%d phi=%d", n.ReceivedThetaIdx, n.ReceivedPhiIdx)
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

	var gotTheta, gotPhi, gotNormalTheta, gotNormalPhi, gotBottomTheta, gotBottomPhi int32
	calls := 0
	n := &Node{TopTiltThetaIdx: 4, TopTiltPhiIdx: -2}
	n.SyncTiltIndex = func(theta, phi, normalTheta, normalPhi, bottomTheta, bottomPhi int32) {
		calls++
		gotTheta, gotPhi, gotNormalTheta, gotNormalPhi = theta, phi, normalTheta, normalPhi
		gotBottomTheta, gotBottomPhi = bottomTheta, bottomPhi
	}

	n.Update(ctx)

	if calls != 1 {
		t.Fatalf("want exactly one opening sync, got %d", calls)
	}
	if gotTheta != 4 || gotPhi != -2 {
		t.Fatalf("opening tilt: want (4,-2), got (%d,%d)", gotTheta, gotPhi)
	}
	// The BOTTOM tilt rides the same opening sync, for the same reason: the mover cannot
	// derive the half turn's sign either, and a bottom left at zero would draw at world +y
	// on top of the opening top tilt.
	wantBottomTheta := int32(4) + Wiring.HalfTurnThetaIdx
	if gotBottomTheta != wantBottomTheta || gotBottomPhi != -2 {
		t.Fatalf("opening bottom tilt: want (%d,-2), got (%d,%d)", wantBottomTheta, gotBottomTheta, gotBottomPhi)
	}
	wantNormalTheta := int32(4) + Wiring.PerpendicularThetaIdx
	if gotNormalTheta != wantNormalTheta || gotNormalPhi != -2 {
		t.Fatalf("opening normal: want (%d,-2), got (%d,%d)", wantNormalTheta, gotNormalTheta, gotNormalPhi)
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
	n := &Node{TopTiltThetaIdx: 5, In: wire.NewInChan(beads, "n1", "In", nil, nil)}

	n.clear()

	if _, ok := n.In.PollRecv(); ok {
		t.Fatal("clear must leave In empty; a bead survived and would restart the exchange")
	}
	if n.TopTiltThetaIdx != 0 {
		t.Fatalf("clear must zero the tilt index, got %d", n.TopTiltThetaIdx)
	}
}

// The beads still CROSSING this node's outgoing wires are not this goroutine's to drop —
// a PacedWire is driven by its source node's own mover. So clear asks (ClearOutBeads),
// and this asserts the ask, which is the whole of what this goroutine decides here.
func TestClearAsksTheMoverToEmptyItsOutgoingWires(t *testing.T) {
	asked := 0
	n := &Node{TopTiltThetaIdx: 5, ClearOutBeads: func() { asked++ }}

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
	n := &Node{TopTiltThetaIdx: 5, ReceivedThetaIdx: 9, ReceivedSet: true,
		In: wire.NewInChan(beads, "n1", "In", nil, nil), VectorIn: in,
		ClearOutBeads: func() { asked++ }}

	in <- Wiring.TiltVectorMsg{Reset: true}
	n.handleVectorCycle()

	if _, ok := n.In.PollRecv(); ok {
		t.Fatal("a received reset marker must drain this node's delivered beads too")
	}
	if asked != 1 {
		t.Fatalf("a received reset marker must ask the mover to empty the outgoing wires, got %d asks", asked)
	}
	if n.TopTiltThetaIdx != 0 || n.ReceivedSet {
		t.Fatalf("a received reset marker must zero the tilt and clear the third arrow; got theta=%d set=%v", n.TopTiltThetaIdx, n.ReceivedSet)
	}
}

// The panel click is the vector exchange's OPENING MOVE, not just the bead's. Without it
// nothing is ever sent on the channel that is not a reply to an arrival, so no arrival ever
// happens, no node records a received direction, and the third arrow cannot be drawn
// anywhere. Asserts what this ONE goroutine's own method emits, per docs/testing-shape.md.
func TestPanelAdjustOpensTheVectorExchange(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{TopTiltThetaIdx: 3, VectorOut: out}

	n.applyTiltEdit(Wiring.TiltEditMsg{Axis: "theta", Up: true})

	select {
	case got := <-out:
		if got.Reset {
			t.Fatal("a panel adjust must send a DIRECTION, not a reset marker")
		}
		// The direction must be derived from the index this click produced (4), not the
		// one it replaced (3) — otherwise the partner aims at a tilt that no longer exists.
		if want := n.outgoingVector(); got != want {
			t.Fatalf("sent %+v, want this node's post-click outgoing vector %+v", got, want)
		}
	default:
		t.Fatal("a panel adjust sent nothing on VectorOut; the exchange has no opening move and the third arrow can never appear")
	}
}

// A RESET is the opposite: it sends the Reset MARKER, never a direction. A direction here
// would restart the very exchange the reset exists to end.
func TestResetSendsAMarkerNotADirection(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{TopTiltThetaIdx: 3, VectorOut: out}

	n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})

	got := <-out
	if !got.Reset {
		t.Fatalf("reset must send the Reset marker, got a direction %+v", got)
	}
}
