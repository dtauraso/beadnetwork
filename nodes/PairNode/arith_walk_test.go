package PairNode

// arith_walk_test.go — the walk as a closed form, which docs/pair-node/math/arith.html rests on.
//
// See docs/process/testing-shape.md for what a test here may assert.

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring"
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
				f := offBy(m, cur, a)

				s := int32(-1) // step's own rule: up unless down is strictly closer
				if offBy(m, cur.next, a) <= offBy(m, cur.prev, a) {
					s = 1
				}

				steps := int32(0)
				for !m.settled(cur, a) {
					cur = steppedTop(m, cur, a)
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
