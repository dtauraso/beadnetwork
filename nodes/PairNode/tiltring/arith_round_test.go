package tiltring

// arith_round_test.go — the closed-form arithmetic docs/pair-node/math/arith.html rests on: one
// round written without a case in it. (Was PairNode/arith_round_test.go.)
//
// See docs/process/testing-shape.md for what a test here may assert.

import (
	"strconv"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

// TestOneRoundIsSignAndRemainder is the other half of what docs/pair-node/math/arith.html rests on: ONE
// update round, written without a case in it.
//
// The page prints its numbers for a 24-point lattice: its 6 is a QUARTER TURN and its 12 a HALF
// TURN, both read off the ring. Everything below is written in those two, and swept on 24 and 48 —
// the two counts differ by a factor the 24-point lattice hides, since there a half turn and a
// quarter turn are the same distance apart as a quarter turn and zero.
//
// The trick is to keep the sign. AngleLength drops it, which is why the rule that uses AngleLength
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
//	                         h - |h - d| IS AngleLength. That abs keeps f exactly (checked
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
	r := NewRing(points)
	// q and h below are these two. The page's 6 and 12 are their values at 24 points.
	quarter, half := r.QuarterTurn, r.HalfTurn
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
	for _, m := range []Machine{
		{Mode: tiltvector.TiltMachineParallel},
		{Mode: tiltvector.TiltMachinePerpendicular},
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
		if m.Mode == tiltvector.TiltMachinePerpendicular {
			shift = quarter
			stops, between = between, stops
		}
		for arr := int32(0); arr < points; arr++ {
			a := r.At(arr)
			for tilt := int32(0); tilt < points; tilt++ {
				cur := r.At(tilt)

				// e reads the SUBTRACTION, not d — no reduction of any kind. It takes
				// t - a modulo a half turn, and the ring is two half turns, so every
				// representative of t - a on it gives the same e. Bringing it into range
				// first would be work the modulus immediately undoes.
				e := ((tilt-arr+shift)%half+half)%half - quarter

				// e = 0 at the stopping values and nowhere else; |e| = a quarter turn
				// exactly at the two values between them.
				gap := ((tilt-arr)%points + points) % points
				if (e == 0) != stops[gap] {
					t.Fatalf("mode=%v t-a=%d: e=%d, but stops=%v", m.Mode, gap, e, stops[gap])
				}
				if (abs32(e) == quarter) != between[gap] {
					t.Fatalf("mode=%v t-a=%d: |e|=%d, but between=%v", m.Mode, gap, abs32(e), between[gap])
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
							m.Mode, gap, up, down)
					}
					if (up-down+points)%points != half {
						t.Fatalf("mode=%v t-a=%d: the two stops %d and %d are not a half turn apart",
							m.Mode, gap, up, down)
					}
				}

				// THE BOTTOM IS WHERE THE SIGN WENT. A node draws two ends of one line,
				// t and t + h, and the two stopping values are one on each. Measure the
				// arrival against BOTH ends and there are two magnitudes, never negative:
				//
				//	top     = AngleLength(t,     a)
				//	bottom  = AngleLength(t + h, a) = h - top
				//
				// The end with the smaller reading is the one this node walks to, and that
				// is a comparison of two counts — no sign, no direction, no minus. Reduce
				// the pair to ONE number and the second reading has to come back as the
				// sign of the first, which is what e is.
				topL := cur.AngleLength(a)
				botL := cur.Opposite.AngleLength(a)
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
				// into it — b is a state the node already has (cur.Opposite), so writing
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
				b := cur.Opposite.Idx
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
				// c = 0, which is |c - q| at its largest). Ties go up in both, as Step
				// does. This is what docs/pair-node/math/arith.html prints as two branches per
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
				if m.Mode == tiltvector.TiltMachinePerpendicular {
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
					t.Fatalf("mode=%v t=%d a=%d c=%d: reducing changed the branch", m.Mode, tilt, arr, c)
				}

				// AND NEITHER d NOR e IS NEEDED. |d-q| against |e-q| is |c-(q-1)| against
				// |c-(q+1)|, whose answer is which side of the quarter c is on:
				//
				//	parallel        c < q -> up      c > q -> down     (c = q stands still)
				//	perpendicular   c >= q -> up     c < q -> down     (c = 0 stands still)
				fromC := c < quarter
				if m.Mode == tiltvector.TiltMachinePerpendicular {
					fromC = c >= quarter
				}
				if e != 0 && fromC != up {
					t.Fatalf("mode=%v t=%d a=%d c=%d: c against the quarter says up=%v, the two neighbours say %v",
						m.Mode, tilt, arr, c, fromC, up)
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
				if m.Mode == tiltvector.TiltMachinePerpendicular {
					byOwnCount = measure >= quarter
				}
				if e != 0 && byOwnCount != up {
					t.Fatalf("mode=%v t=%d a=%d: bit=%d measure=%d says up=%v, step says %v",
						m.Mode, tilt, arr, bit, measure, byOwnCount, up)
				}

				// h - c, and NOT reduced: at c = 0 the other end reads a half turn, h,
				// not 0. Reducing it collapses the two ends onto each other, which is the
				// one thing this measurement exists to keep apart.
				cr := half - c
				var byEnd bool // does t go up?
				switch {
				case acuteTop && m.Mode == tiltvector.TiltMachineParallel:
					byEnd = c < quarter
				case acuteTop:
					byEnd = c >= quarter
				case m.Mode == tiltvector.TiltMachineParallel:
					byEnd = cr > quarter // flipped
				default:
					byEnd = cr <= quarter // flipped
				}
				if e != 0 && byEnd != up {
					t.Fatalf("mode=%v t=%d a=%d: acuteTop=%v c=%d cr=%d says up=%v, step says %v",
						m.Mode, tilt, arr, acuteTop, c, cr, byEnd, up)
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
						m.Mode, tilt, arr, c, map[bool]string{true: "top", false: "bottom"}[c == u],
						got, (c+1)%half)
				}
				if e != 0 {
					wantNext := cur.Prev
					if up {
						wantNext = cur.Next
					}
					if got := steppedTop(m, cur, a); got != wantNext {
						t.Fatalf("mode=%v t=%d a=%d: |d-q|=%d |e-q|=%d chose %d, step chose %d",
							m.Mode, tilt, arr, abs32(d-quarter), abs32(eNbr-quarter), wantNext.Idx, got.Idx)
					}
				}

				// f comes straight off c, with no intermediate: parallel stops where c is
				// the quarter, perpendicular where c is 0, and the two are complements
				// because they stop at opposite ends of the same measurement.
				if m.Mode == tiltvector.TiltMachineParallel {
					if got, want := offBy(m, cur, a), abs32(c-quarter); got != want {
						t.Fatalf("parallel t=%d a=%d: c=%d gives |c-q|=%d but offBy=%d",
							tilt, arr, c, want, got)
					}
				} else if got, want := offBy(m, cur, a), quarter-abs32(c-quarter); got != want {
					t.Fatalf("perpendicular t=%d a=%d: c=%d gives q-|c-q|=%d but offBy=%d",
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
				normL := cur.Quarter.AngleLength(a)
				onNormalLine := normL == 0 || cur.Quarter.Opposite.AngleLength(a) == 0
				// AND f IS THE SMALLER OF THE TWO READINGS of that line. So the count and
				// the stop are one statement — "how far the arrival is off my line" — with
				// no subtraction from the quarter and no case for which arrangement.
				antiNormL := cur.Quarter.Opposite.AngleLength(a)
				line := [2]int32{topL, botL}
				if m.Mode == tiltvector.TiltMachineParallel {
					line = [2]int32{normL, antiNormL}
				}
				if got := offBy(m, cur, a); got != min(line[0], line[1]) {
					t.Fatalf("mode=%v t=%d a=%d: readings %d and %d, min=%d but offBy=%d",
						m.Mode, tilt, arr, line[0], line[1], min(line[0], line[1]), got)
				}

				stopped := e == 0
				if m.Mode == tiltvector.TiltMachinePerpendicular && stopped != onTiltLine {
					t.Fatalf("perpendicular t=%d a=%d: stopped=%v but on tilt line=%v",
						tilt, arr, stopped, onTiltLine)
				}
				if m.Mode == tiltvector.TiltMachineParallel && stopped != onNormalLine {
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
					t.Fatalf("mode=%v t=%d a=%d: remainder %d outside a half turn", m.Mode, tilt, arr, r12)
				}
				if got := offBy(m, cur, a); got != abs32(r12-quarter) {
					t.Fatalf("mode=%v t=%d a=%d: |r-q|=%d but offBy=%d",
						m.Mode, tilt, arr, abs32(r12-quarter), got)
				}
				wantLine := ((arr+quarter-shift)%half + half) % half
				if e == 0 && ((tilt%half)+half)%half != wantLine {
					t.Fatalf("mode=%v t=%d a=%d: stopped off the line — t mod a half turn = %d, want %d",
						m.Mode, tilt, arr, ((tilt%half)+half)%half, wantLine)
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
						m.Mode, arr, tilt, abs32(eFromAbs), abs32(e))
				}

				// And the angle length needs no min: |t - a| runs the ring, and the length
				// is its distance to the nearer end of that range — h - |h - |t - a||.
				l := half - abs32(half-absDiff)
				if cur.AngleLength(a) != l {
					t.Fatalf("a=%d t=%d: h-|h-|t-a||=%d but AngleLength=%d",
						arr, tilt, l, cur.AngleLength(a))
				}
				if want := min(absDiff, points-absDiff); l != want {
					t.Fatalf("a=%d t=%d: got %d but min(|t-a|, %d-|t-a|)=%d", arr, tilt, l, points, want)
				}

				if got := offBy(m, cur, a); got != abs32(e) {
					t.Fatalf("mode=%v a=%d t=%d |t-a|=%d: |e|=%d but offBy=%d",
						m.Mode, arr, tilt, absDiff, abs32(e), got)
				}
				// AND THE PAGE'S OWN ARITHMETIC, asked of the machine rather than
				// rebuilt here. Everything above derives the page's forms in this test
				// and checks the machine agrees; these two call what machine.go computes
				// the page's way and check it against what it computes the other way. That
				// is the comparison the switch-over rests on, so it is made on every pair
				// of both lattices, including the stopped ones.
				if got, _ := cur.NearerEndCount(a); got != c {
					t.Fatalf("mode=%v t=%d a=%d: NearerEndCount=%d but c=%d", m.Mode, tilt, arr, got, c)
				}
				if _, atBottom := cur.NearerEndCount(a); atBottom == acuteTop {
					t.Fatalf("mode=%v t=%d a=%d: NearerEndCount says atBottom=%v with u=%d",
						m.Mode, tilt, arr, atBottom, u)
				}
				// THE RULE ASKS ITS SECOND QUESTION WITHOUT A DISTANCE. Step is one
				// subtraction against a quarter turn, so the direction it picks must
				// agree with the nearer of the two ways round — worked out here, since
				// the machine no longer works it out anywhere.
				if !m.Settled(cur, a) {
					upDist := ((m.Stopping().At(r)-c)%half + half) % half
					wantUp := upDist <= quarter
					if wantUp != (upDist <= half-upDist) {
						t.Fatalf("mode=%v t=%d a=%d c=%d: up-count %d against a quarter says %v, against the way down says %v",
							m.Mode, tilt, arr, c, upDist, wantUp, upDist <= half-upDist)
					}
				}

				if (e == 0) != m.Settled(cur, a) {
					t.Fatalf("mode=%v a=%d t=%d d=%d: e=%d but settled=%v",
						m.Mode, arr, tilt, d, e, m.Settled(cur, a))
				}
				if e == 0 {
					continue
				}
				want := ((tilt-sign(e))%points + points) % points
				if got := steppedTop(m, cur, a).Idx; got != want {
					t.Fatalf("mode=%v a=%d t=%d d=%d e=%d: step gave %d, t-sign(e)=%d",
						m.Mode, arr, tilt, d, e, got, want)
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
