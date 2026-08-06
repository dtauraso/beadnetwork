package Wiring

import "testing"

// TiltVectorIsAcute is now pure integer index arithmetic on the θ-only 24-step lattice
// (task/drop-tilt-vector-phi) — no trig, no epsilon. These tests exercise that whole
// lattice directly, per docs/testing-shape.md: one function's own decision, no
// cross-goroutine involvement.

// classify names the three regions TiltVectorIsAcute must produce for a given difference
// d in [0, FullTurnThetaIdx): acute (d < 6 or d > 18), perpendicular (d == 6 or d == 18),
// obtuse otherwise. Used to build the expected table below without hand-listing every
// index.
func wantAcuteForDiff(d int32) bool {
	return d < PerpendicularThetaIdx || d > FullTurnThetaIdx-PerpendicularThetaIdx
}

// Every d from 0..23 (the whole lattice) classified: the two perpendicular values (6, 18)
// must NOT be acute, and their immediate neighbours (5,7 and 17,19) MUST be acute.
func TestTiltVectorIsAcuteAcrossWholeLattice(t *testing.T) {
	for d := int32(0); d < FullTurnThetaIdx; d++ {
		a := TiltVectorMsg{ThetaIdx: 0}
		b := TiltVectorMsg{ThetaIdx: d}
		got := TiltVectorIsAcute(a, b)
		want := wantAcuteForDiff(d)
		if got != want {
			t.Errorf("d=%d: TiltVectorIsAcute = %v, want %v", d, got, want)
		}
	}
	// Explicit spot-check of the two perpendicular values and their neighbours, since these
	// are the exact boundary the old float-epsilon test existed to guard.
	for _, d := range []int32{PerpendicularThetaIdx, FullTurnThetaIdx - PerpendicularThetaIdx} {
		a := TiltVectorMsg{ThetaIdx: 0}
		b := TiltVectorMsg{ThetaIdx: d}
		if TiltVectorIsAcute(a, b) {
			t.Errorf("d=%d (perpendicular) must NOT be acute", d)
		}
	}
	// The neighbour ONE STEP TOWARD ACUTE from each perpendicular value must be acute (5 and
	// 19); the neighbour one step the OTHER way (7 and 17) is obtuse, not acute — asserted
	// above by the full-lattice loop, which already covers it via wantAcuteForDiff.
	for _, d := range []int32{PerpendicularThetaIdx - 1, FullTurnThetaIdx - PerpendicularThetaIdx + 1} {
		a := TiltVectorMsg{ThetaIdx: 0}
		b := TiltVectorMsg{ThetaIdx: d}
		if !TiltVectorIsAcute(a, b) {
			t.Errorf("d=%d (neighbour of a perpendicular value, on the acute side) must BE acute", d)
		}
	}
}

// Node1's base direction subtracts, so NEGATIVE θ indices are the common case in
// production. A sign-keeping `%` (Go's own) gets exactly this wrong; the reduction here
// must be FLOOR-correct regardless of which operand (or both) is negative.
func TestTiltVectorIsAcuteAcrossWholeLatticeWithNegativeIndices(t *testing.T) {
	bases := []int32{0, -37, -12, 5, -6}
	for _, base := range bases {
		for d := int32(0); d < FullTurnThetaIdx; d++ {
			a := TiltVectorMsg{ThetaIdx: base}
			b := TiltVectorMsg{ThetaIdx: base + d}
			got := TiltVectorIsAcute(a, b)
			want := wantAcuteForDiff(d)
			if got != want {
				t.Errorf("base=%d d=%d (a=%d b=%d): TiltVectorIsAcute = %v, want %v",
					base, d, a.ThetaIdx, b.ThetaIdx, got, want)
			}
		}
	}
	// Both operands negative, explicitly, since that is the case a sign-keeping `%` breaks
	// most visibly (negative % negative can still yield a negative or otherwise
	// non-canonical remainder depending on operand signs).
	a := TiltVectorMsg{ThetaIdx: -30}
	b := TiltVectorMsg{ThetaIdx: -6}
	if !TiltVectorIsAcute(a, b) {
		t.Fatalf("a=-30 b=-6 (diff=24 steps = a full turn = same direction) must be acute")
	}
}

// The two-dot property the straightening rule depends on: top-acute and bottom-acute (the
// bottom being the top's exact antipode, HalfTurnThetaIdx away) are MUTUALLY EXCLUSIVE, and
// neither is acute exactly when the arrival is perpendicular to the top (and therefore also
// to the bottom, its antipode). Asserted directly against the new integer implementation
// across the whole lattice, negative indices included.
func TestTopAndBottomAcuteAreMutuallyExclusiveAndBothFalseAtPerpendicular(t *testing.T) {
	tops := []int32{0, 5, -9, -37, 100}
	for _, top := range tops {
		bottom := TiltVectorMsg{ThetaIdx: top + HalfTurnThetaIdx}
		topVec := TiltVectorMsg{ThetaIdx: top}
		for d := int32(0); d < FullTurnThetaIdx; d++ {
			arrived := TiltVectorMsg{ThetaIdx: top + d}
			acuteTop := TiltVectorIsAcute(arrived, topVec)
			acuteBottom := TiltVectorIsAcute(arrived, bottom)
			if acuteTop && acuteBottom {
				t.Fatalf("top=%d d=%d: both top and bottom reported acute — must be mutually exclusive", top, d)
			}
			atPerp := d == PerpendicularThetaIdx || d == FullTurnThetaIdx-PerpendicularThetaIdx
			if atPerp && (acuteTop || acuteBottom) {
				t.Fatalf("top=%d d=%d (perpendicular to top): want neither dot acute, got acuteTop=%v acuteBottom=%v",
					top, d, acuteTop, acuteBottom)
			}
		}
	}
}
