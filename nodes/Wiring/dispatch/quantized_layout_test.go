package dispatch

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/beadcrud"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/layoutquant"
	"github.com/dtauraso/wirefold/nodes/Wiring/positionfile"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// TestCommitNodeMoveLocalDrawsQuantizedNotRawTarget: under the quantized scene lattice, a
// drag's COMMITTED position (what applyCenter draws, what gets persisted, what neighbors
// re-quantize against) must be the LATTICE POINT implied by measureScalar/offsetScenePolar,
// never the raw continuous drag target — the bug this whole change fixes
// (docs/investigations/which-lattice-a-node-lives-on.md "Why the drag makes it worst": the node used to
// glide continuously while its own chain beads moved in bead-distance jumps). Proof of
// failure: the raw target chosen below is deliberately OFF the lattice (not an exact
// multiple of stepR/stepTheta/stepPhi), so "commit the raw target" and "commit the
// quantized point" give two different, distinguishable answers — asserting against the
// wrong one (the raw target) is shown to fail.
func TestCommitNodeMoveLocalDrawsQuantizedNotRawTarget(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	if !md.lq.QuantizedLayout {
		t.Fatal("test assumes quantizedLayout is on by default")
	}
	nm, ok := md.mr.nodeGeoms["2"]
	if !ok {
		t.Fatal("no nodeMover for dst")
	}
	before, ok := md.mr.centerOfNode("2")
	if !ok {
		t.Fatal("no center for dst")
	}
	srcCenter, ok := md.mr.centerOfNode("1")
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
	want := quantizedDragTarget(md, "2", target)

	md.lq.CommitNodeMoveLocal(md.mr.nodeGeoms, md.mr.edgeMovers, &md.UI, nm, target)

	got, ok := md.mr.centerOfNode("2")
	if !ok {
		t.Fatal("no center for dst after commit")
	}
	// (1) The committed center must NOT be the raw target — proves the fix is active
	// (reverting it back to `nm.ApplyCenter(newPos, ...)` with the raw target would make
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
	// (3) Consistency (PLAN.md): the node moves because of ONE bead operation (add or
	// remove) on ONE touching bead, never a value derived from the raw drag's own
	// magnitude — so (since this drag target is deliberately off-lattice and far from
	// "2"'s pre-drag position) it must have moved SOME distance rather than holding.
	// The Cartesian SIZE of that move is whatever the winning verdict's own geometry
	// implies (beadCrudImpliedCentre) — a REMOVE lands exactly on the removed bead's own
	// centre and an ADD one bead length beyond the newly added bead, along that edge's
	// own axis; neither is pinned to lattice.BeadStepR itself (a REMOVE's distance from
	// prevPos is nodeTorusOuterR+lattice.BeadTorusOuterR, which can exceed BeadStepR for a
	// node kind with a large torus). TestCommitNodeMoveLocalRemoveTakesBeadsPlace and
	// TestCommitNodeMoveLocalAddMovesOneBeadBeyondNewBead pin the exact magnitude/axis for
	// each verdict directly.
	moved := got.Sub(before).Length()
	if moved < 1e-9 {
		t.Fatalf("dst never moved on a drag whose target is far off its pre-drag position: before=%+v got=%+v", before, got)
	}
}

