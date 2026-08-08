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

// TestOneRoundIsSignAndRemainder is the other half of what docs/pair-node/arith.html rests on: ONE
// update round, written without a case in it.
//
// The trick is to stop folding. angleLength throws the sign away, which is why the rule that uses
// it needs a comparison to get direction back. Keep the RAW signed difference d = t - a instead —
// unreduced, not even brought onto the ring — and both arrangements are the same statement, since
// the rests sit every 12 apart in d and the arrangement is a SHIFT of 6 inside the modulus:
//
//	parallel       e = (d mod 12) - 6          rests at d = -6, +6
//	perpendicular  e = ((d + 6) mod 12) - 6    rests at d = -12, 0, +12
//
// e is then the signed distance to the nearest rest, and the whole round is t_after = t - sign(e),
// with |e| equal to fromRest. No branch, no minimum, no list.
//
// Two reductions that look necessary and are not, both checked below:
//
//	the fold on d   24 is a whole number of 12s, so every representative of t - a gives the
//	                same e. Reducing d first is work the modulus immediately undoes.
//	abs on d        |d|, folded, IS the angle length — so the folded magnitude the other form
//	                starts from is this same gap with its sign dropped, and dropping the sign
//	                is exactly what costs it the direction.
//
// The ties — a tilt equidistant from two rests, which is where a comparison rule has to be told
// what to prefer — need NO special case here: they land on e = -6 by themselves, and -6 turns up,
// which is exactly what step does. 96 of the 1152 pairs are such ties, so this is not a corner
// being skirted.
func TestOneRoundIsSignAndRemainder(t *testing.T) {
	const points = 24
	r := newRing(points)
	sign := func(x int32) int32 {
		switch {
		case x > 0:
			return 1
		case x < 0:
			return -1
		}
		return 0
	}
	for _, m := range []tiltMachine{
		{mode: Wiring.TiltMachineParallel},
		{mode: Wiring.TiltMachinePerpendicular},
	} {
		shift := int32(0)
		if m.mode == Wiring.TiltMachinePerpendicular {
			shift = 6
		}
		for arr := int32(0); arr < points; arr++ {
			a := r.at(arr)
			for tilt := int32(0); tilt < points; tilt++ {
				cur := r.at(tilt)

				// The RAW difference — no folding, no reduction. e takes it modulo 12
				// below, and 24 is a whole number of 12s, so every representative of
				// t - a on the ring gives the same e. Folding first would be work that
				// the modulus immediately undoes.
				d := tilt - arr
				e := ((d+shift)%12+12)%12 - 6

				// and folding it IS the angle length: |d folded| = l, which is the number
				// the magnitude form starts from. Same gap, with its sign dropped.
				folded := ((tilt-arr)%points + points) % points
				if folded > points/2 {
					folded -= points
				}
				if got := cur.angleLength(a); got != abs32(folded) {
					t.Fatalf("a=%d t=%d: |d|=%d but angleLength=%d", arr, tilt, abs32(folded), got)
				}

				if got := m.fromRest(cur, a); got != abs32(e) {
					t.Fatalf("mode=%v a=%d t=%d d=%d: |e|=%d but fromRest=%d",
						m.mode, arr, tilt, d, abs32(e), got)
				}
				if (e == 0) != m.settled(cur, a) {
					t.Fatalf("mode=%v a=%d t=%d d=%d: e=%d but settled=%v",
						m.mode, arr, tilt, d, e, m.settled(cur, a))
				}
				if e == 0 {
					continue
				}
				want := ((tilt-sign(e))%points + points) % points
				if got := m.step(cur, a).idx; got != want {
					t.Fatalf("mode=%v a=%d t=%d d=%d e=%d: step gave %d, t-sign(e)=%d",
						m.mode, arr, tilt, d, e, got, want)
				}
			}
		}
	}
}

