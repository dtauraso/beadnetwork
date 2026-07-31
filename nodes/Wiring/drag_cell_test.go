package Wiring

// drag_cell_test.go — pure single-goroutine tests of the CELL model that replaced
// walkBeadPath (quantized_move.go's dragCandidateCells/chooseDragCandidate): "each end
// bead from each neighbor should have a range of polar motion where possible drag points
// a node can be dragged to like cells" (the task's own model, docs/bead-lattice.md). Per
// docs/testing-shape.md, most of these assert what ONE goroutine (the dragged node's own
// mover, reached directly here with no Start()) decided given its own stored state —
// dragCandidateCells/chooseDragCandidate read only nm's own LayoutHolder/partnerCenters,
// both already seeded at load time (node_move.go's post-build seed loop), so no mover
// goroutine needs to be running to exercise them. The one exception
// (TestDragExactTangencyAtTargetTorus) drives real goroutines and waits on a breadcrumb,
// exactly like this package's other neighbor-requantize assertions (subtree_persist_test.go)
// — that one test is about a NEIGHBOR's own requantize, which only that neighbor's own
// goroutine can produce.

import (
	"context"
	"math"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// TestDragExactTangencyAtTargetTorus is the user's actual requirement: after a committed
// drag, the SOURCE's chain (src -> dst) tangent to dst's torus EXACTLY — for a drag target
// whose raw displacement is deliberately NOT a whole bead — because the committed position
// is one of dragCandidateCells' exact outputs (an integer index times a step constant),
// never a measured-then-rounded distance.
func TestDragExactTangencyAtTargetTorus(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := md.Start(ctx)
	t.Cleanup(func() { cancel(); wg.Wait() })

	nm, ok := md.mr.nodeMovers["dst"]
	if !ok {
		t.Fatal("no nodeMover for dst")
	}
	before, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst")
	}
	// Deliberately off-lattice: 13.7/-4.3/9.1 combine to a displacement that is not a
	// whole multiple of wire.BeadStepR (8.96) in any axis.
	target := before.Add(vec3{X: 13.7, Y: -4.3, Z: 9.1})
	if d := target.Sub(before).Length(); math.Mod(d, wire.BeadStepR) < 1e-6 || math.Mod(d, wire.BeadStepR) > wire.BeadStepR-1e-6 {
		t.Fatalf("test setup bug: raw displacement %v happens to be a whole bead multiple of %v", d, wire.BeadStepR)
	}
	want, wantOK := md.lq.chooseDragCandidate(nm, target)
	if !wantOK {
		t.Fatal("chooseDragCandidate found no candidate cells for dst")
	}

	var dbg syncBuffer
	md.tr.SetSink(&dbg)
	if !md.RootMove("dst", target) {
		t.Fatal("RootMove(dst) returned false")
	}
	pollDragConverged(t, md, "dst", want)
	// src (the neighbor that OWNS the outgoing chain to dst) re-quantizes its own local
	// polar to dst on ITS OWN goroutine (neighborSetCRequantize) — wait for its
	// "abc-drag" breadcrumb before reading, exactly like subtree_persist_test.go.
	waitForAbcDrag(t, &dbg, "src")

	lhSrc, ok := md.lq.layoutHolders["src"]
	if !ok {
		t.Fatal("no LayoutHolder for src")
	}
	var lp wire.LocalPolar
	found := false
	for _, cand := range lhSrc.LocalPolarsSnapshot() {
		if cand.To == "dst" {
			lp, found = cand, true
		}
	}
	if !found {
		t.Fatal("src has no local polar entry for dst after the drag")
	}

	srcCenter, ok := md.centerOfNode("src")
	if !ok {
		t.Fatal("no center for src")
	}
	dstCenter, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst")
	}
	liveDist := dstCenter.Sub(srcCenter).Length()

	count := edgeStepCount(lp, "SrcNode", "SinkNode")
	chainEndFarEdge := nodeTorusOuterR("SrcNode") + float64(count)*wire.BeadStepR
	targetSurfaceDist := liveDist - nodeTorusOuterR("SinkNode")
	if diff := math.Abs(chainEndFarEdge - targetSurfaceDist); diff > 1e-6 {
		t.Fatalf("chain end (%.6f) does not exactly tangent dst's torus surface (%.6f): diff=%.6f — the committed position is not landing exactly on an integer bead cell", chainEndFarEdge, targetSurfaceDist, diff)
	}
}

