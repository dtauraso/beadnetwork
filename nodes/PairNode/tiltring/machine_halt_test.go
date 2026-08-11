package tiltring

// machine_halt_test.go — each machine's own halt/step decision, asked of ONE state at a time.
//
// Every test here calls a decision function directly and checks what it returned. None of them
// starts a goroutine, and none checks that two nodes reached each other: delivery, ordering and
// the absence of races are guaranteed by construction in this network (docs/process/testing-shape.md),
// so a test asserting them would exercise Go's runtime rather than this rule.
//
// The rule these cover is the one the last session's log kept disproving, and each test names
// the failure it would have caught. (Was PairNode/machine_halt_test.go.)

import (
	"testing"
)

func TestPerpendicularHaltsOnItsOwnTwoSeparations(t *testing.T) {
	r := testRing()
	m := perpendicular
	top := r.At(0)

	// The arrival on this node's own top (angle length 0) and on its bottom (a half turn) are the
	// same relationship seen from the two ends of the pair, since each receives the other's
	// normal. Both are this machine's halt.
	for _, sep := range []int32{0, r.HalfTurn} {
		if !m.Settled(top, r.At(sep)) {
			t.Errorf("angle length %d is perpendicular and must halt this machine", sep)
		}
	}
	// A quarter turn is PARALLEL. This machine must walk over it, not stop on it: stopping at
	// any halt let a disturbed perpendicular pair meet a quarter turn on the way home and take
	// up parallel there, in both directions of disturbance.
	if m.Settled(top, r.At(r.QuarterTurn)) {
		t.Error("a quarter turn is parallel — the perpendicular machine must step over it, not halt")
	}
}

func TestParallelHaltsOnlyOnAQuarterTurn(t *testing.T) {
	r := testRing()
	m := parallel
	top := r.At(0)

	if !m.Settled(top, r.At(r.QuarterTurn)) {
		t.Error("a quarter turn is parallel and must halt this machine")
	}
	for _, sep := range []int32{0, r.HalfTurn} {
		if m.Settled(top, r.At(sep)) {
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
	for sep := int32(0); sep < r.Points; sep++ {
		arrival := r.At(sep)
		top := r.At(0)

		perp := perpendicular
		if !perp.Settled(top, arrival) {
			if got, was := offBy(perp, steppedTop(perp, top, arrival), arrival), offBy(perp, top, arrival); got >= was {
				t.Errorf("perpendicular at angle length %d: step left miss at %d, was %d", sep, got, was)
			}
		}
		par := parallel
		if !par.Settled(top, arrival) {
			if got, was := offBy(par, steppedTop(par, top, arrival), arrival), offBy(par, top, arrival); got >= was {
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
	top := r.At(r.QuarterTurn + 1)
	arrival := r.At(0)
	if m.Settled(top, arrival) {
		t.Fatal("one step off a quarter turn is not the perpendicular halt")
	}
	stepped := steppedTop(m, top, arrival)
	if stepped == top {
		t.Fatal("the perpendicular machine stood still one step off a quarter turn")
	}
	if m.Settled(stepped, arrival) == false && offBy(m, stepped, arrival) >= offBy(m, top, arrival) {
		t.Error("the step did not close on the perpendicular halt")
	}
}

// TestTheTwoMissesAreComplements locks the identity the one-machine fold rests on: the modes are
// not merely alike, they are one rule read in two directions. If this ever fails, the stopping counts
// have stopped being midpoints of each other and the two modes are genuinely separate rules
// again — which is the reading under which the split into two files was right (machine.go's
// header, docs/pair-node/rules/audit.html).
func TestTheTwoMissesAreComplements(t *testing.T) {
	r := testRing()
	top := r.At(0)
	for sep := int32(0); sep < r.Points; sep++ {
		arrival := r.At(sep)
		perp := offBy(perpendicular, top, arrival)
		par := offBy(parallel, top, arrival)
		if perp+par != r.QuarterTurn {
			t.Errorf("angle length %d: perpendicular miss %d + parallel miss %d = %d, want the quarter turn %d",
				sep, perp, par, perp+par, r.QuarterTurn)
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
	top := r.At(0)
	for _, m := range []Machine{Setting, perpendicular, parallel} {
		// A row is one count or "anywhere", so the home set is built from the row rather
		// than read off a list — the shape the data actually has.
		row := m.Stopping()
		home := func(c int32) bool { return row.Anywhere || c == row.At(r) }
		for sep := int32(0); sep < r.Points; sep++ {
			// The halt is asked about the count from the NEARER END, which is the number a
			// stopping-count row is written in. It is not the folded angle length: the two
			// agree only where the nearer end is the top.
			c, _ := top.NearerEndCount(r.At(sep))
			if got := m.Settled(top, r.At(sep)); got != home(c) {
				t.Errorf("%v at arrival %d (count %d): halted=%v, its row says %v",
					m, sep, c, got, home(c))
			}
		}
	}
}

func TestSeparationIsTheShortWayRound(t *testing.T) {
	r := testRing()
	// Never more than a half turn, and the same number whichever side the target sits — the
	// halt tests compare it against exact values, so a long-way answer would name the wrong
	// state.
	for a := int32(0); a < r.Points; a++ {
		for b := int32(0); b < r.Points; b++ {
			sep := r.At(a).AngleLength(r.At(b))
			if sep < 0 || sep > r.HalfTurn {
				t.Fatalf("AngleLength(%d,%d) = %d is outside 0..%d", a, b, sep, r.HalfTurn)
			}
			if back := r.At(b).AngleLength(r.At(a)); back != sep {
				t.Fatalf("AngleLength(%d,%d) = %d but AngleLength(%d,%d) = %d", a, b, sep, b, a, back)
			}
		}
	}
}
