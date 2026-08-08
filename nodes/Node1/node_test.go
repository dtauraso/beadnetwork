package Node1

// node_test.go — the pair rule, asked of ONE node at a time.
//
// Every test here calls a decision function directly and checks what it returned. None of them
// starts a goroutine, and none checks that two nodes reached each other: delivery, ordering and
// the absence of races are guaranteed by construction in this network (docs/testing-shape.md),
// so a test asserting them would exercise Go's runtime rather than this rule.
//
// The rule these cover is the one the last session's log kept disproving, and each test names
// the failure it would have caught.

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// a 48-point ring: a quarter turn is 12, a half turn 24 — the lattice the live scene runs at,
// so the numbers here read the same as the rows in the probe log.
func testRing() *ring { return newRing(48) }

// The two chosen modes, named once here so the tests read as the modes rather than as calls.
// `setting` is the zero value and lives in machine.go, since production code names it too.
var (
	perpendicular = machineFor(Wiring.TiltMachinePerpendicular)
	parallel      = machineFor(Wiring.TiltMachineParallel)
)

func TestPerpendicularHaltsOnItsOwnTwoSeparations(t *testing.T) {
	r := testRing()
	m := perpendicular
	top := r.at(0)

	// The arrival on this node's own top (angle length 0) and on its bottom (a half turn) are the
	// same relationship seen from the two ends of the pair, since each receives the other's
	// normal. Both are this machine's halt.
	for _, sep := range []int32{0, r.halfTurn} {
		if !m.settled(top, r.at(sep)) {
			t.Errorf("angle length %d is perpendicular and must halt this machine", sep)
		}
	}
	// A quarter turn is PARALLEL. This machine must walk over it, not stop on it: stopping at
	// any halt let a disturbed perpendicular pair meet a quarter turn on the way home and take
	// up parallel there, in both directions of disturbance.
	if m.settled(top, r.at(r.quarterTurn)) {
		t.Error("a quarter turn is parallel — the perpendicular machine must step over it, not halt")
	}
}

func TestParallelHaltsOnlyOnAQuarterTurn(t *testing.T) {
	r := testRing()
	m := parallel
	top := r.at(0)

	if !m.settled(top, r.at(r.quarterTurn)) {
		t.Error("a quarter turn is parallel and must halt this machine")
	}
	for _, sep := range []int32{0, r.halfTurn} {
		if m.settled(top, r.at(sep)) {
			t.Errorf("angle length %d is perpendicular — the parallel machine must step over it", sep)
		}
	}
}

func TestEachMachineStepsTowardItsOwnHalt(t *testing.T) {
	r := testRing()
	// Every starting angle length, both machines: one step must never leave the node further from
	// its own halt than it was. This is the property the whole rule rests on, and the acute test
	// broke it at exactly one place — a quarter turn, where it reported "not acute" and the rule
	// read that as stand still.
	for sep := int32(0); sep < r.points; sep++ {
		arrival := r.at(sep)
		top := r.at(0)

		perp := perpendicular
		if !perp.settled(top, arrival) {
			if got, was := perp.fromRest(perp.step(top, arrival), arrival), perp.fromRest(top, arrival); got >= was {
				t.Errorf("perpendicular at angle length %d: step left miss at %d, was %d", sep, got, was)
			}
		}
		par := parallel
		if !par.settled(top, arrival) {
			if got, was := par.fromRest(par.step(top, arrival), arrival), par.fromRest(top, arrival); got >= was {
				t.Errorf("parallel at angle length %d: step left miss at %d, was %d", sep, got, was)
			}
		}
	}
}

func TestPerpendicularStepsThroughTheParallelHalt(t *testing.T) {
	r := testRing()
	m := perpendicular
	// Standing one step off a quarter turn, walking home to angle length 0, this machine must pass
	// over the quarter turn rather than stop on it. Sitting still here is the freeze the acute
	// test produced; halting here is the capture that produced a parallel pair from a
	// perpendicular one.
	top := r.at(r.quarterTurn + 1)
	arrival := r.at(0)
	if m.settled(top, arrival) {
		t.Fatal("one step off a quarter turn is not the perpendicular halt")
	}
	stepped := m.step(top, arrival)
	if stepped == top {
		t.Fatal("the perpendicular machine stood still one step off a quarter turn")
	}
	if m.settled(stepped, arrival) == false && m.fromRest(stepped, arrival) >= m.fromRest(top, arrival) {
		t.Error("the step did not close on the perpendicular halt")
	}
}

