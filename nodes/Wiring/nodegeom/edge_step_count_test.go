package nodegeom

import (
	"testing"

	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// edge_step_count_test.go — EdgeStepCount's own formula, independent of chainBeads' node-
// mover plumbing: the live center-to-center distance minus both nodes' torus steps, clamped
// to a minimum of 1, and rounded for a near-integer distance.

// EdgeStepCount pins the formula (docs/bead-model/bead-lattice.md "The count") directly, independent of
// chainBeads' node-mover plumbing. It now takes the LIVE center-to-center distance directly
// (K = round(dist/BeadStepR)) rather than a stored LocalPolar, so a distance already an exact
// multiple of BeadStepR (as every test fixture here uses) is plain integer subtraction with
// nothing to round.
func TestEdgeStepCount(t *testing.T) {
	dist := 200 * lattice.BeadStepR
	got := EdgeStepCount(dist, "Input", "Time")
	want := 200 - NodeTorusSteps("Input") - NodeTorusSteps("Time")
	if got != want {
		t.Fatalf("EdgeStepCount = %d, want %d", got, want)
	}
	if want < 1 {
		t.Fatal("test fixture collapsed to <1 step; pick a larger separation")
	}
}

// A collapsed or negative gap clamps to a minimum of 1 bead — an edge is never zero-length.
func TestEdgeStepCountClampsToMinimumOne(t *testing.T) {
	dist := 1 * lattice.BeadStepR // 1 bead step of separation, far inside both tori
	if got := EdgeStepCount(dist, "Input", "Time"); got != 1 {
		t.Fatalf("EdgeStepCount(collapsed) = %d, want 1", got)
	}
}

// EdgeStepCount rounds a NEAR-integer distance to the nearest bead step rather than
// truncating or requiring exactness — the round() exists so a node mid-way through
// placement (a live distance a hair off an exact multiple, from float accumulation) never
// silently drops into the wrong bucket. A node whose live distance happens to land on the
// bead lattice never actually exercises the rounding in practice, but the function must
// still behave sanely on the inputs it can receive.
func TestEdgeStepCountRoundsNearIntegerDistance(t *testing.T) {
	exact := 50 * lattice.BeadStepR
	nudged := exact + 1e-6
	if got, want := EdgeStepCount(nudged, "Input", "Input"), EdgeStepCount(exact, "Input", "Input"); got != want {
		t.Fatalf("EdgeStepCount should round a near-integer distance the same as the exact one: got %d want %d", got, want)
	}
}
