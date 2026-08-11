package tiltring

// arith_fromrest_test.go — the resting lengths derived from the possible gaps, which
// docs/pair-node/rules/audit.html and update-rules rest on.
//
// See docs/process/testing-shape.md for what a test here may assert. (Was
// PairNode/arith_fromrest_test.go.)

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

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
// The two being complements is the audit's "one is the other upside down" (docs/pair-node/rules/audit.html),
// here as arithmetic rather than as a picture. This sweeps both lattices the model runs on, every
// tilt against every arrival, so the page cannot claim a shortcut the code does not honour.
func TestFromRestIsTheQuarterOffset(t *testing.T) {
	for _, points := range []int32{24, 48} {
		r := NewRing(points)
		perp := Machine{Mode: tiltvector.TiltMachinePerpendicular}
		par := Machine{Mode: tiltvector.TiltMachineParallel}
		for tilt := int32(0); tilt < points; tilt++ {
			for arr := int32(0); arr < points; arr++ {
				from, a := r.At(tilt), r.At(arr)
				q := abs32(from.AngleLength(a) - r.QuarterTurn)
				if got := offBy(par, from, a); got != q {
					t.Fatalf("points=%d tilt=%d arrival=%d: parallel offBy=%d, want q=%d",
						points, tilt, arr, got, q)
				}
				if got := offBy(perp, from, a); got != r.QuarterTurn-q {
					t.Fatalf("points=%d tilt=%d arrival=%d: perpendicular offBy=%d, want quarter-q=%d",
						points, tilt, arr, got, r.QuarterTurn-q)
				}
			}
		}
	}
}
