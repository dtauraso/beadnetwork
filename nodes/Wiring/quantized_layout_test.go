package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"math"
	"testing"
)

// measureScalars/deriveCenters round-trip: a node's world center, re-measured about the
// scene center (flat absolute scene-polar — every node independent) and then re-derived,
// reproduces the original center. Scalars are chosen directly (exact integer cells) rather
// than measured from an arbitrary world center — measurement quantizes to the nearest
// cell, so round-tripping an off-grid center is lossy by design; the invariant under test
// is that measure(derive(scalars)) reproduces the SAME scalars (idempotent on-grid points).
func TestMeasureScalarsRoundTrips(t *testing.T) {
	sceneCenter := vec3{}
	scalars := map[string]quantizedOffset{
		"g": {},
		"p": {iTheta: 4, iPhi: 2, iR: 3},
		"c": {iTheta: 2, iPhi: -3, iR: 2},
	}
	derived := deriveCenters(scalars, sceneCenter)
	ids := map[string]bool{"g": true, "p": true, "c": true}
	remeasured := measureScalars(derived, ids, sceneCenter, scalars)
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
	// Offsets are exact multiples of stepR (wire.BeadStepR, now the scene lattice's own
	// radial cell) so the round-trip below is lossless — an off-grid offset would be
	// legitimately rounded to the nearest cell, which TestMeasureScalarsRoundTrips'
	// doc comment already covers; this test's point is independence from other nodes,
	// not quantization rounding.
	centers := map[string]vec3{
		"a": sceneCenter.Add(vec3{X: 5 * wire.BeadStepR, Y: 0, Z: 0}),
		"b": sceneCenter.Add(vec3{X: 0, Y: 0, Z: 8 * wire.BeadStepR}),
	}
	ids := map[string]bool{"a": true, "b": true}
	offs := measureScalars(centers, ids, sceneCenter, nil)
	if _, ok := offs["a"]; !ok {
		t.Fatal("expected an offset for a")
	}
	if _, ok := offs["b"]; !ok {
		t.Fatal("expected an offset for b")
	}

	// Re-derive: each node's center comes straight back from the scene center, with no
	// dependency on the other node's offset.
	derived := deriveCenters(offs, sceneCenter)
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
	stale := quantizedOffset{
		iTheta: 6, cTheta: oldStepTheta,
		iPhi: 12, cPhi: oldStepTheta,
		iR: 13, cR: oldStepR,
	}
	got := normalizeOffset(stale)

	// The STEP must land on the CURRENT lattice constants.
	if got.cTheta != stepTheta || got.cPhi != stepPhi || got.cR != stepR {
		t.Fatalf("normalizeOffset did not rewrite the step to the current lattice: got=%+v want steps=(%g,%g,%g)",
			got, stepTheta, stepPhi, stepR)
	}
	// The INDEX must have moved WITH the step — this is the failure mode PROVEN below by
	// simulating "step-only" normalization (what a naive fix would do: overwrite cR/cTheta
	// without touching iTheta/iPhi/iR) and showing the represented WORLD VALUE is wrong.
	wantTheta := 6.0 * oldStepTheta / stepTheta
	wantPhi := 12.0 * oldStepTheta / stepPhi
	wantR := 13.0 * oldStepR / stepR
	if math.Abs(float64(got.iTheta)-wantTheta) > 0.5 || math.Abs(float64(got.iPhi)-wantPhi) > 0.5 || math.Abs(float64(got.iR)-wantR) > 0.5 {
		t.Fatalf("normalizeOffset did not preserve the world distance: got=%+v want index near (%g,%g,%g)",
			got, wantTheta, wantPhi, wantR)
	}

	// Proof of failure: the represented world value (index * step) BEFORE and AFTER
	// normalization must be the SAME, up to the rounding a single ROUND(...) introduces
	// (at most half the NEW step — the same tolerance LoadLocalPolars' own conversion
	// accepts) — this is the invariant a "step-only" bug (index left at the OLD lattice's
	// count, step silently swapped to the new lattice's size) breaks by a much larger
	// margin. Simulate that exact bug here (not by reverting production code) and show it
	// fails this same tolerance, proving the check can catch it.
	beforeR := float64(stale.iR) * stale.cR
	afterR := float64(got.iR) * got.cR
	if math.Abs(beforeR-afterR) > stepR/2 {
		t.Fatalf("normalizeOffset changed the represented world R distance by more than half a step: before=%g after=%g", beforeR, afterR)
	}
	buggyStepOnly := stale
	buggyStepOnly.cR = stepR // step rewritten, index left untouched — the bug
	buggyAfterR := float64(buggyStepOnly.iR) * buggyStepOnly.cR
	if math.Abs(beforeR-buggyAfterR) <= stepR/2 {
		t.Fatalf("test setup invalid: the step-only bug should have changed the represented R distance by more than half a step, but didn't (before=%g buggy=%g)", beforeR, buggyAfterR)
	}
}

