package Node1

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
// what turned a pair one way forever. The bead now only paces; stepFromVector's acute tests are
// the only rule that turns a tilt on an arrival, and they are asserted below.

// applyTiltEdit is what the RESET button, the START TILT button, and the tilt-angle panel
// (TiltVectorButtons.tsx / TiltVectorAnglePanel.tsx) each drive; this asserts THIS node
// kind's own decision for a reset, from any starting indices: both indices land on 0 (the
// start position), and — unlike Start — no bead is placed ("the kick" only fires for
// Start, never a stop-and-return).
func TestApplyTiltEditResetReturnsBothIndicesToZero(t *testing.T) {
	for _, start := range []int32{
		0,
		5,
		-9,
		Wiring.PerpendicularThetaIdx,
	} {
		n := &Node{TopTiltThetaIdx: start}
		placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
		if placeBead {
			t.Fatalf("reset from theta=%d must place no bead, got placeBead=true", start)
		}
		if n.TopTiltThetaIdx != 0 {
			t.Fatalf("reset from theta=%d: want index 0, got theta=%d", start, n.TopTiltThetaIdx)
		}
	}
}

// The drawn coplanar normal must sit exactly +6 steps (90°) in θ from the tilt, φ
// unchanged, for Node1 — pure index arithmetic reusing Wiring.PerpendicularThetaIdx, not
// a cross product (Part 1 of this task). These starting indices all sit within the first
// pole-crossing zone (no odd number of Wiring.HalfTurnThetaIdx poles crossed yet), so no
// half-turn flip applies here — the flip itself is asserted separately, below.
func TestCoplanarNormalIsPlusSixStepsInTheta(t *testing.T) {
	for _, theta := range []int32{0, 5, 9} {
		n := &Node{TopTiltThetaIdx: theta}
		norm := n.coplanarNormal()
		if norm.ThetaIdx != theta+Wiring.PerpendicularThetaIdx {
			t.Fatalf("coplanarNormal theta: want tilt+%d=%d, got %d", Wiring.PerpendicularThetaIdx, theta+Wiring.PerpendicularThetaIdx, norm.ThetaIdx)
		}
		// topTilt is the stored index NAMED as a direction — the acute tests read it as one
		// operand, so it must be exactly the index and never a derived value.
		if got := n.topTilt().ThetaIdx; got != theta {
			t.Fatalf("topTilt must be the stored index itself: want %d, got %d", theta, got)
		}
	}
}

// coplanarNormal gains an extra half turn (Wiring.HalfTurnThetaIdx) whenever an ODD number
// of poles (Wiring.HalfTurnThetaIdx-sized buckets, floor-divided) has been crossed by
// TopTiltThetaIdx. This is a PURE function of the index — no stored crossing flag — so it
// must be asserted directly at several bucket boundaries in BOTH directions, explicitly
// including negative indices: Node1's base direction subtracts, so negative indices are the
// common case, and Go's truncating `/` gets exactly this case wrong.
func TestCoplanarNormalFlipsParityAcrossPoleCrossings(t *testing.T) {
	half := Wiring.HalfTurnThetaIdx
	quarter := Wiring.PerpendicularThetaIdx
	cases := []struct {
		theta    int32
		wantFlip bool
	}{
		// Positive side: [0,11] no crossing (poles=0, even); [12,23] one crossing (odd);
		// [24,...] two crossings (even).
		{0, false},
		{half - 1, false},   // 11
		{half, true},        // 12: first pole
		{half + 1, true},    // 13
		{2*half - 1, true},  // 23
		{2 * half, false},   // 24: second pole
		{2*half + 1, false}, // 25
		// Negative side: floor-division buckets [-12,-1] as poles=-1 (odd); [-24,-13] as
		// poles=-2 (even); [-36,-25] as poles=-3 (odd) — this is the asymmetric-looking but
		// correct floor bucketing the task calls for, and the case truncating division gets
		// wrong.
		{-1, true},
		{-half, true},       // -12
		{-half - 1, false},  // -13
		{-2 * half, false},  // -24
		{-2*half - 1, true}, // -25
		{-3 * half, true},   // -36
	}
	for _, c := range cases {
		n := &Node{TopTiltThetaIdx: c.theta}
		norm := n.coplanarNormal()
		want := c.theta + quarter
		if c.wantFlip {
			want += half
		}
		if norm.ThetaIdx != want {
			t.Fatalf("theta=%d (wantFlip=%v): coplanarNormal theta want %d, got %d",
				c.theta, c.wantFlip, want, norm.ThetaIdx)
		}
	}
}

