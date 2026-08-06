package Node2

import (
	"context"
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
// what turned a pair one way forever. The bead now only paces; stepFromVector's dots are
// the only rule that turns a tilt on an arrival, and they are asserted below.

// applyTiltEdit is what the RESET button, the START TILT button, and the tilt-angle panel
// (TiltVectorButtons.tsx / TiltVectorAnglePanel.tsx) each drive; this asserts THIS node
// kind's own decision for a reset, from any starting indices: both indices land on 0 (the
// start position), and — unlike Start — no bead is placed ("the kick" only fires for
// Start, never a stop-and-return).
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

// The drawn coplanar normal must sit exactly −6 steps (90°) in θ from the tilt, φ
// unchanged, for Node2 — pure index arithmetic reusing Wiring.PerpendicularThetaIdx, not
// a cross product (Part 1 of this task). Node2's sign is the mirror of Node1's.
func TestCoplanarNormalIsMinusSixStepsInTheta(t *testing.T) {
	for _, start := range []struct{ theta, phi int32 }{
		{0, 0}, {5, -3}, {-9, 12},
	} {
		n := &Node{TopTiltThetaIdx: start.theta, TopTiltPhiIdx: start.phi}
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
	n := &Node{TopTiltThetaIdx: 5, TopTiltPhiIdx: -2, VectorIn: vectorIn}
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
	n := &Node{TopTiltThetaIdx: 3, TopTiltPhiIdx: 1, VectorOut: out}
	placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Axis: "theta", Up: true})
	if placeBead {
		t.Fatalf("a plain adjust must place NO bead, got placeBead=true")
	}
	if n.TopTiltThetaIdx != 4 || n.TopTiltPhiIdx != 1 {
		t.Fatalf("adjust theta up: want theta=4 phi=1, got theta=%d phi=%d", n.TopTiltThetaIdx, n.TopTiltPhiIdx)
	}
	select {
	case v := <-out:
		t.Fatalf("a plain adjust must send NOTHING on VectorOut; got %+v", v)
	default:
	}
}

// outgoingVector reverses the coplanar normal by 180° in θ — Node2's own direction is
// +12 index steps (+180°, two of the 6-step quarter turns) — and leaves φ untouched.
// This is THIS node's own arithmetic, asserted without any channel involved.
func TestOutgoingVectorIsPlus180StepsInThetaOnly(t *testing.T) {
	n := &Node{TopTiltThetaIdx: 3, TopTiltPhiIdx: -7}
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

// stepFromVector's TWO dots decide both whether to move and WHICH WAY. Leaning toward this
// node's own TOP tilt vector takes Node2's base direction; leaning toward its BOTTOM tilt
// vector takes the reverse. These assert one goroutine's own arithmetic, no channel
// involved (docs/testing-shape.md).
func TestStepFromVectorTakesBaseDirectionWhenAcuteWithTop(t *testing.T) {
	// Tilt at index 0 points at world +y; an arrival at index 0 is the same direction, so
	// dot(arrived, top) = 1 (acute) and dot(arrived, bottom) = -1.
	n := &Node{TopTiltThetaIdx: 0}
	if moved := n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: 0}); !moved || n.TopTiltThetaIdx != 1 {
		t.Fatalf("acute with the TOP tilt: want moved=true thetaIdx=1, got moved=%v thetaIdx=%d",
			moved, n.TopTiltThetaIdx)
	}
}

func TestStepFromVectorReversesWhenAcuteWithBottom(t *testing.T) {
	// A half turn from the tilt: now dot(arrived, bottom) = 1 and dot(arrived, top) = -1,
	// so the step must go the OTHER way from the case above.
	n := &Node{TopTiltThetaIdx: 0}
	if moved := n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: Wiring.HalfTurnThetaIdx}); !moved || n.TopTiltThetaIdx != -1 {
		t.Fatalf("acute with the BOTTOM tilt: want moved=true thetaIdx=-1, got moved=%v thetaIdx=%d",
			moved, n.TopTiltThetaIdx)
	}
}

// Exactly perpendicular to the tilt axis is the ONE case neither dot claims — no step, and
// the caller must send nothing. This is the exchange's stop condition, and it is about the
// ARRIVED direction, not about this node sitting on any particular index: the node here is
// AT PerpendicularThetaIdx and still would have stepped had the arrival leaned either way.
func TestStepFromVectorStopsWhenNeitherDotIsAcute(t *testing.T) {
	n := &Node{TopTiltThetaIdx: Wiring.PerpendicularThetaIdx}
	perp := Wiring.TiltVectorMsg{ThetaIdx: n.TopTiltThetaIdx + Wiring.PerpendicularThetaIdx}
	if moved := n.stepFromVector(perp); moved {
		t.Fatalf("perpendicular arrival: want moved=false, got true (index now %d)", n.TopTiltThetaIdx)
	}
	if n.TopTiltThetaIdx != Wiring.PerpendicularThetaIdx {
		t.Fatalf("perpendicular arrival must not move the index; got %d, want %d",
			n.TopTiltThetaIdx, Wiring.PerpendicularThetaIdx)
	}
	// ...and the same node DOES step when the arrival leans, proving the stop above came
	// from the dots and not from where this node happens to sit.
	if moved := n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: n.TopTiltThetaIdx}); !moved {
		t.Fatal("a leaning arrival must still step a node sitting at the perpendicular index")
	}
}