// TestCommitNodeMoveLocalNeverMovesTowardMouseTarget is the test PLAN.md required and the
// prior build's test suite did not catch: "A test must fail if the node's centre is ever
// set from the mouse target." It does NOT lean on quantizedDragTarget (the shared oracle
// commitNodeMoveLocal and quantizedDragTarget both call through resolveBeadCrudMove) —
// that would only prove production agrees with the oracle, not that the oracle itself is
// right. Instead it hand-computes the WRONG answer directly (the deleted-then-rebuilt
// walkBeadPath formula: prevPos moved one lattice.BeadStepR toward the raw target) and
// asserts the real commit does NOT match it, for both a REMOVE-triggering drag and an
// ADD-triggering drag on the same single-neighbour fixture.
func TestCommitNodeMoveLocalNeverMovesTowardMouseTarget(t *testing.T) {
	cursorFollow := func(prevPos, target vec3) vec3 {
		delta := target.Sub(prevPos)
		if delta.Length() < 1e-9 {
			return prevPos
		}
		return prevPos.Add(delta.Normalize().Scale(lattice.BeadStepR))
	}

	t.Run("remove", func(t *testing.T) {
		root := writeTree(t)
		md := loadTreeMD(t, root)
		nm := md.mr.nodeGeoms["2"]
		before, ok := md.mr.centerOfNode("2")
		if !ok {
			t.Fatal("no center for dst")
		}
		beads := layoutquant.DragTouchingBeads(md.mr.edgeMovers, nm, before)
		if len(beads) == 0 {
			t.Fatal("dst has no touching beads to judge")
		}
		// Land exactly on the touching bead's own SOURCE point: |third| == 0, well under
		// one bead length, so its verdict is beadCrudRemove.
		target := beads[0].Source
		wrong := cursorFollow(before, target)

		md.lq.CommitNodeMoveLocal(md.mr.nodeGeoms, md.mr.edgeMovers, &md.UI, nm, target)
		got, ok := md.mr.centerOfNode("2")
		if !ok {
			t.Fatal("no center for dst after commit")
		}
		if d := got.Sub(wrong).Length(); d < 1e-6 {
			t.Fatalf("commitNodeMoveLocal moved toward the mouse target (the old walkBeadPath formula), not from the bead operation: got=%+v cursor-follow=%+v", got, wrong)
		}
	})

	t.Run("add", func(t *testing.T) {
		root := writeTree(t)
		md := loadTreeMD(t, root)
		nm := md.mr.nodeGeoms["2"]
		before, ok := md.mr.centerOfNode("2")
		if !ok {
			t.Fatal("no center for dst")
		}
		srcCenter, ok := md.mr.centerOfNode("1")
		if !ok {
			t.Fatal("no center for src")
		}
		outward := before.Sub(srcCenter).Normalize()
		// Far enough outward, aligned with the touching bead's own axis, that the angle
		// gate admits an ADD.
		target := before.Add(outward.Scale(40))
		wrong := cursorFollow(before, target)

		md.lq.CommitNodeMoveLocal(md.mr.nodeGeoms, md.mr.edgeMovers, &md.UI, nm, target)
		got, ok := md.mr.centerOfNode("2")
		if !ok {
			t.Fatal("no center for dst after commit")
		}
		if d := got.Sub(wrong).Length(); d < 1e-6 {
			t.Fatalf("commitNodeMoveLocal moved toward the mouse target (the old walkBeadPath formula), not from the bead operation: got=%+v cursor-follow=%+v", got, wrong)
		}
	})
}

// TestCommitNodeMoveLocalRemoveTakesBeadsPlace pins PLAN.md's REMOVE rule positively: "bead
// removed -> the node moves to take that bead's place." The node's new centre must equal
// the removed bead's own former centre EXACTLY — not a value derived from the drag target,
// not one bead length, not any other distance.
func TestCommitNodeMoveLocalRemoveTakesBeadsPlace(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	nm := md.mr.nodeGeoms["2"]
	before, ok := md.mr.centerOfNode("2")
	if !ok {
		t.Fatal("no center for dst")
	}
	beads := layoutquant.DragTouchingBeads(md.mr.edgeMovers, nm, before)
	if len(beads) != 1 {
		t.Fatalf("fixture assumption: dst has exactly one touching bead, got %d", len(beads))
	}
	removedBeadCentre := beads[0].Centre
	target := beads[0].Source // |third| == 0 < one bead length -> beadCrudRemove

	verdict, _ := beadcrud.BeadCrudDecide(beads[0].Source, beads[0].Centre, target, target.Sub(before), lattice.BeadStepR)
	if verdict != beadcrud.BeadCrudRemove {
		t.Fatalf("fixture assumption: this drag should verdict beadCrudRemove, got %v", verdict)
	}

	md.lq.CommitNodeMoveLocal(md.mr.nodeGeoms, md.mr.edgeMovers, &md.UI, nm, target)
	got, ok := md.mr.centerOfNode("2")
	if !ok {
		t.Fatal("no center for dst after commit")
	}
	if d := got.Sub(removedBeadCentre).Length(); d > 1e-6 {
		t.Fatalf("dst's new centre should be exactly the removed bead's former centre: got=%+v want=%+v (off by %g)", got, removedBeadCentre, d)
	}
}