// RESET must also drain any value already sitting on VectorIn (depth-1 latest-wins), so
// it cannot arrive on the NEXT cycle and immediately step the tilt again, undoing the
// reset. Verified by observing the channel is empty afterward.
func TestApplyTiltEditResetDrainsVectorIn(t *testing.T) {
	vectorIn := make(chan Wiring.TiltVectorMsg, 1)
	vectorIn <- Wiring.TiltVectorMsg{ThetaIdx: 99}
	n := &Node{TopTiltThetaIdx: 5, VectorIn: vectorIn}
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
	n := &Node{TopTiltThetaIdx: 3, VectorOut: out}
	placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Up: true})
	if placeBead {
		t.Fatalf("a plain adjust must place NO bead, got placeBead=true")
	}
	if n.TopTiltThetaIdx != 4 {
		t.Fatalf("adjust theta up: want theta=4, got theta=%d", n.TopTiltThetaIdx)
	}
	select {
	case v := <-out:
		t.Fatalf("a plain adjust must send NOTHING on VectorOut; got %+v", v)
	default:
	}
}

// outgoingVector reverses the coplanar normal by 180° in θ — Node1's own direction is
// −12 index steps (−180°, two of the 6-step quarter turns). There is no φ any more.
// This is THIS node's own arithmetic, asserted without any channel involved.
func TestOutgoingVectorIsMinus180StepsInThetaOnly(t *testing.T) {
	// NEGATIVE indices are included deliberately: Node1's base direction SUBTRACTS, so a
	// running node spends most of its life below zero, and every one of these derivations
	// is plain integer arithmetic that must not care about the sign.
	for _, theta := range []int32{3, 0, -1, -7, -Wiring.HalfTurnThetaIdx} {
		n := &Node{TopTiltThetaIdx: theta}
		norm := n.coplanarNormal()
		out := n.outgoingVector()
		if want := norm.ThetaIdx - 2*Wiring.PerpendicularThetaIdx; out.ThetaIdx != want {
			t.Fatalf("theta=%d: outgoingVector want %d (norm - 12 steps), got %d", theta, want, out.ThetaIdx)
		}
		// The outgoing vector is the coplanar normal's exact antipode, so it must be a half
		// turn from it however the pole flip landed on the normal itself.
		if diff := norm.ThetaIdx - out.ThetaIdx; diff != Wiring.HalfTurnThetaIdx {
			t.Fatalf("theta=%d: outgoing must sit a half turn (%d) from the normal, got %d", theta, Wiring.HalfTurnThetaIdx, diff)
		}
		// And the bottom tilt is the TOP's exact antipode, which is what makes the two acute
		// tests exact opposites of each other — the property the whole step rule rests on.
		if bottom := n.bottomTilt(); bottom.ThetaIdx-n.topTilt().ThetaIdx != Wiring.HalfTurnThetaIdx {
			t.Fatalf("theta=%d: bottom must sit a half turn from the top, got %d", theta, bottom.ThetaIdx-n.topTilt().ThetaIdx)
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
	n := &Node{TopTiltThetaIdx: 0}
	if moved := n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: 0}); !moved || n.TopTiltThetaIdx != -1 {
		t.Fatalf("acute with the TOP tilt: want moved=true thetaIdx=-1, got moved=%v thetaIdx=%d",
			moved, n.TopTiltThetaIdx)
	}
}

