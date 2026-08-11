package tiltring

// arith_walk_test.go — the walk as a closed form, which docs/pair-node/math/arith.html rests on.
//
// See docs/process/testing-shape.md for what a test here may assert. (Was
// PairNode/arith_walk_test.go.)

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

// TestTheWalkIsClosedForm holds the WHERE IT STOPS claim that docs/pair-node/math/arith.html makes, and
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
// The direction is read the way Step reads it — whichever neighbour is closer, ties up — and then
// held FIXED for the rest of the walk here. That is the real content: it is not obvious that the
// sign cannot flip halfway, and at the fold points (l = 0 and l = 12) the two neighbours score the
// same, which is exactly where a walk could turn back. It does not.
//
// This does NOT say the pair is closed form. A live exchange moves both ends, so a is not held;
// arith.html says so on the page rather than leaving the reader to assume otherwise.
func TestTheWalkIsClosedForm(t *testing.T) {
	const points = 24
	r := NewRing(points)
	for _, m := range []Machine{
		{Mode: tiltvector.TiltMachineParallel},
		{Mode: tiltvector.TiltMachinePerpendicular},
	} {
		for arr := int32(0); arr < points; arr++ {
			a := r.At(arr)
			for tilt := int32(0); tilt < points; tilt++ {
				cur := r.At(tilt)
				f := offBy(m, cur, a)

				s := int32(-1) // Step's own rule: up unless down is strictly closer
				if offBy(m, cur.Next, a) <= offBy(m, cur.Prev, a) {
					s = 1
				}

				steps := int32(0)
				for !m.Settled(cur, a) {
					cur = steppedTop(m, cur, a)
					steps++
					if steps > 2*points {
						t.Fatalf("mode=%v arrival=%d tilt=%d: never settled", m.Mode, arr, tilt)
					}
				}
				if steps != f {
					t.Fatalf("mode=%v arrival=%d tilt=%d: settled after %d arrivals, f said %d",
						m.Mode, arr, tilt, steps, f)
				}
				if want := ((tilt+s*f)%points + points) % points; cur.Idx != want {
					t.Fatalf("mode=%v arrival=%d tilt=%d: stopped at %d, t+s*f said %d (s=%d f=%d)",
						m.Mode, arr, tilt, cur.Idx, want, s, f)
				}
			}
		}
	}
}
