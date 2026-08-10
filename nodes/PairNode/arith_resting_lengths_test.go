package PairNode

// arith_resting_lengths_test.go — the resting lengths derived from the possible gaps, which
// docs/pair-node/math/arith.html rests on.
//
// See docs/process/testing-shape.md for what a test here may assert.

import "testing"

// TestRestingLengthsFollowFromTheGaps checks that the resting lengths are DERIVED rather than
// chosen. A tilt is a LINE, so reversing it changes nothing and tilt and tilt+12 are the same
// tilt — which leaves exactly four gaps g between the two tilts that the pair can hold:
//
//	parallel       g = 0 or 12     the two lines coincide
//	perpendicular  g = 6 or 18     a quarter turn between them
//
// What ARRIVES is the partner's normal, a = p + 6, already a quarter turn on, so the angle
// length this node measures is the gap with that quarter turn taken off — which turns those
// four gaps into L = 6 and L in {0, 12}. That is where stoppingCounts comes from — counted from
// the nearer end, {0, 12} is the single count 0 — and this
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
