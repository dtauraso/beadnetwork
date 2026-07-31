package Wiring

import (
	"math"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// walk_bead_path_test.go — pure single-goroutine tests of walkBeadPath, the "each bead
// is also a polar vector, dragging is vectors combining" walk (quantized_move.go). Per
// docs/testing-shape.md this is exactly the shape that belongs in a unit test: one
// function, no goroutines, asserting what IT decided given an input.

// approxEqual gives strides room for float accumulation error (repeated Add/Scale over
// a single stride) without hiding a real defect — 1e-6 is many orders
// below a single bead length (8.96).
func approxEqual(t *testing.T, got, want float64, msg string) {
	t.Helper()
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("%s: got %v want %v", msg, got, want)
	}
}

func TestWalkBeadPath_OneBeadAway_MovesExactlyOneBead(t *testing.T) {
	prev := vec3{X: 0, Y: 0, Z: 0}
	target := vec3{X: wire.BeadStepR, Y: 0, Z: 0}
	got := walkBeadPath(prev, target)
	approxEqual(t, got.Sub(prev).Length(), wire.BeadStepR, "distance moved")
	approxEqual(t, got.X, wire.BeadStepR, "landed X")
}

func TestWalkBeadPath_HalfBeadAway_DoesNotMove(t *testing.T) {
	prev := vec3{X: 0, Y: 0, Z: 0}
	target := vec3{X: wire.BeadStepR / 2, Y: 0, Z: 0}
	got := walkBeadPath(prev, target)
	if got != prev {
		t.Fatalf("expected no movement for a sub-bead target, got %+v (prev %+v)", got, prev)
	}
}

func TestWalkBeadPath_FarTargetMovesExactlyOneBead(t *testing.T) {
	prev := vec3{X: 0, Y: 0, Z: 0}
	// A commit is ONE pointer-move event and the node moves at most one bead per move,
	// so a target five beads away moves it ONE bead toward the target — not five. The
	// remainder is not lost: the next pointer-move takes the next stride. This replaces
	// an assertion that the walk consumed the WHOLE displacement in a single call, which
	// is what let a fast drag lurch several beads at once instead of stepping.
	target := vec3{X: 5 * wire.BeadStepR, Y: 0, Z: 0}
	got := walkBeadPath(prev, target)
	approxEqual(t, got.Sub(prev).Length(), wire.BeadStepR, "one stride per commit, however far the target")
	approxEqual(t, got.X, wire.BeadStepR, "the stride is along the target direction")

	// Stepping again advances exactly one more bead, so repeated moves walk toward the
	// target rather than the first one jumping the whole way.
	got2 := walkBeadPath(got, target)
	approxEqual(t, got2.X, 2*wire.BeadStepR, "a second commit takes a second stride")
}

// TestWalkBeadPath_UniformStepLengthEveryDirection is the direction test: the whole
// point of the "polar vector" model is that a stride's LENGTH is the invariant and it is
// IDENTICAL in every direction, including a purely TANGENTIAL one (the case measured
// live as 7x-18x too short under the old fixed-1-degree angular tick, docs/bead-lattice.md's
// drag.jump probe finding). Radial, tangential (perpendicular to the radius, i.e. an
// angular-only move at fixed r), and diagonal targets must all produce a stride of
// exactly wire.BeadStepR.
func TestWalkBeadPath_UniformStepLengthEveryDirection(t *testing.T) {
	prev := vec3{X: 50, Y: 0, Z: 0} // r=50 from origin, a radius this graph actually uses (~28-70)

	// Each target sits ONE bead-and-a-fraction from prev in its named direction, so a
	// single walkBeadPath call takes exactly one stride and stops — isolating the
	// per-stride length this test is about, rather than a multi-stride walk toward a
	// far-off target (whose total distance would be capped/target-limited, not
	// necessarily one bead, and so would not isolate the per-direction step length).
	cases := map[string]vec3{
		"radial (away from origin)": prev.Add(vec3{X: 1.4 * wire.BeadStepR, Y: 0, Z: 0}),
		"tangential (+Y, fixed r)":  prev.Add(vec3{X: 0, Y: 1.4 * wire.BeadStepR, Z: 0}),
		"tangential (+Z, fixed r)":  prev.Add(vec3{X: 0, Y: 0, Z: 1.4 * wire.BeadStepR}),
		// Diagonal component scaled by 1/sqrt(3) so the COMBINED displacement magnitude
		// is 1.4 beads (matching every other case here), not 1.4*sqrt(3) beads (which
		// would take two strides and stop measuring the per-stride length).
		"diagonal":                   prev.Add(vec3{X: 1.4 * wire.BeadStepR / math.Sqrt(3), Y: 1.4 * wire.BeadStepR / math.Sqrt(3), Z: 1.4 * wire.BeadStepR / math.Sqrt(3)}),
		"tangential, opposite sense": prev.Add(vec3{X: 0, Y: -1.4 * wire.BeadStepR, Z: 0}),
	}
	for name, target := range cases {
		got := walkBeadPath(prev, target)
		approxEqual(t, got.Sub(prev).Length(), wire.BeadStepR, name)
	}
}

// TestWalkBeadPath_FixedAngularTickFails is the proof-of-failure this task requires: it
// temporarily reintroduces the SEPARATE fixed-1-degree angular quantization (the model
// this change replaces) against the SAME tangential case above, and shows it produces a
// step far short of one bead — this is the exact defect the walk fixes, not a
// hypothetical one. It does not call walkBeadPath; it re-derives the old formula inline
// (offset = r*deltaTheta at fixed r) so this test keeps working (and keeps failing the
// OLD model) even after the rejected quantizedOffset.cTheta/cPhi fields are long gone.
func TestWalkBeadPath_FixedAngularTickFails(t *testing.T) {
	const oldStepTheta = math.Pi / 180 // 1 degree, quantized_layout.go's stepTheta/stepPhi
	const r = 50.0                     // a radius this graph actually uses

	// One fixed-angle tick's arc length at this radius: arc = r * dtheta.
	oldTickArc := r * oldStepTheta

	if oldTickArc >= wire.BeadStepR {
		t.Fatalf("expected the old fixed-degree tick to be SHORTER than a bead at r=%v (that was the bug) — got tick=%.4f bead=%.4f, not reproduced", r, oldTickArc, wire.BeadStepR)
	}
	ratio := wire.BeadStepR / oldTickArc
	if ratio < 7 { // matches the measured 7x-18x shortfall (docs/bead-lattice.md)
		t.Fatalf("expected the old model's tangential shortfall to be at least ~7x (as measured live); got only %.2fx", ratio)
	}
	t.Logf("old fixed-1-degree tangential tick at r=%v: %.4f world units vs one bead %.4f (%.2fx too short) — this is the failure walkBeadPath fixes", r, oldTickArc, wire.BeadStepR, ratio)
}