// TestTheWalkIsClosedForm is what docs/pair-node/arith.html rests on: that for a HELD arrival the
// whole walk can be written down rather than run. Three claims, swept over both modes and every
// (arrival, tilt) pair:
//
//	how many arrivals it takes  =  f          fromRest is a count, not just a comparison
//	the direction               =  s          decided by the first arrival and never revisited
//	where it stops              =  t + s*f    so no intermediate state is ever needed
//
// The direction is read the way step reads it — whichever neighbour is closer, ties up — and then
// held FIXED for the rest of the walk here. That is the real content: it is not obvious that the
// sign cannot flip halfway, and at the fold points (l = 0 and l = 12) the two neighbours score the
// same, which is exactly where a walk could turn back. It does not.
//
// This does NOT say the pair is closed form. A live exchange moves both ends, so a is not held;
// arith.html says so on the page rather than leaving the reader to assume otherwise.
func TestTheWalkIsClosedForm(t *testing.T) {
	const points = 24
	r := newRing(points)
	for _, m := range []tiltMachine{
		{mode: Wiring.TiltMachineParallel},
		{mode: Wiring.TiltMachinePerpendicular},
	} {
		for arr := int32(0); arr < points; arr++ {
			a := r.at(arr)
			for tilt := int32(0); tilt < points; tilt++ {
				cur := r.at(tilt)
				f := m.fromRest(cur, a)

				s := int32(-1) // step's own rule: up unless down is strictly closer
				if m.fromRest(cur.next, a) <= m.fromRest(cur.prev, a) {
					s = 1
				}

				steps := int32(0)
				for !m.settled(cur, a) {
					cur = m.step(cur, a)
					steps++
					if steps > 2*points {
						t.Fatalf("mode=%v arrival=%d tilt=%d: never settled", m.mode, arr, tilt)
					}
				}
				if steps != f {
					t.Fatalf("mode=%v arrival=%d tilt=%d: settled after %d arrivals, f said %d",
						m.mode, arr, tilt, steps, f)
				}
				if want := ((tilt+s*f)%points + points) % points; cur.idx != want {
					t.Fatalf("mode=%v arrival=%d tilt=%d: stopped at %d, t+s*f said %d (s=%d f=%d)",
						m.mode, arr, tilt, cur.idx, want, s, f)
				}
			}
		}
	}
}

// TestFromRestIsTheQuarterOffset checks the closed form the update-rules page states, against the
// fromRest the node actually runs. fromRest is a minimum over the mode's resting-length list, which
// is the general shape and the one a new mode joins. But BOTH real modes rest symmetrically about
// the quarter — parallel at the quarter itself, perpendicular at the two ends of the angle-length
// range — so for both of them fromRest is a function of how far the angle length sits off the
// quarter, and no minimum over a list is needed to say it:
//
//	q = |L - quarter|
//	parallel       fromRest = q
//	perpendicular  fromRest = quarter - q
//
// The two being complements is the audit's "one is the other upside down" (docs/pair-node/audit.html),
// here as arithmetic rather than as a picture. This sweeps both lattices the model runs on, every
// tilt against every arrival, so the page cannot claim a shortcut the code does not honour.
func TestFromRestIsTheQuarterOffset(t *testing.T) {
	for _, points := range []int32{24, 48} {
		r := newRing(points)
		perp := tiltMachine{mode: Wiring.TiltMachinePerpendicular}
		par := tiltMachine{mode: Wiring.TiltMachineParallel}
		for tilt := int32(0); tilt < points; tilt++ {
			for arr := int32(0); arr < points; arr++ {
				from, a := r.at(tilt), r.at(arr)
				q := abs32(from.angleLength(a) - r.quarterTurn)
				if got := par.fromRest(from, a); got != q {
					t.Fatalf("points=%d tilt=%d arrival=%d: parallel fromRest=%d, want q=%d",
						points, tilt, arr, got, q)
				}
				if got := perp.fromRest(from, a); got != r.quarterTurn-q {
					t.Fatalf("points=%d tilt=%d arrival=%d: perpendicular fromRest=%d, want quarter-q=%d",
						points, tilt, arr, got, r.quarterTurn-q)
				}
			}
		}
	}
}

// TestRestingLengthsFollowFromTheGaps checks that the resting lengths are DERIVED rather than
// chosen. A tilt is a LINE, so reversing it changes nothing and tilt and tilt+12 are the same
// tilt — which leaves exactly four gaps g between the two tilts that the pair can hold:
//
//	parallel       g = 0 or 12     the two lines coincide
//	perpendicular  g = 6 or 18     a quarter turn between them
//
// What ARRIVES is the partner's normal, a = p + 6, already a quarter turn on, so the angle
// length this node measures is the gap with that quarter turn taken off — which turns those
// four gaps into L = 6 and L in {0, 12}. That is where restingLengths comes from, and this
// sweeps every (partner, tilt) pair to confirm it, including that no other gap produces a
// resting length by accident.
func TestRestingLengthsFollowFromTheGaps(t *testing.T) {
	r := newRing(24)
	for p := int32(0); p < 24; p++ {
		a := r.at((p + 6) % 24) // the partner sends its normal, not its tilt
		for tilt := int32(0); tilt < 24; tilt++ {
			g := ((tilt-p)%24 + 24) % 24
			L := r.at(tilt).angleLength(a)
			parallel := g == 0 || g == 12
			perpendicular := g == 6 || g == 18
			switch {
			case parallel && L != 6:
				t.Fatalf("p=%d tilt=%d g=%d: a parallel gap but L=%d, want 6", p, tilt, g, L)
			case perpendicular && L != 0 && L != 12:
				t.Fatalf("p=%d tilt=%d g=%d: a perpendicular gap but L=%d, want 0 or 12", p, tilt, g, L)
			case !parallel && !perpendicular && (L == 0 || L == 6 || L == 12):
				t.Fatalf("p=%d tilt=%d g=%d: neither gap, yet L=%d is a resting length", p, tilt, g, L)
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