// TestRestingLengthsFollowFromTheCongruences checks that the resting lengths are DERIVED rather
// than chosen. A tilt is a LINE, so reversing it changes nothing, which makes the tilts live in
// Z24 modulo the half turn — and in those terms the two arrangements are congruences on the gap
// g between the two tilts:
//
//	parallel       g = 0 (mod 12)     the two lines coincide
//	perpendicular  g = 6 (mod 12)     a quarter turn between them
//
// What ARRIVES is the partner's normal, a = p + 6, so the angle length this node measures is
// |g - 6| folded — which turns the two congruences into L = 6 and L in {0, 12}. That is where
// restingLengths comes from, and this sweeps every (partner, tilt) pair to confirm it, including
// that no other gap produces a resting length by accident.
func TestRestingLengthsFollowFromTheCongruences(t *testing.T) {
	r := newRing(24)
	for p := int32(0); p < 24; p++ {
		a := r.at((p + 6) % 24) // the partner sends its normal, not its tilt
		for tilt := int32(0); tilt < 24; tilt++ {
			g := ((tilt-p)%24 + 24) % 24
			L := r.at(tilt).angleLength(a)
			switch {
			case g%12 == 0 && L != 6:
				t.Fatalf("p=%d tilt=%d g=%d: parallel congruence but L=%d, want 6", p, tilt, g, L)
			case g%12 == 6 && L != 0 && L != 12:
				t.Fatalf("p=%d tilt=%d g=%d: perpendicular congruence but L=%d, want 0 or 12", p, tilt, g, L)
			case g%12 != 0 && g%12 != 6 && (L == 0 || L == 6 || L == 12):
				t.Fatalf("p=%d tilt=%d g=%d: neither congruence, yet L=%d is a resting length", p, tilt, g, L)
			}
		}
	}
}

// TestTheTwoMissesAreComplements locks the identity the one-machine fold rests on: the modes are
// not merely alike, they are one rule read in two directions. If this ever fails, the home sets
// have stopped being midpoints of each other and the two modes are genuinely separate rules
// again — which is the reading under which the split into two files was right (machine.go's
// header, docs/pair-node/audit.html).
func TestTheTwoMissesAreComplements(t *testing.T) {
	r := testRing()
	top := r.at(0)
	for sep := int32(0); sep < r.points; sep++ {
		arrival := r.at(sep)
		perp := perpendicular.fromRest(top, arrival)
		par := parallel.fromRest(top, arrival)
		if perp+par != r.quarterTurn {
			t.Errorf("angle length %d: perpendicular miss %d + parallel miss %d = %d, want the quarter turn %d",
				sep, perp, par, perp+par, r.quarterTurn)
		}
	}
}

// TestAModeHaltsExactlyOnItsHomeSet checks that the shared rule agrees with each mode's declared
// data: halted is true at exactly the declared homes and nowhere else, for every mode in the
// table. It reads the home set rather than naming angle lengths, so a mode added to the table is
// covered here without editing this test.
//
// It cannot catch a home set that is simply WRONG — both sides of the comparison move together —
// and that is not its job: what it catches is miss, the fold, or halted drifting away from the
// data, which is the failure the one-machine fold could introduce.
func TestAModeHaltsExactlyOnItsHomeSet(t *testing.T) {
	r := testRing()
	top := r.at(0)
	for _, m := range []tiltMachine{setting, perpendicular, parallel} {
		home := map[int32]bool{}
		for _, h := range m.resting(r) {
			home[h] = true
		}
		for sep := int32(0); sep < r.points; sep++ {
			// angle length folds the long way round into the short one, so the halt is asked
			// about the folded reading — which is the number a home set is written in.
			folded := top.angleLength(r.at(sep))
			if got := m.settled(top, r.at(sep)); got != home[folded] {
				t.Errorf("%v at angle length %d (folds to %d): halted=%v, home set says %v",
					m, sep, folded, got, home[folded])
			}
		}
	}
}

func TestSeparationIsTheShortWayRound(t *testing.T) {
	r := testRing()
	// Never more than a half turn, and the same number whichever side the target sits — the
	// halt tests compare it against exact values, so a long-way answer would name the wrong
	// state.
	for a := int32(0); a < r.points; a++ {
		for b := int32(0); b < r.points; b++ {
			sep := r.at(a).angleLength(r.at(b))
			if sep < 0 || sep > r.halfTurn {
				t.Fatalf("angleLength(%d,%d) = %d is outside 0..%d", a, b, sep, r.halfTurn)
			}
			if back := r.at(b).angleLength(r.at(a)); back != sep {
				t.Fatalf("angleLength(%d,%d) = %d but angleLength(%d,%d) = %d", a, b, sep, b, a, back)
			}
		}
	}
}

