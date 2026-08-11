package beadindex

import (
	"math"
	"testing"

	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// lit_bead_index_test.go — beadindex.LitBeadIndex: every index is reachable as traversal
// progress sweeps [0,1), equal elapsed time lights the same bead regardless of an edge's own
// step count, one bead index advances per dwell, and t outside [0,1) lights nothing.

// THE REGRESSION TEST for the reported bug: every bead index must be reachable as traversal
// progress t sweeps the whole edge. The bug was two coordinate systems — lighting quantised
// distance against the wire's PORT-TO-PORT arc while the chain was laid out to a longer
// center-distance-minus-radii span — so the tail of the chain could never be lit, however far
// t climbed. Under the bead lattice both layout and lighting read the SAME integer step
// count, so this now sweeps t directly against that integer.
func TestChainBeadsEveryIndexIsReachable(t *testing.T) {
	const steps = 28 // an arbitrary, but multi-bead, step count
	seen := make(map[int]bool, steps)
	const sweeps = 100000
	for i := 0; i <= sweeps; i++ {
		tt := float64(i) / float64(sweeps)
		if tt >= 1 {
			tt = math.Nextafter(1, 0) // sweep right up to, but not touching, t=1
		}
		idx, ok := LitBeadIndex(tt, steps)
		if !ok {
			continue
		}
		seen[idx] = true
	}
	for i := 0; i < steps; i++ {
		if !seen[i] {
			t.Errorf("bead index %d of %d was never reached while sweeping t in [0,1) — unreachable tail bead", i, steps)
		}
	}
	if len(seen) != steps {
		t.Errorf("saw %d distinct indices, want exactly %d (0..steps-1)", len(seen), steps)
	}
}

// The invariant two versions of litBeadIndex violated: two beads placed in ONE emission travel at
// the same world speed, so after the same ELAPSED time they must light the same bead index —
// whatever their edges' STEP COUNTS. Node 1's two edges differ in length (hence step count) —
// give them a different ratio to each other than a naive proportional guess, so a version
// that gets the per-edge scaling wrong fails.
//
// Driving this by elapsed time rather than by a chosen distance is the point: it is what the
// screen shows, and it is what caught t*centerDistance / t*chordLength, where a per-edge
// ratio reintroduced an offset that t*steps does not have (dwell is UNIFORM per step, so
// elapsed/dwell is the same integer progress for every edge regardless of its own step count).
func TestLitBeadIndexSameElapsedLightsSameBead(t *testing.T) {
	const longSteps, shortSteps = 32, 28

	for elapsed := 0.0; elapsed < 120; elapsed += 0.25 {
		coveredSteps := elapsed / lattice.DwellTicksPerBead // elapsed is in the SAME ticks unit as dwell
		gotLong, okLong := LitBeadIndex(coveredSteps/longSteps, longSteps)
		gotShort, okShort := LitBeadIndex(coveredSteps/shortSteps, shortSteps)
		if !okLong || !okShort {
			continue
		}
		if gotLong != gotShort {
			t.Fatalf("elapsed %.2f (covered %.2f steps): long edge lit bead %d, short edge lit bead %d — equal elapsed must light the same index",
				elapsed, coveredSteps, gotLong, gotShort)
		}
	}
}

// One bead index per dwell of travel — the constant dwell the design rests on. If this
// drifts, the lit bead is no longer moving at the uniform pulse speed.
func TestLitBeadIndexAdvancesOncePerStep(t *testing.T) {
	const steps = 25
	for i := 0; i < steps; i++ {
		got, ok := LitBeadIndex(float64(i)/steps, steps)
		if !ok {
			t.Fatalf("bead %d: t=%.4f reported off-chain", i, float64(i)/steps)
		}
		if got != i {
			t.Errorf("t=%.4f (bead %d's own position) lit index %d, want %d", float64(i)/steps, i, got, i)
		}
	}
}

// t outside [0,1) — before departure, or having arrived — lights nothing.
func TestLitBeadIndexOffChainLightsNothing(t *testing.T) {
	const steps = 25
	if _, ok := LitBeadIndex(-0.01, steps); ok {
		t.Error("t<0 (not yet departed) lit a bead; want nothing lit")
	}
	if _, ok := LitBeadIndex(1, steps); ok {
		t.Error("t=1 (arrived at the target) lit a bead; want nothing lit")
	}
}
