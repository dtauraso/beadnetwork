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
	"strconv"
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
// The page prints its numbers for a 24-point lattice: its 6 is a QUARTER TURN and its 12 a HALF
// TURN, both read off the ring. Everything below is written in those two, and swept on 24 and 48 —
// the two counts differ by a factor the 24-point lattice hides, since there a half turn and a
// quarter turn are the same distance apart as a quarter turn and zero.
//
// The trick is to keep the sign. angleLength drops it, which is why the rule that uses angleLength
// needs a comparison to get direction back. Work from the plain subtraction t - a instead — which
// may be negative and may fall outside the ring, and is left that way — and both arrangements are
// the same statement, since the stopping values sit a half turn apart and the arrangement is a
// SHIFT of a quarter inside the modulus:
//
//	parallel       e = ((t - a) mod h) - q          stops where t - a = q or q + h
//	perpendicular  e = ((t - a + q) mod h) - q      stops where t - a = 0 or h
//
// e is then how many slots t is from a stopping value and on which side of it, and the whole round
// is t_after = t - sign(e), with |e| equal to fromRest. No branch, no minimum, no list.
//
// Two things that look necessary and are not, both checked below:
//
//	putting t - a in range   the ring is two half turns, so t-a and t-a ± points all give the
//	                         same e. Bringing it into range first is work the modulus undoes.
//	using d                  d = |t - a| is the same subtraction with the abs on it, and
//	                         h - |h - d| IS angleLength. That abs keeps f exactly (checked
//	                         below) and loses the TURN: with t=0, an arrival one slot up must
//	                         turn down and one slot down must turn up, yet d and l read the
//	                         same for both. So e reads t - a and never d.
//
// A comparison rule needs telling what to prefer when a tilt sits the same distance from two
// stopping values. This needs nothing, and NOT because the -q in the formula quietly settles it:
// at those values the two stopping values are a half turn apart, so they are the same LINE at two
// indices, and a quarter up and a quarter down stop equally soon on the same arrangement.
// There is no preference to encode. Whatever sign e takes there picks an index, not an outcome —
// checked below, so the claim does not rest on the shape of the expression.
func TestOneRoundIsSignAndRemainder(t *testing.T) {
	// BOTH LATTICES, because the numbers the page prints are the 24-point lattice's names for
	// two ring counts, not constants. Written with 6 and 12 in it this swept 1152 pairs that
	// cannot tell a half turn from a whole one — at 24 the wrong one of the two is off by
	// exactly the factor that leaves the arithmetic looking right. quarter and half are read
	// off the ring here, and every number below is one of them.
	for _, points := range []int32{24, 48} {
		t.Run(strconv.Itoa(int(points)), func(t *testing.T) { oneRoundSweep(t, points) })
	}
}