// TestCommitNodeMoveLocalAddMovesOneBeadBeyondNewBead pins PLAN.md's ADD rule positively:
// a bead is added at the next chain position, and the node's new centre is one bead length
// beyond it, along the chain's own axis — never toward the raw drag target.
func TestCommitNodeMoveLocalAddMovesOneBeadBeyondNewBead(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	nm := md.mr.nodeGeoms["2"]
	before, ok := md.mr.centerOfNode("2")
	if !ok {
		t.Fatal("no center for dst")
	}
	srcCenter, ok := md.mr.centerOfNode("1")
	if !ok {
		t.Fatal("no center for src")
	}
	beads := layoutquant.DragTouchingBeads(md.mr.edgeMovers, nm, before)
	if len(beads) != 1 {
		t.Fatalf("fixture assumption: dst has exactly one touching bead, got %d", len(beads))
	}
	outward := before.Sub(srcCenter).Normalize()
	target := before.Add(outward.Scale(40))

	dragVector := target.Sub(before)
	verdict, _ := beadcrud.BeadCrudDecide(beads[0].Source, beads[0].Centre, target, dragVector, lattice.BeadStepR)
	if verdict != beadcrud.BeadCrudAdd {
		t.Fatalf("fixture assumption: this drag should verdict beadCrudAdd, got %v", verdict)
	}
	// Hand-computed expected centre, independent of beadCrudImpliedCentre's own
	// implementation: the new bead sits one bead length CLOSER to the node than the old
	// touching bead (along the chain axis), and the node's new centre is one bead length
	// further BEYOND that new bead, away from the neighbour.
	newBeadCentre := beads[0].Centre.Sub(beads[0].AimDir.Scale(lattice.BeadStepR))
	wantNodeCentre := newBeadCentre.Sub(beads[0].AimDir.Scale(lattice.BeadStepR))

	md.lq.CommitNodeMoveLocal(md.mr.nodeGeoms, md.mr.edgeMovers, &md.UI, nm, target)
	got, ok := md.mr.centerOfNode("2")
	if !ok {
		t.Fatal("no center for dst after commit")
	}
	if d := got.Sub(wantNodeCentre).Length(); d > 1e-6 {
		t.Fatalf("dst's new centre should be one bead length beyond the newly added bead, along the chain axis: got=%+v want=%+v (off by %g)", got, wantNodeCentre, d)
	}
}