// TestDragExactTangencyAtTargetTorus_FailsWithUnroundedDistance is the PROOF that the test
// above actually catches the defect it claims to: it recomputes the SAME chain-end-vs-
// target-surface comparison but deliberately measures the SOURCE's own QuantIR by rounding
// a fresh live distance the way the pre-cell-model code used to (round(freshRadius/rStep) —
// see requantizePoleTraced's doc comment on why that's now confined to a NEW/never-quantized
// entry) instead of trusting an already-exact, cell-committed one, and shows the resulting
// residue is close to the ~0.31 world-unit gap this task's own probe measured live.
func TestDragExactTangencyAtTargetTorus_FailsWithUnroundedDistance(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	lhSrc, ok := md.lq.layoutHolders["src"]
	if !ok {
		t.Fatal("no LayoutHolder for src")
	}
	var lp wire.LocalPolar
	found := false
	for _, cand := range lhSrc.LocalPolarsSnapshot() {
		if cand.To == "dst" {
			lp, found = cand, true
		}
	}
	if !found {
		t.Fatal("src has no local polar entry for dst")
	}
	srcCenter, ok := md.centerOfNode("src")
	if !ok {
		t.Fatal("no center for src")
	}
	dstCenter, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst")
	}
	// Nudge dst's live center by LESS than one bead — small enough that no legal cell
	// changed, but enough that a MEASURED-then-rounded QuantIR (the old defect) would
	// silently swallow the fraction instead of it showing up as a gap.
	dstCenter = dstCenter.Add(vec3{X: 0.31, Y: 0, Z: 0})
	liveDist := dstCenter.Sub(srcCenter).Length()

	_, _, rStep := lp.EffectiveSteps()
	// Reproduce the OLD defect: round the fresh live distance instead of trusting the
	// already-committed exact integer.
	buggyIR := int(math.Round(liveDist / rStep))
	buggyLP := lp
	buggyLP.QuantIR = buggyIR

	count := edgeStepCount(buggyLP, "SrcNode", "SinkNode")
	chainEndFarEdge := nodeTorusOuterR("SrcNode") + float64(count)*wire.BeadStepR
	targetSurfaceDist := liveDist - nodeTorusOuterR("SinkNode")
	diff := math.Abs(chainEndFarEdge - targetSurfaceDist)
	if diff < 0.05 {
		t.Fatalf("expected the round()-based measurement to reproduce a residue near the measured ~0.31 world units, got only %.6f — this test no longer demonstrates the defect it exists to prove", diff)
	}
	t.Logf("round()-based measurement residue: %.6f world units (proves the defect the cell model closes)", diff)
}

// synthCellAt reconstructs the world position ONE index delta implies, via the SAME
// offset formula dragCandidateCells uses (fromAxisFrame + polar2cart, neighbor minus
// offset) — a from-scratch re-derivation kept deliberately independent of
// dragCandidateCells' own code, so this test does not just call the function under test
// and trust its own output. Unlike dragCandidateCells' production {0,0,0} case (which
// returns the node's byte-identical CURRENT drawn position — nodeWorldPos(nm.geom), a
// value this synthetic single-cell test never sets up meaningfully), every delta here
// including {0,0,0} goes through the SAME formula, so distances between cells measure
// the LATTICE's own spacing rather than an arbitrary geom default.
func synthCellAt(lp wire.LocalPolar, pole dir, neighborPos vec3, delta [3]int) vec3 {
	stepTheta, stepPhi, stepR := lp.EffectiveSteps()
	iR := lp.QuantIR + delta[2]
	ndir := fromAxisFrame(pole, float64(lp.QuantITheta+delta[0])*stepTheta, float64(lp.QuantIPhi+delta[1])*stepPhi)
	off := polar2cart(polar{R: float64(iR) * stepR, Theta: ndir.Theta, Phi: ndir.Phi})
	return neighborPos.Sub(off)
}