// oneRoundSweep is TestOneRoundIsSignAndRemainder's body on one lattice.
func oneRoundSweep(t *testing.T, points int32) {
	r := newRing(points)
	// q and h below are these two. The page's 6 and 12 are their values at 24 points.
	quarter, half := r.quarterTurn, r.halfTurn
	var sawTop, sawBottom int // c comes from the top end, or from the bottom end
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
		// The two stopping values of t - a, and the two values exactly between them. A tilt
		// is a line, so t and t + h are the same tilt, so each arrangement stops at two
		// values a half turn apart — which leaves a value a quarter from each, and that is
		// where a rule that compares neighbours has to be told what to prefer.
		//
		// The two sets are each other's: one arrangement stops where the other is stuck
		// between, which is the + q inside the modulus seen from the other end.
		shift := int32(0)
		stops := map[int32]bool{quarter: true, quarter + half: true}
		between := map[int32]bool{0: true, half: true}
		if m.mode == Wiring.TiltMachinePerpendicular {
			shift = quarter
			stops, between = between, stops
		}
		for arr := int32(0); arr < points; arr++ {
			a := r.at(arr)
			for tilt := int32(0); tilt < points; tilt++ {
				cur := r.at(tilt)

				// e reads the SUBTRACTION, not d — no reduction of any kind. It takes
				// t - a modulo a half turn, and the ring is two half turns, so every
				// representative of t - a on it gives the same e. Bringing it into range
				// first would be work the modulus immediately undoes.
				e := ((tilt-arr+shift)%half+half)%half - quarter

				// e = 0 at the stopping values and nowhere else; |e| = a quarter turn
				// exactly at the two values between them.
				gap := ((tilt-arr)%points + points) % points
				if (e == 0) != stops[gap] {
					t.Fatalf("mode=%v t-a=%d: e=%d, but stops=%v", m.mode, gap, e, stops[gap])
				}
				if (abs32(e) == quarter) != between[gap] {
					t.Fatalf("mode=%v t-a=%d: |e|=%d, but between=%v", m.mode, gap, abs32(e), between[gap])
				}

				// AND THE TIE HAS NO WRONG ANSWER, so nothing rests on the sign there.
				// The two stopping values are a half turn apart and a tilt is a line, so
				// they are the SAME arrangement at two indices. From a value between them,
				// a quarter up and a quarter down both stop, in the same number of
				// arrivals, on the same line. The minus in the formula picks which index
				// it walks to; it does not pick between a right and a wrong answer.
				if between[gap] {
					up, down := ((gap+quarter)%points+points)%points, ((gap-quarter)%points+points)%points
					if !stops[up] || !stops[down] {
						t.Fatalf("mode=%v t-a=%d: a quarter either way gives %d and %d, not both stops",
							m.mode, gap, up, down)
					}
					if (up-down+points)%points != half {
						t.Fatalf("mode=%v t-a=%d: the two stops %d and %d are not a half turn apart",
							m.mode, gap, up, down)
					}
				}

				// THE BOTTOM IS WHERE THE SIGN WENT. A node draws two ends of one line,
				// t and t + h, and the two stopping values are one on each. Measure the
				// arrival against BOTH ends and there are two magnitudes, never negative:
				//
				//	top     = angleLength(t,     a)
				//	bottom  = angleLength(t + h, a) = h - top
				//
				// The end with the smaller reading is the one this node walks to, and that
				// is a comparison of two counts — no sign, no direction, no minus. Reduce
				// the pair to ONE number and the second reading has to come back as the
				// sign of the first, which is what e is.
				topL := cur.angleLength(a)
				botL := cur.opposite.angleLength(a)
				if topL+botL != half {
					t.Fatalf("t=%d a=%d: top=%d bottom=%d, do not sum to a half turn",
						tilt, arr, topL, botL)
				}

				// Measured from the two INDICES, no abs anywhere. Count from each end up
				// to the arrival:
				//
				//	b = (t + h) mod points         the bottom's own index
				//	u = (t - a) mod points         from the top
				//	v = (b - a) mod points         from the bottom
				//
				// v is measured FROM THE BOTTOM, not from the top with a half turn folded
				// into it — b is a state the node already has (cur.opposite), so writing
				// t + h - a here would be reaching past the vector to rebuild it.
				//
				// ONE TEST, applied to each count on its own: under a half turn, the
				// count IS the distance; at or over, the distance is points minus it. No
				// cross-reference between the two ends, and no third quantity — the
				// distance from an end is that end's own count, tested.
				// ONE letter holds the result. The two counts differ by h, so exactly one
				// of them is under h, and that one already IS the acute angle at its own
				// end — nothing has to be computed for it and nothing is overwritten:
				//
				//	c = whichever of u, v is under a half turn
				//
				// The other end is h - c, and nothing downstream needs it: |c - q| is the
				// same either way, since |(h - c) - q| = |q - c|.
				b := cur.opposite.idx
				u := ((tilt-arr)%points + points) % points
				v := ((b-arr)%points + points) % points
				c := u
				if u >= half {
					c = v
				}
				if c >= half {
					t.Fatalf("t=%d a=%d: u=%d v=%d, neither under a half turn", tilt, arr, u, v)
				}

				// THE SECOND COUNT IS NOT NEEDED. The bottom is the top's other side, so
				// v is u a half turn on — and picking whichever is under a half turn is the
				// same as taking u modulo a half turn:
				//
				//	c = (t - a) mod h
				//
				// No v, no test, and nothing to remember about which end was acute.
				if got := ((tilt-arr)%half + half) % half; got != c {
					t.Fatalf("t=%d a=%d: (t-a) mod a half turn = %d but c = %d", tilt, arr, got, c)
				}
				if c != topL && c != botL {
					t.Fatalf("t=%d a=%d: c=%d is neither the top angle %d nor the bottom %d",
						tilt, arr, c, topL, botL)
				}
				// THE TWO ARRANGEMENTS DIFFER BY THE DIRECTION OF ONE INEQUALITY. Both
				// compare |c - q| at the two tilts one slot away; parallel walks to the smaller
				// (its stop is c = q) and perpendicular toward the larger (its stop is
				// c = 0, which is |c - q| at its largest). Ties go up in both, as step
				// does. This is what docs/pair-node/arith.html prints as two branches per
				// arrangement, so it cannot be left as "nearer its stop".
				cAt := func(x int32) int32 {
					uu := ((x-arr)%points + points) % points
					if uu < half {
						return uu
					}
					return ((x+half-arr)%points + points) % points
				}
				// d and e are the page's names for c worked out at t+1 and at t-1 — and
				// they need no counts of their own: turning one slot moves BOTH counts by
				// one, so the acute angle just steps round a ring of h.
				d, eNbr := cAt(tilt+1), cAt(tilt-1)
				if d != (c+1)%half || eNbr != (c+half-1)%half {
					t.Fatalf("t=%d a=%d: c=%d, but d=%d e=%d (want %d and %d)",
						tilt, arr, c, d, eNbr, (c+1)%half, (c+half-1)%half)
				}
				up := abs32(d-quarter) <= abs32(eNbr-quarter)
				rawUp := abs32(c+1-quarter) <= abs32(c-1-quarter)
				if m.mode == Wiring.TiltMachinePerpendicular {
					up = abs32(d-quarter) >= abs32(eNbr-quarter)
					rawUp = abs32(c+1-quarter) >= abs32(c-1-quarter)
				}

				// The mod h is what makes d and e ANGLES: without it, c-1 at c = 0 is -1,
				// which is not an angle, and |−1 − q| is not a reading of anything.
				//
				// c = 0 is the ONLY place the reduced and un-reduced numbers disagree
				// (|e-q| is q-1 reduced, q+1 not) — and it is perpendicular's stop, so no
				// comparison runs there. Wherever a comparison IS run, the two agree,
				// which is why the modulus can be justified as "d and e are angles"
				// rather than as a correction the branches depend on.
				if e != 0 && rawUp != up {
					t.Fatalf("mode=%v t=%d a=%d c=%d: reducing changed the branch", m.mode, tilt, arr, c)
				}

				// AND NEITHER d NOR e IS NEEDED. |d-q| against |e-q| is |c-(q-1)| against
				// |c-(q+1)|, whose answer is which side of the quarter c is on:
				//
				//	parallel        c < q -> up      c > q -> down     (c = q stands still)
				//	perpendicular   c >= q -> up     c < q -> down     (c = 0 stands still)
				fromC := c < quarter
				if m.mode == Wiring.TiltMachinePerpendicular {
					fromC = c >= quarter
				}
				if e != 0 && fromC != up {
					t.Fatalf("mode=%v t=%d a=%d c=%d: c against the quarter says up=%v, the two neighbours say %v",
						m.mode, tilt, arr, c, fromC, up)
				}

				// THE ACUTE END, AND THE SAME RULE STATED ON IT.
				//
				// Which end a is nearer is a bit: the top when u is under a half turn, the
				// bottom otherwise. Stated on the BOTTOM the angle is measured the other
				// way round — the complementary angle, cr = h - c — and every comparison
				// flips, because counting counter-clockwise reverses the order.
				//
				// The two descriptions must pick the same move; that is the whole claim.
				acuteTop := u < half

				// STATED ON THE END'S OWN COUNT there is no reversal at all: when the
				// bottom is nearer, v is itself under a half turn and the comparisons read
				// exactly as the top's do. The reversal only appears if the bottom is
				// measured by the COMPLEMENTARY angle, h - c, which counts the other way.
				bit := 0
				measure := u
				if !acuteTop {
					bit, measure = 1, v
				}
				byOwnCount := measure < quarter
				if m.mode == Wiring.TiltMachinePerpendicular {
					byOwnCount = measure >= quarter
				}
				if e != 0 && byOwnCount != up {
					t.Fatalf("mode=%v t=%d a=%d: bit=%d measure=%d says up=%v, step says %v",
						m.mode, tilt, arr, bit, measure, byOwnCount, up)
				}

				// h - c, and NOT reduced: at c = 0 the other end reads a half turn, h,
				// not 0. Reducing it collapses the two ends onto each other, which is the
				// one thing this measurement exists to keep apart.
				cr := half - c
				var byEnd bool // does t go up?
				switch {
				case acuteTop && m.mode == Wiring.TiltMachineParallel:
					byEnd = c < quarter
				case acuteTop:
					byEnd = c >= quarter
				case m.mode == Wiring.TiltMachineParallel:
					byEnd = cr > quarter // flipped
				default:
					byEnd = cr <= quarter // flipped
				}
				if e != 0 && byEnd != up {
					t.Fatalf("mode=%v t=%d a=%d: acuteTop=%v c=%d cr=%d says up=%v, step says %v",
						m.mode, tilt, arr, acuteTop, c, cr, byEnd, up)
				}

				// AND WHICH END c CAME FROM DOES NOT ENTER THE RULE. The two ends turn
				// TOGETHER — b = t + h, so t+1 makes b+1 — which means u and v both gain
				// one and c steps up whichever end supplied it. There is no end whose
				// update runs backwards, and no variable is needed to remember which one
				// it was: both cases are swept here, and the same c-against-q rule holds.
				if c == u {
					sawTop++
				} else {
					sawBottom++
				}
				if got := cAt(tilt + 1); got != (c+1)%half {
					t.Fatalf("mode=%v t=%d a=%d: c=%d came from the %s, but t+1 gives %d not %d",
						m.mode, tilt, arr, c, map[bool]string{true: "top", false: "bottom"}[c == u],
						got, (c+1)%half)
				}
				if e != 0 {
					wantNext := cur.prev
					if up {
						wantNext = cur.next
					}
					if got := m.step(cur, a); got != wantNext {
						t.Fatalf("mode=%v t=%d a=%d: |d-q|=%d |e-q|=%d chose %d, step chose %d",
							m.mode, tilt, arr, abs32(d-quarter), abs32(eNbr-quarter), wantNext.idx, got.idx)
					}
				}

				// f comes straight off c, with no intermediate: parallel stops where c is
				// the quarter, perpendicular where c is 0, and the two are complements
				// because they stop at opposite ends of the same measurement.
				if m.mode == Wiring.TiltMachineParallel {
					if got, want := m.fromRest(cur, a), abs32(c-quarter); got != want {
						t.Fatalf("parallel t=%d a=%d: c=%d gives |c-q|=%d but fromRest=%d",
							tilt, arr, c, want, got)
					}
				} else if got, want := m.fromRest(cur, a), quarter-abs32(c-quarter); got != want {
					t.Fatalf("perpendicular t=%d a=%d: c=%d gives q-|c-q|=%d but fromRest=%d",
						tilt, arr, c, want, got)
				}

				// And each arrangement stops when the arrival lands ON one of the node's
				// own drawn lines — which is what the two stopping values ARE:
				//
				//	perpendicular   the arrival lies on the TILT line, t or t+h
				//	parallel        the arrival lies on the NORMAL line, t+q or t+q+h
				//
				// Both are stated with unsigned readings and no direction at all.
				onTiltLine := topL == 0 || botL == 0
				normL := cur.quarter.angleLength(a)
				onNormalLine := normL == 0 || cur.quarter.opposite.angleLength(a) == 0
				// AND f IS THE SMALLER OF THE TWO READINGS of that line. So the count and
				// the stop are one statement — "how far the arrival is off my line" — with
				// no subtraction from the quarter and no case for which arrangement.
				antiNormL := cur.quarter.opposite.angleLength(a)
				line := [2]int32{topL, botL}
				if m.mode == Wiring.TiltMachineParallel {
					line = [2]int32{normL, antiNormL}
				}
				if got := m.fromRest(cur, a); got != min(line[0], line[1]) {
					t.Fatalf("mode=%v t=%d a=%d: readings %d and %d, min=%d but fromRest=%d",
						m.mode, tilt, arr, line[0], line[1], min(line[0], line[1]), got)
				}

				stopped := e == 0
				if m.mode == Wiring.TiltMachinePerpendicular && stopped != onTiltLine {
					t.Fatalf("perpendicular t=%d a=%d: stopped=%v but on tilt line=%v",
						tilt, arr, stopped, onTiltLine)
				}
				if m.mode == Wiring.TiltMachineParallel && stopped != onNormalLine {
					t.Fatalf("parallel t=%d a=%d: stopped=%v but on normal line=%v",
						tilt, arr, stopped, onNormalLine)
				}

				// NO SIGN AT ALL, once the answer is read as a LINE rather than an index.
				// The two stopping values are a half turn apart — one on this node's top,
				// one on its bottom — so they are ONE arrangement, and which of the two a
				// walk reaches is a fact about indices, not about the pair. Taken mod h,
				// where a tilt and its bottom are the same number, the whole rule has no
				// direction in it:
				//
				//	u = (t - a + shift) mod h     under a half turn, never negative
				//	f = |u - q|                   arrivals, a magnitude
				//	ends on the line  t = a + q (mod h)  parallel
				//	                  t = a     (mod h)  perpendicular
				//
				// e and sign(e) name the index it walks to. Nothing above needs them.
				r12 := ((tilt-arr+shift)%half + half) % half
				if r12 < 0 || r12 >= half {
					t.Fatalf("mode=%v t=%d a=%d: remainder %d outside a half turn", m.mode, tilt, arr, r12)
				}
				if got := m.fromRest(cur, a); got != abs32(r12-quarter) {
					t.Fatalf("mode=%v t=%d a=%d: |r-q|=%d but fromRest=%d",
						m.mode, tilt, arr, abs32(r12-quarter), got)
				}
				wantLine := ((arr+quarter-shift)%half + half) % half
				if e == 0 && ((tilt%half)+half)%half != wantLine {
					t.Fatalf("mode=%v t=%d a=%d: stopped off the line — t mod a half turn = %d, want %d",
						m.mode, tilt, arr, ((tilt%half)+half)%half, wantLine)
				}

				// Taking the abs BEFORE the modulus keeps the size and loses the turn:
				// |t - a| makes the two sides of a identical, but a node one slot below it
				// must turn opposite to one slot above — t=0, a=1 turns down to the last
				// index, while t=0 with a there turns up to 1. Same distance, opposite
				// direction.
				absDiff := abs32(tilt - arr)
				eFromAbs := ((absDiff+shift)%half+half)%half - quarter
				if abs32(eFromAbs) != abs32(e) {
					t.Fatalf("mode=%v a=%d t=%d: |e| from |t-a| is %d, from t-a is %d",
						m.mode, arr, tilt, abs32(eFromAbs), abs32(e))
				}

				// And the angle length needs no min: |t - a| runs the ring, and the length
				// is its distance to the nearer end of that range — h - |h - |t - a||.
				l := half - abs32(half-absDiff)
				if cur.angleLength(a) != l {
					t.Fatalf("a=%d t=%d: h-|h-|t-a||=%d but angleLength=%d",
						arr, tilt, l, cur.angleLength(a))
				}
				if want := min(absDiff, points-absDiff); l != want {
					t.Fatalf("a=%d t=%d: got %d but min(|t-a|, %d-|t-a|)=%d", arr, tilt, l, points, want)
				}

				if got := m.fromRest(cur, a); got != abs32(e) {
					t.Fatalf("mode=%v a=%d t=%d |t-a|=%d: |e|=%d but fromRest=%d",
						m.mode, arr, tilt, absDiff, abs32(e), got)
				}
				// AND THE PAGE'S OWN ARITHMETIC, asked of the machine rather than
				// rebuilt here. Everything above derives the page's forms in this test
				// and checks the machine agrees; these two call what machine.go computes
				// the page's way and check it against what it computes the other way. That
				// is the comparison the switch-over rests on, so it is made on every pair
				// of both lattices, including the stopped ones.
				if got, _ := cur.nearerEndCount(a); got != c {
					t.Fatalf("mode=%v t=%d a=%d: nearerEndCount=%d but c=%d", m.mode, tilt, arr, got, c)
				}
				if _, atBottom := cur.nearerEndCount(a); atBottom == acuteTop {
					t.Fatalf("mode=%v t=%d a=%d: nearerEndCount says atBottom=%v with u=%d",
						m.mode, tilt, arr, atBottom, u)
				}
				if got := m.stopsAtCount(cur, a); got != m.settled(cur, a) {
					t.Fatalf("mode=%v t=%d a=%d c=%d: stopsAtCount=%v but settled=%v",
						m.mode, tilt, arr, c, got, m.settled(cur, a))
				}
				// The direction is compared as the STATE CHOSEN, not as a distance: c is
				// (t - a) and not (a - t), and the two agree at every stop while
				// disagreeing in the turn, so a swapped subtraction survives settled and
				// shows up only here.
				if e != 0 {
					wantByCount := cur.prev
					if m.countGoesUp(cur, a) {
						wantByCount = cur.next
					}
					if got := m.step(cur, a); got != wantByCount {
						t.Fatalf("mode=%v t=%d a=%d c=%d: countGoesUp chose %d, step chose %d",
							m.mode, tilt, arr, c, wantByCount.idx, got.idx)
					}
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
	// Both ends really do supply c across the sweep, so the claim that it does not
	// matter which one did is a claim about cases that occur, not an empty one.
	if sawTop == 0 || sawBottom == 0 {
		t.Fatalf("c came from the top %d times and the bottom %d — one case never happened",
			sawTop, sawBottom)
	}
}

// TestTheWalkIsClosedForm holds the WHERE IT STOPS claim that docs/pair-node/arith.html makes, and
// two more the page deliberately does not. For a HELD arrival, swept over both modes and every
// (arrival, tilt) pair:
//
//	where it stops              =  t + s*f    THIS is on the page, as a line mod 12
//	how many messages it takes  =  f          not on the page: a pair's message count is an
//	                                          outcome of the exchange, not something a node
//	                                          computes, and printing it invited it to be read
//	                                          as a quantity a node holds
//	the direction               =  s          decided by the first message and never revisited
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