func TestMachineIsReadFromTheGapNotFromOneTilt(t *testing.T) {
	r := testRing()
	// What arrives is the partner's NORMAL, a quarter turn off its tilt. So for a chosen pair of
	// tilts, the arrival this node sees is partnerTilt + a quarter.
	arrivalFor := func(partnerTilt int32) *tiltState { return r.at(partnerTilt).quarter }

	cases := []struct {
		name             string
		ownTilt, partner int32
		want             Wiring.TiltMachine
	}{
		// The case the live test runs: one node clicked to a quarter turn, the other left at 0.
		{"a quarter turn apart", 12, 0, Wiring.TiltMachinePerpendicular},
		{"a quarter turn apart, the other way", 0, 12, Wiring.TiltMachinePerpendicular},
		// BOTH tilted, still a quarter apart — the gap is the pair's, not one node's angle
		// against zero, which is what reading a single tilt got wrong.
		{"both tilted, a quarter apart", 20, 8, Wiring.TiltMachinePerpendicular},
		{"the same direction", 7, 7, Wiring.TiltMachineParallel},
		// One step short of a quarter turn — the gap a click reads mid-setup. Deciding here is
		// what locked a pair to the wrong machine while the tilt was still on its way.
		{"one step off a quarter turn", 11, 0, Wiring.TiltMachineParallel},
		{"an ordinary acute gap", 3, 0, Wiring.TiltMachineParallel},
	}
	for _, c := range cases {
		n := &Node{Ring: r, Top: r.at(c.ownTilt)}
		if got := n.machineForGap(arrivalFor(c.partner)); got != c.want {
			t.Errorf("%s (tilts %d and %d): chose %v, want %v", c.name, c.ownTilt, c.partner, got, c.want)
		}
	}
}

// TestASettingNodeHoldsWhereverItStands is the behaviour that used to be an `if Machine == nil`
// exemption in stepFromVector and is now a consequence of the setting mode's home set: a node
// still deciding which machine it runs is already halted at every angle length, so no arrival can
// move it. A zero-value Node is in that mode without being put there.
func TestASettingNodeHoldsWhereverItStands(t *testing.T) {
	r := testRing()
	for sep := int32(0); sep < r.points; sep++ {
		n := &Node{Ring: r, Top: r.at(5)}
		if n.Machine != setting {
			t.Fatalf("a fresh node is not in the setting mode: %v", n.Machine)
		}
		n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: sep})
		if n.Top != r.at(5) {
			t.Errorf("arrival at angle length %d moved a node that is still being set up: top %d, want 5",
				sep, n.Top.idx)
		}
	}
}

// TestASettingNodeTellsTheOtherEndNothing: the mode's own choice is TiltMachineNone, so the wire
// says "no choice carried" without outgoingVector testing for one.
func TestASettingNodeTellsTheOtherEndNothing(t *testing.T) {
	r := testRing()
	n := &Node{Ring: r, Top: r.at(0)}
	if got := n.outgoingVector().Machine; got != Wiring.TiltMachineNone {
		t.Errorf("a node still being set up announced machine %v, want none", got)
	}
	n.adoptMachine(Wiring.TiltMachineParallel)
	if got := n.outgoingVector().Machine; got != Wiring.TiltMachineParallel {
		t.Errorf("after adopting, announced %v, want parallel", got)
	}
}

func TestAdoptedMachineSticksUntilCleared(t *testing.T) {
	r := testRing()
	n := &Node{Ring: r, Top: r.at(0)}

	n.adoptMachine(Wiring.TiltMachinePerpendicular)
	if n.Machine != perpendicular {
		t.Fatalf("adopt did not take: running %v", n.Machine)
	}
	// A second choice — a click landing mid-run, or the partner's own answer arriving — must not
	// switch a running machine. Re-deciding on a jitter click switched a started perpendicular
	// pair to parallel one step after START.
	n.adoptMachine(Wiring.TiltMachineParallel)
	if n.Machine != perpendicular {
		t.Errorf("a later choice switched a running machine: now %v", n.Machine)
	}
	// RESET is the one thing that releases it.
	n.clear()
	if n.Machine != setting {
		t.Errorf("reset left a machine running: %v", n.Machine)
	}
}

func TestNoMachineMeansNoMovement(t *testing.T) {
	r := testRing()
	n := &Node{Ring: r, Top: r.at(5)}
	// Before any start, and after a reset, an arrival moves nothing. The node used to infer a
	// machine from the arrival here, and that inference always answered perpendicular, because
	// closing on the arrival IS the perpendicular measure.
	before := n.topState()
	for _, sep := range []int32{0, 1, 7, 12, 24, 40} {
		n.stepFromVector(Wiring.TiltVectorMsg{ThetaIdx: sep, Points: r.points})
		if got := n.topState(); got != before {
			t.Fatalf("arrival at %d moved the tilt with no machine running: %d -> %d",
				sep, before.idx, got.idx)
		}
	}
}