// A received reset zeroes this node and REPLIES WITH NOTHING — a reply would bounce the
// reset back and forth forever instead of ending the exchange.
func TestReceivedResetZeroesAndDoesNotReply(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	in := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{TopTiltThetaIdx: 5, TopTiltPhiIdx: 3, VectorOut: out, VectorIn: in}

	in <- Wiring.TiltVectorMsg{Reset: true}
	n.handleVectorCycle(0)

	if n.TopTiltThetaIdx != 0 || n.TopTiltPhiIdx != 0 {
		t.Fatalf("a received reset must zero both indices; got theta=%d phi=%d", n.TopTiltThetaIdx, n.TopTiltPhiIdx)
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
	n := &Node{TopTiltThetaIdx: 2, VectorIn: in}

	in <- Wiring.TiltVectorMsg{ThetaIdx: 7, PhiIdx: -4}
	n.handleVectorCycle(0)

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
	n.handleVectorCycle(0)
	in <- Wiring.TiltVectorMsg{ThetaIdx: -2, PhiIdx: 11}
	n.handleVectorCycle(0)

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

	n.handleVectorCycle(0)

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
	n.handleVectorCycle(0)

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
	wantBottomTheta := int32(4) - Wiring.HalfTurnThetaIdx
	if gotBottomTheta != wantBottomTheta || gotBottomPhi != -2 {
		t.Fatalf("opening bottom tilt: want (%d,-2), got (%d,%d)", wantBottomTheta, gotBottomTheta, gotBottomPhi)
	}
	wantNormalTheta := int32(4) - Wiring.PerpendicularThetaIdx
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
	n := &Node{TopTiltThetaIdx: 5, In: wire.NewInChan(beads, "n2", "In", nil, nil)}

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
		In: wire.NewInChan(beads, "n2", "In", nil, nil), VectorIn: in,
		ClearOutBeads: func() { asked++ }}

	in <- Wiring.TiltVectorMsg{Reset: true}
	n.handleVectorCycle(0)

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

// START is the vector exchange's OPENING MOVE (task/pair-node-owns-itself split — this used
// to be a panel adjust's side effect). Without it nothing is ever sent on the channel that
// is not a reply to an arrival, so no arrival ever happens, no node records a received
// direction, and the third arrow cannot be drawn anywhere. It sends from whatever angles are
// CURRENTLY set and changes NO index. Asserts what this ONE goroutine's own method emits,
// per docs/testing-shape.md.
// TestStartIsIgnoredByThisKind: START belongs to Node1 alone. The button addresses every
// node the panel lists (TS holds no knowledge of which node is node 1), so a Start record
// DOES reach this kind — and must do nothing at all here. If both ends opened the exchange,
// each would also be answering the other's opener in the same round.
func TestStartIsIgnoredByThisKind(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{TopTiltThetaIdx: 3, TopTiltPhiIdx: -2, VectorOut: out}

	placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Start: true})

	if placeBead {
		t.Fatal("Start must place NO bead on this kind, got placeBead=true")
	}
	if n.TopTiltThetaIdx != 3 || n.TopTiltPhiIdx != -2 {
		t.Fatalf("Start must change NO index; got theta=%d phi=%d, want unchanged theta=3 phi=-2", n.TopTiltThetaIdx, n.TopTiltPhiIdx)
	}
	select {
	case got := <-out:
		t.Fatalf("Start must send NOTHING from this kind; got %+v — the exchange is opened from node 1's end only", got)
	default:
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

// The bead now travels WITH the vector: it is placed by the vector branch when the dots
// actually move this node, not by the bead branch on every round trip. So the bead loop
// lives and dies with the exchange it paces, instead of circulating on its own in this
// kind's fixed direction forever.
func TestBeadIsPlacedByTheVectorStepNotByABeadArrival(t *testing.T) {
	ctx := context.Background()
	pw := wire.NewPacedWire(1, 1.0)
	out := wire.NewPacedOutNoGeom(pw, ctx, "Node2", "Out", nil, wire.RuleFireAndForget, 1, "")
	in := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{TopTiltThetaIdx: 0, VectorIn: in, Out: out}

	// An arrival that LEANS: the dots move this node, so a bead goes out with the reply.
	in <- Wiring.TiltVectorMsg{ThetaIdx: 0}
	n.handleVectorCycle(1)
	pw.DriveOneCycle(ctx, 2)
	if _, _, ok := pw.RecvTick(); !ok {
		t.Fatal("a vector step must place its own bead; nothing was placed")
	}

	// An arrival that is exactly PERPENDICULAR: nothing steps, so nothing is placed and the
	// bead loop ends here rather than being handed on regardless.
	n.TopTiltThetaIdx = 0
	in <- Wiring.TiltVectorMsg{ThetaIdx: Wiring.PerpendicularThetaIdx}
	n.handleVectorCycle(3)
	for tick := int64(4); tick < 10; tick++ {
		pw.DriveOneCycle(ctx, tick)
		if _, _, ok := pw.RecvTick(); ok {
			t.Fatal("nothing steps on a perpendicular arrival, so no bead may be placed")
		}
	}
}