// TestCommitNodeMoveLocalPersistsQuantizedNotRawPolar closes a coverage hole the tests
// above miss entirely: they all assert `ApplyCenter`'s DRAWN center
// (md.mr.centerOfNode), never the scene-polar written to nodes/<id>/position.json by
// persistQuantOffset (commit_node_move.go, quant_offset_persist.go). A prior injected bug
// — computing committedPolar from newPos (the raw drag target) instead of committedPos
// (the quantized lattice landing point) — passed every existing node-drag test, because
// none of them read the persisted file; it only surfaces on reload
// (.claude/rules/persistence-ownership.md, docs/process/testing-shape.md's persistence
// exception is the one case a test may go through real disk bytes). This test drives a
// real commitNodeMoveLocal with EnableEditPersist armed (same setup as
// TestCommitNodeMoveLocalDrawsQuantizedNotRawTarget: an off-lattice drag target so the raw
// target and the quantized landing point are distinguishable) and reads position.json back
// off a t.TempDir() tree, asserting the persisted scene-polar corresponds to committedPos
// (what got drawn/committed), never to the raw drag target's polar.
func TestCommitNodeMoveLocalPersistsQuantizedNotRawPolar(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	md.EnableEditPersist(root)
	if !md.lq.QuantizedLayout {
		t.Fatal("test assumes quantizedLayout is on by default")
	}
	nm, ok := md.mr.nodeGeoms["2"]
	if !ok {
		t.Fatal("no nodeMover for dst")
	}
	before, ok := md.mr.centerOfNode("2")
	if !ok {
		t.Fatal("no center for dst")
	}
	srcCenter, ok := md.mr.centerOfNode("1")
	if !ok {
		t.Fatal("no center for src")
	}
	outward := before.Sub(srcCenter).Normalize()
	// Off-lattice, same as TestCommitNodeMoveLocalDrawsQuantizedNotRawTarget: stepR is
	// 8.96, so +30 world units is not an exact multiple of it — the raw target's own
	// scene-polar and the committed (quantized) center's scene-polar are distinguishable.
	target := before.Add(outward.Scale(30))

	md.lq.CommitNodeMoveLocal(md.mr.nodeGeoms, md.mr.edgeMovers, &md.UI, nm, target)

	got, ok := md.mr.centerOfNode("2")
	if !ok {
		t.Fatal("no center for dst after commit")
	}
	wantPolar := geom.Cart2polar(got.Sub(md.UI.SceneSphere.Center))
	rawTargetPolar := geom.Cart2polar(target.Sub(md.UI.SceneSphere.Center))

	// Assert real bytes on disk, per docs/process/testing-shape.md's persistence
	// exception — persistQuantOffset is synchronous on the calling (this test's own)
	// goroutine, so no wait/barrier is needed before reading the file back.
	raw, err := os.ReadFile(positionfile.FilePath(root, "2"))
	if err != nil {
		t.Fatalf("reading position.json: %v", err)
	}
	var pf positionfile.JSON
	if err := json.Unmarshal(raw, &pf); err != nil {
		t.Fatalf("unmarshal position.json: %v", err)
	}

	// (1) The persisted scene-polar must NOT be the raw drag target's polar — proves the
	// injected bug (committedPolar := geom.Cart2polar(newPos.Sub(...)) in the quantized
	// branch) is caught.
	if closePolar(pf.ScenePolarR, pf.ScenePolarTheta, pf.ScenePolarPhi,
		rawTargetPolar.R, rawTargetPolar.Theta, rawTargetPolar.Phi) {
		t.Fatalf("position.json persisted the RAW drag target's polar instead of the quantized committed polar: "+
			"persisted=(%g,%g,%g) raw-target-polar=(%g,%g,%g)",
			pf.ScenePolarR, pf.ScenePolarTheta, pf.ScenePolarPhi,
			rawTargetPolar.R, rawTargetPolar.Theta, rawTargetPolar.Phi)
	}
	// (2) The persisted scene-polar must equal committedPos's own polar (what was
	// actually drawn/committed) — the positive half of the assertion.
	if !closePolar(pf.ScenePolarR, pf.ScenePolarTheta, pf.ScenePolarPhi,
		wantPolar.R, wantPolar.Theta, wantPolar.Phi) {
		t.Fatalf("position.json's persisted polar does not match the committed center's own polar: "+
			"persisted=(%g,%g,%g) want=(%g,%g,%g)",
			pf.ScenePolarR, pf.ScenePolarTheta, pf.ScenePolarPhi,
			wantPolar.R, wantPolar.Theta, wantPolar.Phi)
	}
}

func closePolar(r1, t1, p1, r2, t2, p2 float64) bool {
	const eps = 1e-6
	d := func(a, b float64) float64 {
		x := a - b
		if x < 0 {
			x = -x
		}
		return x
	}
	return d(r1, r2) < eps && d(t1, t2) < eps && d(p1, p2) < eps
}