func TestStepFromVectorReversesWhenAcuteWithBottom(t *testing.T) {
	// A half turn from the tilt: now it is 0 steps from the BOTTOM (acute) and a half turn
	// from the top (not acute),
	// so the step must go the OTHER way from the case above.
	n := &Node{TopTiltThetaIdx: 0}
	if moved := n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: Wiring.HalfTurnThetaIdx}); !moved || n.TopTiltThetaIdx != 1 {
		t.Fatalf("acute with the BOTTOM tilt: want moved=true thetaIdx=1, got moved=%v thetaIdx=%d",
			moved, n.TopTiltThetaIdx)
	}
}

// Exactly perpendicular to both the top and bottom tilt is the HALT case: stepFromVector
// steps nothing and reports moved=false — this is how the vector exchange comes to rest.
func TestStepFromVectorHaltsWhenNeitherDotIsAcute(t *testing.T) {
	n := &Node{TopTiltThetaIdx: Wiring.PerpendicularThetaIdx}
	before := n.TopTiltThetaIdx
	perp := Wiring.TiltVectorMsg{ThetaIdx: n.TopTiltThetaIdx + Wiring.PerpendicularThetaIdx}
	if moved := n.stepFromVector(perp); moved {
		t.Fatal("stepFromVector must report moved=false on a perpendicular arrival, got true")
	}
	if n.TopTiltThetaIdx != before {
		t.Fatalf("a perpendicular arrival must step NOTHING; got %d, want unchanged %d",
			n.TopTiltThetaIdx, before)
	}
}