// openingOutcome is what one opening of the exchange came to: which machine the pair took up,
// how many rounds it took, and how far apart the two TILTS ended.
type openingOutcome struct {
	machine Wiring.TiltMachine
	rounds  int
	tiltGap int32
	settled bool
}

// runOpening plays one exchange from a starting pair of tilts, calling the same decision
// functions the live node calls, in the order its own loop calls them.
//
// It runs no goroutines and opens no channels: it is the RULE being iterated, not the network
// being exercised. The order is the real one — node 1 opens (START belongs to id 1), node 2
// answers, and from then on each end reads the other's last normal and steps once.
func runOpening(r *ring, tiltA, tiltB int32) openingOutcome {
	a := &Node{Ring: r, Top: r.at(tiltA)}
	b := &Node{Ring: r, Top: r.at(tiltB)}

	// One end reads an arrival exactly as handleVectorCycle does: adopt what the sender says it
	// is running, then choose from the gap if still running nothing, then step.
	read := func(n *Node, arrival *tiltState, senderRuns Wiring.TiltMachine) bool {
		n.adoptMachine(senderRuns)
		if n.Machine == setting {
			n.adoptMachine(n.machineForGap(arrival))
		}
		before := n.topState()
		if !n.Machine.settled(before, arrival) {
			n.Top = n.Machine.step(before, arrival)
		}
		return n.topState() != before
	}
	runs := func(n *Node) Wiring.TiltMachine { return n.Machine.choice() }
	normal := func(n *Node) *tiltState { return n.topState().quarter }

	out := openingOutcome{}
	// A full turn of rounds is far more than any settling walk needs — the longest is a quarter
	// turn's worth of steps — so reaching the cap means it never settled.
	for out.rounds = 1; out.rounds <= int(r.points); out.rounds++ {
		movedB := read(b, normal(a), runs(a))
		movedA := read(a, normal(b), runs(b))
		if !movedA && !movedB {
			out.settled = true
			break
		}
	}
	out.machine = runs(b)
	out.tiltGap = a.topState().angleLength(b.topState())
	return out
}

func TestEveryOpeningSettlesOnTheMachineItChose(t *testing.T) {
	r := newRing(24)
	// The counts this sweep produces are what docs/pair-node/math.html reports, so run it with
	// -v after changing the rule and put the new numbers on the page rather than leaving the
	// old ones there.
	var perp, par, worst int
	var same, reversed int
	for tiltA := int32(0); tiltA < r.points; tiltA++ {
		for tiltB := int32(0); tiltB < r.points; tiltB++ {
			got := runOpening(r, tiltA, tiltB)
			switch {
			case got.machine == Wiring.TiltMachinePerpendicular:
				perp++
			case got.tiltGap == 0:
				par, same = par+1, same+1
			default:
				par, reversed = par+1, reversed+1
			}
			if got.rounds > worst {
				worst = got.rounds
			}
			if !got.settled {
				t.Errorf("opening (%d,%d) never settled: still moving after %d rounds",
					tiltA, tiltB, got.rounds)
				continue
			}
			// Settling means the two TILTS ended in the relationship the chosen machine is for.
			// PARALLEL ACCEPTS TWO OF THEM: the same direction, and a half turn apart, which is
			// the same LINE with the arrows reversed. The halt cannot tell those apart, because
			// angle length folds the long way round into the short one and a reversed partner
			// lands the same distance off. Perpendicular has the one, a quarter turn.
			ok := got.tiltGap == r.quarterTurn
			if got.machine == Wiring.TiltMachineParallel {
				ok = got.tiltGap == 0 || got.tiltGap == r.halfTurn
			}
			if !ok {
				t.Errorf("opening (%d,%d) chose %v but settled with the tilts %d apart",
					tiltA, tiltB, got.machine, got.tiltGap)
			}
		}
	}
	t.Logf("openings=%d perpendicular=%d parallel=%d (same direction=%d, reversed=%d) worst rounds=%d",
		r.points*r.points, perp, par, same, reversed, worst)
}

func TestARingMustHaveAWholeQuarterTurn(t *testing.T) {
	// A quarter turn has to be a whole number of states or the coplanar normal and the
	// perpendicular halt name nothing on the ring.
	defer func() {
		if recover() == nil {
			t.Error("a 10-point lattice has no quarter turn and must not build")
		}
	}()
	newRing(10)
}
