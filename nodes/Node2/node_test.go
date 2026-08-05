package Node2

import (
	"context"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// The straightening loop's whole reactive rule now lives on THIS node kind's own
// goroutine (stepTilt) — this asserts what that ONE goroutine's own
// method decides, per docs/testing-shape.md: index arithmetic only, no mover/second
// goroutine involved.

// Node2 always ADDS one step of π/12, whichever side of perpendicular it is on — the
// direction is a property of the KIND, not of where the tilt currently sits. Its partner
// moves the opposite way by the same step, so a pair turns symmetrically.
func TestStepAlwaysMovesTheSameDirection(t *testing.T) {
	below := &Node{TiltThetaIdx: 3}
	if moved := below.stepTilt(); !moved || below.TiltThetaIdx != 4 {
		t.Fatalf("from below perpendicular, want moved=true thetaIdx=4, got moved=%v thetaIdx=%d", moved, below.TiltThetaIdx)
	}

	above := &Node{TiltThetaIdx: 9}
	if moved := above.stepTilt(); !moved || above.TiltThetaIdx != 10 {
		t.Fatalf("from above perpendicular, want moved=true thetaIdx=10, got moved=%v thetaIdx=%d", moved, above.TiltThetaIdx)
	}
}

// AT perpendicular, a call changes nothing and reports no move — this is the loop's
// termination, not a missed case.
func TestStepStopsAtPerpendicular(t *testing.T) {
	n := &Node{TiltThetaIdx: Wiring.PerpendicularThetaIdx}
	if moved := n.stepTilt(); moved {
		t.Fatalf("at perpendicular, want moved=false, got true (index now %d)", n.TiltThetaIdx)
	}
	if n.TiltThetaIdx != Wiring.PerpendicularThetaIdx {
		t.Fatalf("at perpendicular, index must not move; got %d, want %d", n.TiltThetaIdx, Wiring.PerpendicularThetaIdx)
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
		n := &Node{TiltThetaIdx: start.theta, TiltPhiIdx: start.phi}
		placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
		if placeBead {
			t.Fatalf("reset from theta=%d phi=%d must place no bead, got placeBead=true", start.theta, start.phi)
		}
		if n.TiltThetaIdx != 0 || n.TiltPhiIdx != 0 {
			t.Fatalf("reset from theta=%d phi=%d: want both indices 0, got theta=%d phi=%d", start.theta, start.phi, n.TiltThetaIdx, n.TiltPhiIdx)
		}
	}
}

// The drawn coplanar normal must sit exactly −6 steps (90°) in θ from the tilt, φ
// unchanged, for Node2 — pure index arithmetic reusing Wiring.PerpendicularThetaIdx, not
// a cross product (Part 1 of this task). Node2's sign is the mirror of Node1's.
func TestCoplanarNormalIsMinusSixStepsInTheta(t *testing.T) {
	for _, start := range []struct{ theta, phi int32 }{
		{0, 0}, {5, -3}, {-9, 12},
	} {
		n := &Node{TiltThetaIdx: start.theta, TiltPhiIdx: start.phi}
		norm := n.coplanarNormal()
		if norm.ThetaIdx != start.theta-Wiring.PerpendicularThetaIdx {
			t.Fatalf("coplanarNormal theta: want tilt-%d=%d, got %d", Wiring.PerpendicularThetaIdx, start.theta-Wiring.PerpendicularThetaIdx, norm.ThetaIdx)
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
	n := &Node{TiltThetaIdx: 5, TiltPhiIdx: -2, VectorIn: vectorIn}
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
	n := &Node{TiltThetaIdx: 3, TiltPhiIdx: 1}
	placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Axis: "theta", Up: true})
	if !placeBead {
		t.Fatalf("adjust must place a bead, got placeBead=false")
	}
	if n.TiltThetaIdx != 4 || n.TiltPhiIdx != 1 {
		t.Fatalf("adjust theta up: want theta=4 phi=1, got theta=%d phi=%d", n.TiltThetaIdx, n.TiltPhiIdx)
	}
}

// outgoingVector reverses the coplanar normal by 180° in θ — Node2's own direction is
// +12 index steps (+180°, two of the 6-step quarter turns) — and leaves φ untouched.
// This is THIS node's own arithmetic, asserted without any channel involved.
func TestOutgoingVectorIsPlus180StepsInThetaOnly(t *testing.T) {
	n := &Node{TiltThetaIdx: 3, TiltPhiIdx: -7}
	norm := n.coplanarNormal()
	// Node2 SUBTRACTS the quarter turn (Node1 adds) — see coplanarNormal's own doc
	// comment: the tilt and the coplanar normal sit −90° apart in θ for Node2.
	if norm.ThetaIdx != 3-Wiring.PerpendicularThetaIdx || norm.PhiIdx != -7 {
		t.Fatalf("coplanarNormal: want theta=%d phi=-7, got theta=%d phi=%d",
			3-Wiring.PerpendicularThetaIdx, norm.ThetaIdx, norm.PhiIdx)
	}
	out := n.outgoingVector()
	wantTheta := norm.ThetaIdx + 2*Wiring.PerpendicularThetaIdx
	if out.ThetaIdx != wantTheta {
		t.Fatalf("outgoingVector theta: want %d (norm + 12 steps), got %d", wantTheta, out.ThetaIdx)
	}
	if out.PhiIdx != norm.PhiIdx {
		t.Fatalf("outgoingVector must leave phi unchanged: norm phi=%d, out phi=%d", norm.PhiIdx, out.PhiIdx)
	}
}

// stepTowardPerpendicularFromVector is Node2's own step decision on a vector-channel
// arrival: away from perpendicular it moves one π/12 step (same direction stepTilt
// always takes for this kind), and reports the move.
func TestStepFromVectorMovesOneStepAwayFromPerpendicular(t *testing.T) {
	n := &Node{TiltThetaIdx: 3}
	received := Wiring.TiltVectorMsg{ThetaIdx: 99, PhiIdx: -5}
	if moved := n.stepTowardPerpendicularFromVector(received); !moved || n.TiltThetaIdx != 4 {
		t.Fatalf("want moved=true thetaIdx=4, got moved=%v thetaIdx=%d", moved, n.TiltThetaIdx)
	}
}

// AT perpendicular, an arriving vector still steps nothing and the caller must send
// nothing — this is the exchange's stop condition, asserted at the decision method
// itself, not by observing two goroutines communicate.
func TestStepFromVectorStopsAtPerpendicular(t *testing.T) {
	n := &Node{TiltThetaIdx: Wiring.PerpendicularThetaIdx}
	received := Wiring.TiltVectorMsg{ThetaIdx: 1, PhiIdx: 1}
	if moved := n.stepTowardPerpendicularFromVector(received); moved {
		t.Fatalf("at perpendicular, want moved=false, got true (index now %d)", n.TiltThetaIdx)
	}
	if n.TiltThetaIdx != Wiring.PerpendicularThetaIdx {
		t.Fatalf("at perpendicular, index must not move; got %d, want %d", n.TiltThetaIdx, Wiring.PerpendicularThetaIdx)
	}
}

// A received reset zeroes this node and REPLIES WITH NOTHING — a reply would bounce the
// reset back and forth forever instead of ending the exchange.
func TestReceivedResetZeroesAndDoesNotReply(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	in := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{TiltThetaIdx: 5, TiltPhiIdx: 3, VectorOut: out, VectorIn: in}

	in <- Wiring.TiltVectorMsg{Reset: true}
	n.handleVectorCycle()

	if n.TiltThetaIdx != 0 || n.TiltPhiIdx != 0 {
		t.Fatalf("a received reset must zero both indices; got theta=%d phi=%d", n.TiltThetaIdx, n.TiltPhiIdx)
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
	n := &Node{TiltThetaIdx: 2, VectorIn: in}

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
	n := &Node{TiltThetaIdx: 2, VectorIn: in}

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

// The third arrow is drawn only while the exchange is RUNNING. An arrival that finds this
// node already at the perpendicular index — where it steps nothing and sends nothing, which
// IS the exchange stopping — clears it instead of showing itself. The drawing stops with the
// exchange rather than leaving its last frame on screen.
func TestReceivedVectorVanishesWhenTheExchangeStops(t *testing.T) {
	n := &Node{TiltThetaIdx: Wiring.PerpendicularThetaIdx, ReceivedThetaIdx: 4, ReceivedPhiIdx: 1, ReceivedSet: true}
	in := make(chan Wiring.TiltVectorMsg, 1)
	n.VectorIn = in
	in <- Wiring.TiltVectorMsg{ThetaIdx: 7, PhiIdx: 2}

	n.handleVectorCycle()

	if n.ReceivedSet || n.ReceivedThetaIdx != 0 || n.ReceivedPhiIdx != 0 {
		t.Fatalf("at perpendicular the third arrow must vanish; got set=%v theta=%d phi=%d",
			n.ReceivedSet, n.ReceivedThetaIdx, n.ReceivedPhiIdx)
	}
}

// This node's own LOCAL reset (applyTiltEdit's Reset branch) clears the received-vector
// record too — a stale received arrow left hanging would contradict the reset.
func TestApplyTiltEditResetClearsReceivedVector(t *testing.T) {
	n := &Node{TiltThetaIdx: 5, ReceivedThetaIdx: 9, ReceivedPhiIdx: -1, ReceivedSet: true}
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
	n := &Node{TiltThetaIdx: 5, VectorOut: out, VectorIn: in,
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

// Update must report this node's OPENING tilt/normal pair BEFORE its loop body runs —
// same reason as Node1's twin of this test: the mover mirrors these and cannot derive the
// normal itself, so staying silent until the first arrival leaves the mover's normal at
// zero, decoding to world +y alongside the opening tilt and superimposing the two drawn
// arrows. Single goroutine: Update runs on an already-cancelled context.
func TestUpdateSyncsOpeningTiltIndexBeforeLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var gotTheta, gotPhi, gotNormalTheta, gotNormalPhi int32
	calls := 0
	n := &Node{TiltThetaIdx: 4, TiltPhiIdx: -2}
	n.SyncTiltIndex = func(theta, phi, normalTheta, normalPhi int32) {
		calls++
		gotTheta, gotPhi, gotNormalTheta, gotNormalPhi = theta, phi, normalTheta, normalPhi
	}

	n.Update(ctx)

	if calls != 1 {
		t.Fatalf("want exactly one opening sync, got %d", calls)
	}
	if gotTheta != 4 || gotPhi != -2 {
		t.Fatalf("opening tilt: want (4,-2), got (%d,%d)", gotTheta, gotPhi)
	}
	wantNormalTheta := int32(4) - Wiring.PerpendicularThetaIdx
	if gotNormalTheta != wantNormalTheta || gotNormalPhi != -2 {
		t.Fatalf("opening normal: want (%d,-2), got (%d,%d)", wantNormalTheta, gotNormalTheta, gotNormalPhi)
	}
}