// stepFromVector's three cases: acute with top steps -1 and returns true, acute with
// bottom steps +1 and returns true, and exactly perpendicular to both steps nothing and
// returns false. Consolidates the three single-case tests above into one table asserting the
// full gate.
func TestStepFromVectorGatesOnBothDotsForAllThreeCases(t *testing.T) {
	cases := []struct {
		name        string
		arrivedIdx  int32
		wantMoved   bool
		wantDeltaTh int32
	}{
		{"acute with top", 0, true, -1},
		{"acute with bottom", Wiring.HalfTurnThetaIdx, true, 1},
		{"exactly perpendicular", Wiring.PerpendicularThetaIdx, false, 0},
	}
	for _, c := range cases {
		n := &Node{TopTiltThetaIdx: 0}
		moved := n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: c.arrivedIdx})
		if moved != c.wantMoved {
			t.Fatalf("%s: want moved=%v, got %v", c.name, c.wantMoved, moved)
		}
		if n.TopTiltThetaIdx != c.wantDeltaTh {
			t.Fatalf("%s: want thetaIdx=%d, got %d", c.name, c.wantDeltaTh, n.TopTiltThetaIdx)
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
	// Node1 sends on oneToTwo and receives on twoToOne; its partner is the mirror image.
	one := &Node{TopTiltThetaIdx: 4, VectorOut: oneToTwo, VectorIn: twoToOne}
	partnerIn := oneToTwo // what the other node owns the receive end of

	// A stale direction is in flight BOTH ways when reset is pressed.
	oneToTwo <- Wiring.TiltVectorMsg{ThetaIdx: 9}
	twoToOne <- Wiring.TiltVectorMsg{ThetaIdx: 9}

	one.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
	if one.TopTiltThetaIdx != 0 {
		t.Fatalf("reset must zero the index; got theta=%d", one.TopTiltThetaIdx)
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
	n := &Node{TopTiltThetaIdx: 5, VectorOut: out, VectorIn: in}

	in <- Wiring.TiltVectorMsg{Reset: true}
	n.handleVectorCycle(0)

	if n.TopTiltThetaIdx != 0 {
		t.Fatalf("a received reset must zero the index; got theta=%d", n.TopTiltThetaIdx)
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
	n := &Node{TopTiltThetaIdx: 2, VectorIn: in}

	in <- Wiring.TiltVectorMsg{ThetaIdx: 7}
	n.handleVectorCycle(0)
	in <- Wiring.TiltVectorMsg{ThetaIdx: -2}
	n.handleVectorCycle(0)

	if n.ReceivedThetaIdx != -2 {
		t.Fatalf("want the LATEST arrival theta=-2, got theta=%d", n.ReceivedThetaIdx)
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
	n := &Node{TopTiltThetaIdx: 0, ReceivedThetaIdx: 4, ReceivedSet: true}
	in := make(chan Wiring.TiltVectorMsg, 1)
	out := make(chan Wiring.TiltVectorMsg, 1)
	n.VectorIn, n.VectorOut = in, out
	arrived := Wiring.TiltVectorMsg{ThetaIdx: Wiring.PerpendicularThetaIdx}
	in <- arrived

	n.handleVectorCycle(0)

	if !n.ReceivedSet || n.ReceivedThetaIdx != arrived.ThetaIdx {
		t.Fatalf("the arrived direction must be recorded even though it halts; got set=%v theta=%d",
			n.ReceivedSet, n.ReceivedThetaIdx)
	}
	if n.TopTiltThetaIdx != 0 {
		t.Fatalf("a perpendicular arrival must step NOTHING; got %d, want unchanged 0", n.TopTiltThetaIdx)
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
	n := &Node{TopTiltThetaIdx: 5, ReceivedThetaIdx: 9, ReceivedSet: true}
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
	n := &Node{TopTiltThetaIdx: 5, VectorOut: out, VectorIn: in,
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
	n := &Node{TopTiltThetaIdx: 4}
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
	wantBottomTheta := int32(4) + Wiring.HalfTurnThetaIdx
	if gotBottomTheta != wantBottomTheta {
		t.Fatalf("opening bottom tilt: want %d, got %d", wantBottomTheta, gotBottomTheta)
	}
	wantNormalTheta := int32(4) + Wiring.PerpendicularThetaIdx
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
func TestStartOpensTheVectorExchangeWithoutChangingAnyIndex(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{TopTiltThetaIdx: 3, VectorOut: out}

	placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Start: true})

	if !placeBead {
		t.Fatal("Start must place a bead, got placeBead=false")
	}
	if n.TopTiltThetaIdx != 3 {
		t.Fatalf("Start must change NO index; got theta=%d, want unchanged theta=3", n.TopTiltThetaIdx)
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
	n := &Node{TopTiltThetaIdx: 3, VectorOut: out}

	n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})

	got := <-out
	if !got.Reset {
		t.Fatalf("reset must send the Reset marker, got a direction %+v", got)
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
	n := &Node{TopTiltThetaIdx: 0, VectorIn: in, Out: out}

	// An arrival that LEANS: the acute tests move this node, so a bead goes out with the reply.
	in <- Wiring.TiltVectorMsg{ThetaIdx: 0}
	n.handleVectorCycle(1)
	pw.DriveOneCycle(ctx, 2)
	if _, _, ok := pw.RecvTick(); !ok {
		t.Fatal("a vector step must place its own bead; nothing was placed")
	}

	// An arrival that is exactly PERPENDICULAR: this halts (steps nothing) and must place NO
	// bead across several subsequent cycles.
	n.TopTiltThetaIdx = 0
	in <- Wiring.TiltVectorMsg{ThetaIdx: Wiring.PerpendicularThetaIdx}
	n.handleVectorCycle(3)
	for tick := int64(4); tick < 10; tick++ {
		pw.DriveOneCycle(ctx, tick)
		if _, _, ok := pw.RecvTick(); ok {
			t.Fatal("a perpendicular arrival must halt and place no bead, but one was placed")
		}
	}
}
