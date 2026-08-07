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

func TestPerpendicularHaltsOnItsOwnTwoSeparations(t *testing.T) {
	r := testRing()
	m := perpendicularMachine{}
	top := r.at(0)

	// The arrival on this node's own top (separation 0) and on its bottom (a half turn) are the
	// same relationship seen from the two ends of the pair, since each receives the other's
	// normal. Both are this machine's halt.
	for _, sep := range []int32{0, r.halfTurn} {
		if !m.halted(top, r.at(sep)) {
			t.Errorf("separation %d is perpendicular and must halt this machine", sep)
		}
	}
	// A quarter turn is PARALLEL. This machine must walk over it, not stop on it: stopping at
	// any halt let a disturbed perpendicular pair meet a quarter turn on the way home and take
	// up parallel there, in both directions of disturbance.
	if m.halted(top, r.at(r.quarterTurn)) {
		t.Error("a quarter turn is parallel — the perpendicular machine must step over it, not halt")
	}
}

func TestParallelHaltsOnlyOnAQuarterTurn(t *testing.T) {
	r := testRing()
	m := parallelMachine{}
	top := r.at(0)

	if !m.halted(top, r.at(r.quarterTurn)) {
		t.Error("a quarter turn is parallel and must halt this machine")
	}
	for _, sep := range []int32{0, r.halfTurn} {
		if m.halted(top, r.at(sep)) {
			t.Errorf("separation %d is perpendicular — the parallel machine must step over it", sep)
		}
	}
}

func TestEachMachineStepsTowardItsOwnHalt(t *testing.T) {
	r := testRing()
	// Every starting separation, both machines: one step must never leave the node further from
	// its own halt than it was. This is the property the whole rule rests on, and the acute test
	// broke it at exactly one place — a quarter turn, where it reported "not acute" and the rule
	// read that as stand still.
	for sep := int32(0); sep < r.points; sep++ {
		arrival := r.at(sep)
		top := r.at(0)

		perp := perpendicularMachine{}
		if !perp.halted(top, arrival) {
			if got, was := perp.miss(perp.step(top, arrival), arrival), perp.miss(top, arrival); got >= was {
				t.Errorf("perpendicular at separation %d: step left miss at %d, was %d", sep, got, was)
			}
		}
		par := parallelMachine{}
		if !par.halted(top, arrival) {
			if got, was := par.miss(par.step(top, arrival), arrival), par.miss(top, arrival); got >= was {
				t.Errorf("parallel at separation %d: step left miss at %d, was %d", sep, got, was)
			}
		}
	}
}

func TestPerpendicularStepsThroughTheParallelHalt(t *testing.T) {
	r := testRing()
	m := perpendicularMachine{}
	// Standing one step off a quarter turn, walking home to separation 0, this machine must pass
	// over the quarter turn rather than stop on it. Sitting still here is the freeze the acute
	// test produced; halting here is the capture that produced a parallel pair from a
	// perpendicular one.
	top := r.at(r.quarterTurn + 1)
	arrival := r.at(0)
	if m.halted(top, arrival) {
		t.Fatal("one step off a quarter turn is not the perpendicular halt")
	}
	stepped := m.step(top, arrival)
	if stepped == top {
		t.Fatal("the perpendicular machine stood still one step off a quarter turn")
	}
	if m.halted(stepped, arrival) == false && m.miss(stepped, arrival) >= m.miss(top, arrival) {
		t.Error("the step did not close on the perpendicular halt")
	}
}

func TestSeparationIsTheShortWayRound(t *testing.T) {
	r := testRing()
	// Never more than a half turn, and the same number whichever side the target sits — the
	// halt tests compare it against exact values, so a long-way answer would name the wrong
	// state.
	for a := int32(0); a < r.points; a++ {
		for b := int32(0); b < r.points; b++ {
			sep := r.at(a).separation(r.at(b))
			if sep < 0 || sep > r.halfTurn {
				t.Fatalf("separation(%d,%d) = %d is outside 0..%d", a, b, sep, r.halfTurn)
			}
			if back := r.at(b).separation(r.at(a)); back != sep {
				t.Fatalf("separation(%d,%d) = %d but separation(%d,%d) = %d", a, b, sep, b, a, back)
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

func TestAdoptedMachineSticksUntilCleared(t *testing.T) {
	r := testRing()
	n := &Node{Ring: r, Top: r.at(0)}

	n.adoptMachine(Wiring.TiltMachinePerpendicular)
	if _, ok := n.Machine.(perpendicularMachine); !ok {
		t.Fatalf("adopt did not take: running %v", n.Machine)
	}
	// A second choice — a click landing mid-run, or the partner's own answer arriving — must not
	// switch a running machine. Re-deciding on a jitter click switched a started perpendicular
	// pair to parallel one step after START.
	n.adoptMachine(Wiring.TiltMachineParallel)
	if _, ok := n.Machine.(perpendicularMachine); !ok {
		t.Errorf("a later choice switched a running machine: now %v", n.Machine)
	}
	// RESET is the one thing that releases it.
	n.clear()
	if n.Machine != nil {
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