// TestDragCellMovesExactlyOneBead pins the "each candidate is exactly one bead away"
// invariant (dragCellDeltas' doc comment) on EACH axis independently, at several radii —
// the angular case measured 7x-18x short under the old fixed-1-degree tick
// (docs/bead-lattice.md's drag.jump probe), so this is the regression the r-dependent
// AngularStepsForR (layout_holder.go) exists to close. Cross-checks its own oracle
// (synthCellAt) against dragCandidateCells' real output first, so a divergence between
// the two would fail loudly rather than silently testing only the oracle.
func TestDragCellMovesExactlyOneBead(t *testing.T) {
	// Chosen directly in quantIR (bead-count) units rather than a raw world radius: a raw
	// radius truncates through int(r/BeadStepR) to a DIFFERENT actual radius than the one
	// steps get derived from, and at a small quantIR the chord-vs-arc gap (chordSlack
	// below assumes small angles) is itself too large for the tolerance — both artifacts
	// of the TEST's setup, not of the production model, so side-stepped by picking a
	// radius large enough (in bead-steps) that neither applies.
	for _, quantIR := range []int{50, 200, 500} {
		lh := &wire.LayoutHolder{}
		r := float64(quantIR) * wire.BeadStepR
		// stepTheta/stepPhi are set EXPLICITLY to the r-derived values here (mirroring
		// what requantizePoleTraced's `fresh` branch actually writes for a brand-new
		// entry) rather than left unset (0): EffectiveSteps' unset fallback is the
		// small FIXED literal (layout_holder.go's degenerate-case constant), not a
		// recompute via AngularStepsForR, so an unset entry here would test the fixed
		// literal instead of the r-dependent model this test exists to pin.
		// Near the EQUATOR (theta0=90°), not near the pole: sin(theta0)=1 keeps stepPhi
		// the same small size as stepTheta (AngularStepsForR: stepPhi = BeadStepR /
		// (r*sin(theta))) — a small theta0 would make the phi line-of-latitude circle
		// much smaller than r, forcing a much LARGER angular step and a correspondingly
		// worse chord-vs-arc approximation, which is a property of the test's own
		// setup choice, not the production model.
		theta0 := 90.0 * math.Pi / 180
		stepTheta, _ := wire.AngularStepsForR(r, 0)
		_, stepPhi := wire.AngularStepsForR(r, theta0)
		iTheta := int(math.Round(theta0 / stepTheta))
		lh.SetLocalPolar("nbr", iTheta, 5, quantIR, stepTheta, stepPhi, wire.BeadStepR)
		neighborPos := vec3{X: r, Y: 0, Z: 0}
		nm := &nodeMover{id: "self", partnerCenters: map[string]vec3{"nbr": neighborPos}}
		lq := &layoutQuantizer{layoutHolders: map[string]*wire.LayoutHolder{"self": lh}}
		cells, ok := lq.dragCandidateCells(nm)
		if !ok {
			t.Fatalf("r=%v: no candidate cells", r)
		}
		var lp wire.LocalPolar
		for _, cand := range lh.LocalPolarsSnapshot() {
			if cand.To == "nbr" {
				lp = cand
			}
		}
		pole := dir(lh.Pole())

		// Cross-check every non-stay-put candidate dragCandidateCells actually produced
		// against the from-scratch oracle.
		for _, d := range dragCellDeltas {
			if d == [3]int{0, 0, 0} || lp.QuantIR+d[2] < 1 {
				continue
			}
			want := synthCellAt(lp, pole, neighborPos, d)
			matched := false
			for _, c := range cells {
				if c.Sub(want).Length() < 1e-9 {
					matched = true
				}
			}
			if !matched {
				t.Fatalf("r=%v delta=%v: oracle cell %+v not found in dragCandidateCells output %v", r, d, want, cells)
			}
		}

		// Delta index order matches dragCandidateCells' own convention exactly
		// (quantized_move.go: d[0]=theta, d[1]=phi, d[2]=radial) — {1,0,0}/{-1,0,0} are
		// THETA deltas, {0,0,1}/{0,0,-1} are RADIAL.
		stay := synthCellAt(lp, pole, neighborPos, [3]int{0, 0, 0})
		plus := synthCellAt(lp, pole, neighborPos, [3]int{0, 0, 1})
		minus := synthCellAt(lp, pole, neighborPos, [3]int{0, 0, -1})
		thetaPlus := synthCellAt(lp, pole, neighborPos, [3]int{1, 0, 0})
		phiPlus := synthCellAt(lp, pole, neighborPos, [3]int{0, 1, 0})

		// Radial axis: exactly one bead (BeadStepR), on every tested radius.
		if diff := math.Abs(stay.Sub(plus).Length() - wire.BeadStepR); diff > 1e-6 {
			t.Fatalf("r=%v: radial +1 cell is %.6f from stay-put, want exactly %.6f", r, stay.Sub(plus).Length(), wire.BeadStepR)
		}
		if diff := math.Abs(stay.Sub(minus).Length() - wire.BeadStepR); diff > 1e-6 {
			t.Fatalf("r=%v: radial -1 cell is %.6f from stay-put, want exactly %.6f", r, stay.Sub(minus).Length(), wire.BeadStepR)
		}
		// Angular axes: arc length (radius * angular step) is exactly one bead by
		// construction (AngularStepsForR derives the step FROM BeadStepR/r) — the chord
		// distance measured here is a hair SHORT of that arc for a finite step (the
		// honest "arc vs. chord" gap node_mover.go's own comment on this now names), so
		// the bound is generous rather than exact equality.
		const chordSlack = 1e-3 // relative
		if diff := math.Abs(stay.Sub(thetaPlus).Length()-wire.BeadStepR) / wire.BeadStepR; diff > chordSlack {
			t.Fatalf("r=%v: theta+1 cell chord %.6f too far from one bead of arc %.6f (relative diff %.6f)", r, stay.Sub(thetaPlus).Length(), wire.BeadStepR, diff)
		}
		if diff := math.Abs(stay.Sub(phiPlus).Length()-wire.BeadStepR) / wire.BeadStepR; diff > chordSlack {
			t.Fatalf("r=%v: phi+1 cell chord %.6f too far from one bead of arc %.6f (relative diff %.6f)", r, stay.Sub(phiPlus).Length(), wire.BeadStepR, diff)
		}
	}
}