// TestCommitNodeMoveLocalDrawsQuantizedNotRawTarget: under the quantized scene lattice, a
// drag's COMMITTED position (what applyCenter draws, what gets persisted, what neighbors
// re-quantize against) must be the LATTICE POINT implied by measureScalar/offsetScenePolar,
// never the raw continuous drag target — the bug this whole change fixes
// (docs/which-lattice-a-node-lives-on.md "Why the drag makes it worst": the node used to
// glide continuously while its own chain beads moved in bead-distance jumps). Proof of
// failure: the raw target chosen below is deliberately OFF the lattice (not an exact
// multiple of stepR/stepTheta/stepPhi), so "commit the raw target" and "commit the
// quantized point" give two different, distinguishable answers — asserting against the
// wrong one (the raw target) is shown to fail.
func TestCommitNodeMoveLocalDrawsQuantizedNotRawTarget(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	if !md.lq.quantizedLayout {
		t.Fatal("test assumes quantizedLayout is on by default")
	}
	nm, ok := md.mr.nodeMovers["dst"]
	if !ok {
		t.Fatal("no nodeMover for dst")
	}
	before, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst")
	}
	srcCenter, ok := md.centerOfNode("src")
	if !ok {
		t.Fatal("no center for src")
	}
	// The angle gate (bead_crud.go, PLAN.md) only admits an ADD when the drag heads
	// AWAY from the touching bead's source, not merely far from "before" in some
	// arbitrary direction — so the target is placed further out along dst's own live
	// bearing away from its one neighbour, src, moved by a distance deliberately off the
	// lattice (stepR is 8.96, so +30 world units is not an exact multiple of it).
	outward := before.Sub(srcCenter).Normalize()
	target := before.Add(outward.Scale(30))

	// Computed BEFORE the commit — quantizedDragTarget reads the node's (and its
	// neighbours') CURRENT centers as the solve's starting configuration, so calling it
	// after commitNodeMoveLocal has already moved dst would race its own answer (see
	// quantizedDragTarget's doc comment in subtree_persist_test.go).
	want := quantizedDragTarget(md, "dst", target)

	md.lq.commitNodeMoveLocal(md, nm, target)

	got, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst after commit")
	}
	// (1) The committed center must NOT be the raw target — proves the fix is active
	// (reverting it back to `nm.applyCenter(newPos, ...)` with the raw target would make
	// this assertion fail, since `got` would then equal `target` exactly).
	if d := got.Sub(target).Length(); d < 1e-6 {
		t.Fatalf("commitNodeMoveLocal drew the RAW target instead of the quantized lattice point: got=%+v raw-target=%+v", got, target)
	}
	// (2) The committed center must equal the bead-CRUD oracle (bead_crud.go, PLAN.md
	// "moving a node is CRUD on the edge beads touching it") — the positive half of the
	// same assertion, pinning WHAT it should be, not just what it shouldn't. This
	// replaced an earlier version that compared against a global bead-cell solver
	// (rejected, PLAN.md "Why the previous attempts were wrong"); quantizedDragTarget
	// (subtree_persist_test.go) is the shared oracle for the replacement (`want` computed
	// above, BEFORE the commit).
	if d := got.Sub(want).Length(); d > 1e-6 {
		t.Fatalf("commitNodeMoveLocal's committed center does not match the bead-crud oracle: got=%+v want=%+v", got, want)
	}
	// (3) Consistency (PLAN.md): the node always moves a bead's distance, never further —
	// the committed center is never more than one wire.BeadStepR away from where it
	// started, and (since this drag target is deliberately off-lattice and far from
	// "dst"'s pre-drag position) it must have moved SOME distance rather than holding.
	moved := got.Sub(before).Length()
	if moved < 1e-9 {
		t.Fatalf("dst never moved on a drag whose target is far off its pre-drag position: before=%+v got=%+v", before, got)
	}
	if moved > wire.BeadStepR+1e-6 {
		t.Fatalf("dst moved more than one bead length in a single commit: moved=%v beadStepR=%v", moved, wire.BeadStepR)
	}
}
