package quantoffset

import (
	"math"
	"testing"

	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// measureScalars/deriveCenters round-trip: a node's world center, re-measured about the
// scene center (flat absolute scene-polar — every node independent) and then re-derived,
// reproduces the original center. Scalars are chosen directly (exact integer cells) rather
// than measured from an arbitrary world center — measurement quantizes to the nearest
// cell, so round-tripping an off-grid center is lossy by design; the invariant under test
// is that measure(derive(scalars)) reproduces the SAME scalars (idempotent on-grid points).
func TestMeasureScalarsRoundTrips(t *testing.T) {
	sceneCenter := vec3{}
	scalars := map[string]QuantizedOffset{
		"g": {},
		"p": {ITheta: 4, IPhi: 2, IR: 3},
		"c": {ITheta: 2, IPhi: -3, IR: 2},
	}
	derived := DeriveCenters(scalars, sceneCenter)
	ids := map[string]bool{"g": true, "p": true, "c": true}
	remeasured := MeasureScalars(derived, ids, sceneCenter, scalars)
	for _, id := range []string{"g", "p", "c"} {
		if remeasured[id] != scalars[id] {
			t.Fatalf("%s: round-trip mismatch remeasured=%+v want=%+v", id, remeasured[id], scalars[id])
		}
	}
}

// TestMeasureScalarsMeasuresEveryNodeAboutSceneCenter asserts every node is measured
// independently about the ONE scene center — there is no reference/parent origin, so a
// node's offset never depends on another node's position.
func TestMeasureScalarsMeasuresEveryNodeAboutSceneCenter(t *testing.T) {
	sceneCenter := vec3{X: 10, Y: 20, Z: 30}
	// Offsets are exact multiples of stepR (lattice.BeadStepR, now the scene lattice's own
	// radial cell) so the round-trip below is lossless — an off-grid offset would be
	// legitimately rounded to the nearest cell, which TestMeasureScalarsRoundTrips'
	// doc comment already covers; this test's point is independence from other nodes,
	// not quantization rounding.
	centers := map[string]vec3{
		"a": sceneCenter.Add(vec3{X: 5 * lattice.BeadStepR, Y: 0, Z: 0}),
		"b": sceneCenter.Add(vec3{X: 0, Y: 0, Z: 8 * lattice.BeadStepR}),
	}
	ids := map[string]bool{"a": true, "b": true}
	offs := MeasureScalars(centers, ids, sceneCenter, nil)
	if _, ok := offs["a"]; !ok {
		t.Fatal("expected an offset for a")
	}
	if _, ok := offs["b"]; !ok {
		t.Fatal("expected an offset for b")
	}

	// Re-derive: each node's center comes straight back from the scene center, with no
	// dependency on the other node's offset.
	derived := DeriveCenters(offs, sceneCenter)
	if d := derived["a"].Sub(centers["a"]).Length(); d > 1e-6 {
		t.Fatalf("a: derived center drifted by %v", d)
	}
	if d := derived["b"].Sub(centers["b"]).Length(); d > 1e-6 {
		t.Fatalf("b: derived center drifted by %v", d)
	}
}

// TestNormalizeOffsetConvertsIndexAndStepTogether: a quantized offset loaded with a STALE
// per-axis step (the old scene lattice, 15°/20.0, before this collapsed onto the bead
// lattice) must have its index AND step converted TOGETHER, preserving the world
// distance/angle the pair represents — exactly the contract layout_holder.go's
// LoadLocalPolars documents and this mirrors. Proof of failure: reverting to "just
// overwrite the step, leave the index" (a bug this project has hit before, per
// LoadLocalPolars' doc comment: "a stale 2.0 became 8.96 and every separation jumped
// 4.5x") is exercised below by asserting the index actually changes, not just the step.
func TestNormalizeOffsetConvertsIndexAndStepTogether(t *testing.T) {
	const oldStepTheta = 0.26179938779914946 // π/12, the old 15° scene cell
	const oldStepR = 20.0                    // the old hand-picked scene radial cell
	stale := QuantizedOffset{
		ITheta: 6, CTheta: oldStepTheta,
		IPhi: 12, CPhi: oldStepTheta,
		IR: 13, CR: oldStepR,
	}
	got := NormalizeOffset(stale)

	// The STEP must land on the CURRENT lattice constants.
	if got.CTheta != stepTheta || got.CPhi != stepPhi || got.CR != stepR {
		t.Fatalf("NormalizeOffset did not rewrite the step to the current lattice: got=%+v want steps=(%g,%g,%g)",
			got, stepTheta, stepPhi, stepR)
	}
	// The INDEX must have moved WITH the step — this is the failure mode PROVEN below by
	// simulating "step-only" normalization (what a naive fix would do: overwrite cR/cTheta
	// without touching iTheta/iPhi/iR) and showing the represented WORLD VALUE is wrong.
	wantTheta := 6.0 * oldStepTheta / stepTheta
	wantPhi := 12.0 * oldStepTheta / stepPhi
	wantR := 13.0 * oldStepR / stepR
	if math.Abs(float64(got.ITheta)-wantTheta) > 0.5 || math.Abs(float64(got.IPhi)-wantPhi) > 0.5 || math.Abs(float64(got.IR)-wantR) > 0.5 {
		t.Fatalf("NormalizeOffset did not preserve the world distance: got=%+v want index near (%g,%g,%g)",
			got, wantTheta, wantPhi, wantR)
	}

	// Proof of failure: the represented world value (index * step) BEFORE and AFTER
	// normalization must be the SAME, up to the rounding a single ROUND(...) introduces
	// (at most half the NEW step — the same tolerance LoadLocalPolars' own conversion
	// accepts) — this is the invariant a "step-only" bug (index left at the OLD lattice's
	// count, step silently swapped to the new lattice's size) breaks by a much larger
	// margin. Simulate that exact bug here (not by reverting production code) and show it
	// fails this same tolerance, proving the check can catch it.
	beforeR := float64(stale.IR) * stale.CR
	afterR := float64(got.IR) * got.CR
	if math.Abs(beforeR-afterR) > stepR/2 {
		t.Fatalf("NormalizeOffset changed the represented world R distance by more than half a step: before=%g after=%g", beforeR, afterR)
	}
	buggyStepOnly := stale
	buggyStepOnly.CR = stepR // step rewritten, index left untouched — the bug
	buggyAfterR := float64(buggyStepOnly.IR) * buggyStepOnly.CR
	if math.Abs(beforeR-buggyAfterR) <= stepR/2 {
		t.Fatalf("test setup invalid: the step-only bug should have changed the represented R distance by more than half a step, but didn't (before=%g buggy=%g)", beforeR, buggyAfterR)
	}
}