// TestDragStayPutDoesNotMove: a drag too small to leave the current cell must not move the
// node at all — CLAUDE.md/the task spec's explicit requirement, kept working from the old
// walkBeadPath-era behavior.
func TestDragStayPutDoesNotMove(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	nm, ok := md.mr.nodeMovers["dst"]
	if !ok {
		t.Fatal("no nodeMover for dst")
	}
	before := nodeWorldPos(nm.geom)
	// A tiny nudge, well under one bead (8.96) in every direction.
	target := before.Add(vec3{X: 0.05, Y: -0.02, Z: 0.01})
	got, ok := md.lq.chooseDragCandidate(nm, target)
	if !ok {
		t.Fatal("chooseDragCandidate found no candidate cells for dst")
	}
	if got != before {
		t.Fatalf("a sub-bead drag target moved the node: before=%+v got=%+v (target=%+v)", before, got, target)
	}
}

// TestCommittedQuantIRMatchesActualDistance: committed QuantIR * BeadStepR must equal the
// node's actual distance to the chosen neighbor — the "by construction" claim the whole
// cell model rests on, since nothing measures a distance and rounds it any more (the
// candidate cartesian point IS index*step, not the other way around).
func TestCommittedQuantIRMatchesActualDistance(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := md.Start(ctx)
	t.Cleanup(func() { cancel(); wg.Wait() })

	nm, ok := md.mr.nodeMovers["dst"]
	if !ok {
		t.Fatal("no nodeMover for dst")
	}
	before, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst")
	}
	target := before.Add(vec3{X: -22.3, Y: 6.6, Z: 17.9})
	want, wantOK := md.lq.chooseDragCandidate(nm, target)
	if !wantOK {
		t.Fatal("chooseDragCandidate found no candidate cells for dst")
	}
	if !md.RootMove("dst", target) {
		t.Fatal("RootMove(dst) returned false")
	}
	pollDragConverged(t, md, "dst", want)

	lh := md.lq.layoutHolders["dst"]
	var lp wire.LocalPolar
	found := false
	for _, cand := range lh.LocalPolarsSnapshot() {
		if cand.To == "src" {
			lp, found = cand, true
		}
	}
	if !found {
		t.Fatal("dst has no local polar entry for src after the drag")
	}
	dstCenter, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst")
	}
	srcCenter, ok := md.centerOfNode("src")
	if !ok {
		t.Fatal("no center for src")
	}
	actualDist := dstCenter.Sub(srcCenter).Length()
	_, _, rStep := lp.EffectiveSteps()
	wantDist := float64(lp.QuantIR) * rStep
	if diff := math.Abs(actualDist - wantDist); diff > 1e-6 {
		t.Fatalf("QuantIR*rStep=%.6f does not match dst's actual distance to src=%.6f (diff=%.6f)", wantDist, actualDist, diff)
	}
}

// TestAngularIndexPreservesAngleAcrossRadiusChange: a radius change preserves the ANGLE
// across the angular index re-derive LoadLocalPolars does (layout_holder.go), not the
// index itself — index and step are ONE value (index*step); a step that changes size
// with the stored index left untouched would silently multiply the represented angle,
// the same bug class the radial QuantIR/StepR conversion already guards against.
func TestAngularIndexPreservesAngleAcrossRadiusChange(t *testing.T) {
	const oldStepTheta = math.Pi / 180 // the old fixed 1-degree tick, predating AngularStepsForR
	const quantIR = 30                 // radius = 30 * wire.LocalStepR (already the current lattice)
	const quantITheta = 41

	angleBefore := float64(quantITheta) * oldStepTheta

	lh := &wire.LayoutHolder{}
	lh.LoadLocalPolars([]wire.LocalPolar{
		{To: "nbr", QuantITheta: quantITheta, QuantIPhi: 0, QuantIR: quantIR,
			StepTheta: oldStepTheta, StepPhi: oldStepTheta, StepR: wire.LocalStepR},
	})
	got := lh.LocalPolarsSnapshot()
	if len(got) != 1 {
		t.Fatalf("want 1 entry, got %d", len(got))
	}
	// The step must have actually changed (proves the re-derive ran, not a no-op).
	if got[0].StepTheta == oldStepTheta {
		t.Fatalf("expected StepTheta to be re-derived away from the old fixed tick %v, still %v", oldStepTheta, got[0].StepTheta)
	}
	angleAfter := float64(got[0].QuantITheta) * got[0].StepTheta
	// Tolerance is half the NEW step — the same rounding tolerance LoadLocalPolars'
	// own radial conversion already accepts (quantized_layout_test.go's
	// TestNormalizeOffsetConvertsIndexAndStepTogether uses the same "within half a
	// step" bound).
	if diff := math.Abs(angleBefore - angleAfter); diff > got[0].StepTheta/2 {
		t.Fatalf("angle not preserved across the re-derive: before=%v after=%v (diff=%v, allowed<=%v)", angleBefore, angleAfter, diff, got[0].StepTheta/2)
	}

	// Proof of failure: leaving the index untouched while only the step changes (the
	// "index left alone, step silently swapped" bug class) would multiply the
	// represented angle by a much larger margin than half a step.
	buggyAngle := float64(quantITheta) * got[0].StepTheta
	if diff := math.Abs(angleBefore - buggyAngle); diff <= got[0].StepTheta/2 {
		t.Fatalf("test setup invalid: the step-only bug should have changed the represented angle by more than half a step, but didn't (before=%v buggy=%v)", angleBefore, buggyAngle)
	}
}
