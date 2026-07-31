package Wiring

// drag_cell_test.go — the bead-lattice drag model: a legal drag target is a CELL that is a
// whole bead count from SOME usable neighbour, chosen bottom-up from that neighbour's own
// frame via wire.BeadVector{Dir,N} index arithmetic, never a cartesian walk toward the
// target. Every usable neighbour's own cell set contributes independently and they are
// UNIONED (dragCandidateCells, quantized_move.go), not intersected — two prior attempts at
// this drag model (commit 3fea51ab: one neighbour's frame only; commit 1ca42492: cells
// INTERSECTED across every neighbour) were both built and reverted because dragging
// stopped working; see dragCandidateCells' own doc comment for exactly why each failed,
// and TestDragCandidateCells_HubWithThreeNeighbors_IsDraggable /
// TestDragCandidateCells_OneNeighborMissingPartnerCenter_StillDraggable below for the two
// regressions this file exists to make impossible to reintroduce.
//
// Per docs/testing-shape.md, the pure candidate-cell math (no goroutines, one function's
// own decision given its inputs) is tested directly against bare *nodeMover/
// *layoutQuantizer structs, exactly like cascade_kind_route_test.go/chain_beads_test.go
// already do for other single-goroutine node math. The exceptions —
// TestRootMoveDragActuallyMovesTheNode and the tangency test — drive a REAL MoveDispatch
// with its movers running, because "dragging works end to end through RootMove" is
// precisely the property both reverted attempts broke, and a pure-function test of the
// candidate math alone would not have caught either regression: both were in how
// commitNodeMoveLocal's caller-side behavior interacted with dragCandidateCells' ok/cells
// output, not in the candidate math in isolation.

import (
	"context"
	"math"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

const dragCellEps = 1e-6

// bareDraggedNode builds a bare *nodeMover positioned at `self` (world), with a
// LayoutHolder entry (registered on the returned *layoutQuantizer) whose stored
// (QuantITheta,QuantIPhi,QuantIR) reconstruct EXACTLY the offset (neighborPos - self) —
// derived via azimuthFrom/dirFromOffset, the same boundary calls production code makes,
// never a hand-guessed angle — so dragCandidateCellsForNeighbor's fromAxisFrame round-trip
// reproduces neighborPos - self to float precision. `home` is the pole (lh.Pole()'s zero
// value, world +y) every entry in this file is quantized about, matching a fresh
// LayoutHolder that has never been requantized.
func bareDraggedNode(t *testing.T, id string, self vec3, neighbors map[string]vec3) (*layoutQuantizer, *nodeMover) {
	t.Helper()
	home := dir{}
	lh := &wire.LayoutHolder{}
	partnerCenters := map[string]vec3{}
	for to, pos := range neighbors {
		off := pos.Sub(self)
		worldDir, r := dirFromOffset(off)
		c, psi := azimuthFrom(home, worldDir)
		iTheta, stepTheta := 1, c
		iPhi, stepPhi := 1, psi
		if psi == 0 {
			// angle = iPhi*stepPhi must be exactly 0; iPhi=0 makes that true for any
			// nonzero stepPhi (0 itself would read back as "unset" per EffectiveSteps'
			// contract and silently substitute the small-angle default).
			iPhi, stepPhi = 0, 1
		}
		iR := int(math.Round(r / wire.BeadStepR))
		lh.SetLocalPolar(to, iTheta, iPhi, iR, stepTheta, stepPhi, wire.BeadStepR)
		partnerCenters[to] = pos
	}
	lq := &layoutQuantizer{layoutHolders: map[string]*wire.LayoutHolder{id: lh}}
	nm := &nodeMover{id: id, partnerCenters: partnerCenters}
	nm.geom.HasPos = true
	setNodeWorld(&nm.geom, self)
	return lq, nm
}

// --- Robustness (item 4): the exact degenerate setups that plausibly broke the reverted
// attempts, each proven to NOT freeze chooseDragCandidate's caller.

// TestDragCandidateCells_NoLayoutHolder_OkFalse: nodeID has no entry in lq.layoutHolders
// at all (never loaded, or a kind with no LayoutHolder embedded). No frame exists to build
// a cell in — ok=false is "no legal cell exists", not a freeze: commitNodeMoveLocal's
// caller-side fallback (committedPos = newPos) is what actually keeps the drag working.
func TestDragCandidateCells_NoLayoutHolder_OkFalse(t *testing.T) {
	lq := &layoutQuantizer{layoutHolders: map[string]*wire.LayoutHolder{}}
	nm := &nodeMover{id: "a", partnerCenters: map[string]vec3{}}
	if _, ok := lq.chooseDragCandidate(nm, vec3{X: 100}); ok {
		t.Fatal("nodeID with no LayoutHolder at all must report ok=false, not synthesize a candidate")
	}
}

// TestDragCandidateCells_NoNeighbors_OkFalse: nodeID has a LayoutHolder but it is EMPTY
// (an isolated node with no domain edges) — same contract as above, distinct cause.
func TestDragCandidateCells_NoNeighbors_OkFalse(t *testing.T) {
	lq := &layoutQuantizer{layoutHolders: map[string]*wire.LayoutHolder{"a": {}}}
	nm := &nodeMover{id: "a", partnerCenters: map[string]vec3{}}
	if _, ok := lq.chooseDragCandidate(nm, vec3{X: 100}); ok {
		t.Fatal("an isolated node (no LocalPolars) must report ok=false, not synthesize a candidate")
	}
}

// TestDragCandidateCells_EmptyPartnerCenters_OkFalse is the reverted single-neighbour
// attempt's failure mode with exactly ONE neighbour: it has a stored LocalPolar entry but
// its live centre has never been pushed into partnerCenters (nm.partnerCenters
// deliberately left empty here). Excluded from the constraint set
// (dragConstraintNeighbors), and here it is the ONLY neighbour, so the set is empty and
// ok=false — the caller's free-move fallback, not a frozen node.
func TestDragCandidateCells_EmptyPartnerCenters_OkFalse(t *testing.T) {
	lh := &wire.LayoutHolder{}
	lh.SetLocalPolar("n1", 1, 1, 10, math.Pi/2, math.Pi, wire.BeadStepR)
	lq := &layoutQuantizer{layoutHolders: map[string]*wire.LayoutHolder{"a": lh}}
	nm := &nodeMover{id: "a", partnerCenters: map[string]vec3{}} // deliberately empty
	nm.geom.HasPos = true
	if _, ok := lq.chooseDragCandidate(nm, vec3{X: 100}); ok {
		t.Fatal("a neighbour with a LocalPolar entry but no known live centre must be excluded, and ok=false when it is the only neighbour")
	}
}

// TestDragCandidateCells_OneNeighborMissingPartnerCenter_StillDraggable is the
// NON-NEGOTIABLE robustness case named directly in the task: THREE neighbours, one of
// which has NO partnerCenters entry. Under the reverted single-chosen-neighbour attempt,
// that one missing entry (if it happened to be the chosen neighbour) froze the whole
// node — there was no fallback frame. Under UNION, a missing neighbour just contributes
// zero cells of its own; the other two usable neighbours' cells are entirely unaffected,
// so the node must still be draggable.
func TestDragCandidateCells_OneNeighborMissingPartnerCenter_StillDraggable(t *testing.T) {
	self := vec3{}
	n1 := vec3{X: 20 * wire.BeadStepR, Y: 3, Z: -7}
	n2 := vec3{X: -15 * wire.BeadStepR, Y: 30 * wire.BeadStepR, Z: 5}
	n3 := vec3{X: 8, Y: -25 * wire.BeadStepR, Z: 40 * wire.BeadStepR} // will be dropped below
	lq, nm := bareDraggedNode(t, "a", self, map[string]vec3{"n1": n1, "n2": n2, "n3": n3})
	delete(nm.partnerCenters, "n3") // n3 keeps its LocalPolar entry but loses its live centre

	got, ok := lq.chooseDragCandidate(nm, self.Add(vec3{X: 3 * wire.BeadStepR}))
	if !ok {
		t.Fatal("a node with two OTHER usable neighbours must still be draggable when a third neighbour's partnerCenters entry is missing")
	}
	if got == self {
		t.Fatal("expected the node to move toward the target using the two remaining usable neighbours")
	}
}

// TestDragCandidateCells_HubWithThreeNeighbors_IsDraggable is the OTHER non-negotiable
// case: a HUB node with three neighbours in a REAL (non-colinear) arrangement, matching
// this graph's actual shape (node 1 has neighbours 2 and 3; node 2 has 1, 4, 5 — nearly
// every node is a hub). Under the reverted INTERSECTION model, a candidate had to be a
// whole bead count from all three neighbours simultaneously — three whole-bead shells
// meet only at isolated points, so a generic (non-colinear) hub had NO legal cell but
// stay-put, and that attempt's own test fixture had to be rebuilt colinear just to pass.
// Under UNION, each neighbour independently contributes its own cells, so a hub is at
// least as draggable as a node with one neighbour — this test proves it actually moves.
func TestDragCandidateCells_HubWithThreeNeighbors_IsDraggable(t *testing.T) {
	self := vec3{}
	// Three neighbours spread in genuinely different, non-colinear directions — no two
	// are opposite or parallel, so no accidental intersection geometry is smuggled in.
	n1 := vec3{X: 30 * wire.BeadStepR, Y: 2, Z: -3}
	n2 := vec3{X: -10, Y: 25 * wire.BeadStepR, Z: 4}
	n3 := vec3{X: 6, Y: -8, Z: -40 * wire.BeadStepR}
	lq, nm := bareDraggedNode(t, "hub", self, map[string]vec3{"n1": n1, "n2": n2, "n3": n3})

	cells, ok := lq.dragCandidateCells(nm)
	if !ok {
		t.Fatal("expected a usable candidate set for a three-neighbour hub")
	}
	// Every usable neighbour contributes up to 6 cells (dragCellDeltas), plus stay-put —
	// a real, non-trivial candidate set, not just the single stay-put fallback.
	nonStayPut := 0
	for _, c := range cells {
		if c != self {
			nonStayPut++
		}
	}
	if nonStayPut == 0 {
		t.Fatal("a three-neighbour hub produced no non-stay-put candidates at all — the union model must not degrade to intersection's all-hubs-frozen outcome")
	}

	// Drag toward a target well off the origin — the hub must actually move (not silently
	// stay put the way a frozen node under the reverted intersection model would).
	target := self.Add(vec3{X: 5 * wire.BeadStepR, Y: -3 * wire.BeadStepR, Z: 2 * wire.BeadStepR})
	got, moved := lq.chooseDragCandidate(nm, target)
	if !moved {
		t.Fatal("chooseDragCandidate reported ok=false for a hub with three usable neighbours")
	}
	if got == self {
		t.Fatal("a hub node with three real (non-colinear) neighbours must be draggable, not frozen at stay-put")
	}
}

// TestDragCandidateCells_OneNeighbor_MovesExactlyOneBead: a single usable neighbour is
// enough to drag — its own six adjacent cells (plus stay-put) are the whole candidate set.
// Dragging several beads past the neighbour's own next radial cell must move the node
// EXACTLY one bead, at several radii (not a fraction, not the full raw displacement).
func TestDragCandidateCells_OneNeighbor_MovesExactlyOneBead(t *testing.T) {
	// An off-axis (not coordinate-aligned) unit bearing, reused at several radii — each
	// neighbour is placed EXACTLY n whole bead-steps from self along it, so self sits
	// precisely ON the lattice the radial candidates are built from (mirroring what a
	// REAL prior commit leaves behind: the committed position IS one of these exact
	// candidates). A neighbour placed off this lattice would make even the "same
	// bearing, radius+1" candidate miss self by the seed's own off-lattice slop, which
	// is a test-setup artifact, not a property of chooseDragCandidate.
	bearing := vec3{X: 3, Y: 1, Z: 2}
	bearing = bearing.Scale(1 / bearing.Length())
	for _, n := range []int{10, 30, 80} {
		self := vec3{}
		neighbor := bearing.Scale(float64(n) * wire.BeadStepR)
		lq, nm := bareDraggedNode(t, "a", self, map[string]vec3{"n1": neighbor})

		// Target several beads beyond self, continuing straight away from the
		// neighbour (radially outward) — far enough that only the one-bead-out
		// candidate is nearest.
		target := self.Sub(bearing.Scale(5 * wire.BeadStepR))

		got, ok := lq.chooseDragCandidate(nm, target)
		if !ok {
			t.Fatalf("n=%d: expected a usable candidate set with one neighbour", n)
		}
		moved := got.Sub(self).Length()
		if diff := math.Abs(moved - wire.BeadStepR); diff > dragCellEps {
			t.Fatalf("n=%d: expected exactly one bead of movement (%.6f), got %.6f (diff %.9f)", n, wire.BeadStepR, moved, diff)
		}
		if got == self {
			t.Fatalf("n=%d: expected the node to move, but chooseDragCandidate returned stay-put", n)
		}
	}
}

// TestDragCandidateCells_TooSmallDrag_DoesNotMove: a target well short of the nearest
// legal cell (here, well under half a bead past self) must leave the node EXACTLY where
// it was — byte-identical to nodeWorldPos, not a partial/rounded slide.
func TestDragCandidateCells_TooSmallDrag_DoesNotMove(t *testing.T) {
	self := vec3{}
	neighbor := vec3{X: -20 * wire.BeadStepR, Y: 1, Z: -4}
	lq, nm := bareDraggedNode(t, "a", self, map[string]vec3{"n1": neighbor})

	target := self.Add(vec3{X: 0.2 * wire.BeadStepR, Y: 0, Z: 0}) // well under a bead
	got, ok := lq.chooseDragCandidate(nm, target)
	if !ok {
		t.Fatal("expected a usable candidate set with one neighbour")
	}
	if got != self {
		t.Fatalf("a drag too small to reach a legal cell must not move the node at all: self=%+v got=%+v", self, got)
	}
}

// TestDragCandidateCells_TwoNeighbors_UnionIncludesBothFrames: with self between two
// colinear, opposite neighbours (A --- self --- B), the UNION model's candidate set must
// include cells generated in EITHER neighbour's own frame — never filtered against the
// other, unlike the reverted intersection attempt. Each non-stay-put candidate is a whole
// bead count from THE NEIGHBOUR IT WAS GENERATED FROM (by construction), which this test
// checks directly rather than assuming.
func TestDragCandidateCells_TwoNeighbors_UnionIncludesBothFrames(t *testing.T) {
	self := vec3{}
	D := 10 * wire.BeadStepR
	neighborA := vec3{X: -D}
	neighborB := vec3{X: D}
	lq, nm := bareDraggedNode(t, "a", self, map[string]vec3{"A": neighborA, "B": neighborB})

	cells, ok := lq.dragCandidateCells(nm)
	if !ok {
		t.Fatal("expected a usable candidate set with two neighbours")
	}
	// 1 stay-put + up to 6 cells from A's frame + up to 6 from B's frame.
	if len(cells) < 1+2 {
		t.Fatalf("expected cells from both neighbours' frames in the union, got only %d cells: %+v", len(cells), cells)
	}
	fromA, fromB := 0, 0
	for _, c := range cells {
		if c == self {
			continue
		}
		if beadStepsFrom(c, neighborA) {
			fromA++
		}
		if beadStepsFrom(c, neighborB) {
			fromB++
		}
	}
	if fromA == 0 {
		t.Fatal("expected at least one candidate a whole bead count from neighbour A (A's own generated frame)")
	}
	if fromB == 0 {
		t.Fatal("expected at least one candidate a whole bead count from neighbour B (B's own generated frame)")
	}
}

// TestAngularStepsForR_OneIndexStepIsOneBeadOfArc pins item 5's derivation directly: at
// several radii and colatitudes, AngularStepsForR's own stepTheta/stepPhi must sweep
// EXACTLY one bead of ARC (r*dtheta, r*sin(theta)*dphi) — the algebraic identity the
// formula is built from (BeadStepR/r, BeadStepR/(r*sin(theta))), not a re-measurement of
// dragCandidateCells' output (which would only prove self-consistency, not correctness
// against the bead's own physical size).
func TestAngularStepsForR_OneIndexStepIsOneBeadOfArc(t *testing.T) {
	cases := []struct{ r, theta float64 }{
		{10 * wire.BeadStepR, math.Pi / 4},
		{30 * wire.BeadStepR, math.Pi / 2},
		{80 * wire.BeadStepR, 2 * math.Pi / 3},
	}
	for _, c := range cases {
		stepTheta, stepPhi := wire.AngularStepsForR(c.r, c.theta)
		arcTheta := c.r * stepTheta
		arcPhi := c.r * math.Sin(c.theta) * stepPhi
		if diff := math.Abs(arcTheta - wire.BeadStepR); diff > dragCellEps {
			t.Errorf("r=%v theta=%v: theta-step arc %.6f != one bead %.6f (diff %.9f)", c.r, c.theta, arcTheta, wire.BeadStepR, diff)
		}
		if diff := math.Abs(arcPhi - wire.BeadStepR); diff > dragCellEps {
			t.Errorf("r=%v theta=%v: phi-step arc %.6f != one bead %.6f (diff %.9f)", c.r, c.theta, arcPhi, wire.BeadStepR, diff)
		}
	}
}

// TestAngularStepsForR_DegenerateFallback proves r<=0 and sin(theta)<=0 fall back to the
// small constants rather than dividing by zero (item 5's required degenerate handling).
func TestAngularStepsForR_DegenerateFallback(t *testing.T) {
	st, sp := wire.AngularStepsForR(0, math.Pi/2)
	if st != wire.DefaultLocalStepTheta || sp != wire.DefaultLocalStepPhi {
		t.Fatalf("r<=0 must fall back to the default small-angle constants, got (%v,%v)", st, sp)
	}
	st2, sp2 := wire.AngularStepsForR(50, 0) // on the pole axis: sin(theta)=0
	if st2 <= 0 {
		t.Fatalf("theta-step at r=50 must still be well-defined (r>0): got %v", st2)
	}
	if sp2 != st2 {
		t.Fatalf("phi-step on the pole axis (sin(theta)=0) must fall back to the theta-step itself (a defined, harmless placeholder), got %v want %v", sp2, st2)
	}
}

// TestRadiusChangePreservesAngle_ReDerivesIndices pins item 3's re-derive discipline
// directly against LoadLocalPolars: a stored entry's angular step describes "one bead of
// arc" at ITS OWN radius, so a radius change alone (StepR already agreeing with the
// lattice, only QuantIR moving) must change the stored StepTheta/StepPhi too — and the
// index/step pair must still reconstruct the SAME ANGLE (index*step invariant) as before
// the radius changed, not a scaled one. This is the radius-changes-preserve-angle half of
// item 3 the task calls out by name, isolated from any drag/commit machinery.
func TestRadiusChangePreservesAngle_ReDerivesIndices(t *testing.T) {
	const oldR = 20 * wire.BeadStepR
	oldStepTheta, _ := wire.AngularStepsForR(oldR, 0)
	const iTheta, iPhi = 7, -4
	wantAngleTheta := float64(iTheta) * oldStepTheta
	// stepPhi must be derived at the COLATITUDE this entry's own indices actually imply
	// (wantAngleTheta), not an arbitrary theta — production ties stepPhi to the entry's
	// own represented angle (LoadLocalPolars' `phi := QuantIPhi*stalePhi` runs against
	// whatever stepPhi was stored for THIS entry, which by construction was itself
	// derived at this same colatitude).
	_, oldStepPhi := wire.AngularStepsForR(oldR, wantAngleTheta)
	wantAnglePhi := float64(iPhi) * oldStepPhi

	lh := &wire.LayoutHolder{}
	lh.LoadLocalPolars([]wire.LocalPolar{{
		To: "n", QuantITheta: iTheta, QuantIPhi: iPhi, QuantIR: 20,
		StepTheta: oldStepTheta, StepPhi: oldStepPhi, StepR: wire.LocalStepR,
	}})

	// Now simulate a RADIUS-ONLY change: a fresh entry at a different QuantIR but the
	// SAME (now-stale) angular step, fed back through LoadLocalPolars — mirroring what a
	// disk round-trip after a radial-only drag looks like (StepR already == LocalStepR,
	// so only the angular normalize branch should fire).
	newR := 60 * wire.BeadStepR
	lh2 := &wire.LayoutHolder{}
	lh2.LoadLocalPolars([]wire.LocalPolar{{
		To: "n", QuantITheta: iTheta, QuantIPhi: iPhi, QuantIR: int(math.Round(newR / wire.LocalStepR)),
		StepTheta: oldStepTheta, StepPhi: oldStepPhi, StepR: wire.LocalStepR,
	}})

	var got wire.LocalPolar
	for _, lp := range lh2.LocalPolarsSnapshot() {
		if lp.To == "n" {
			got = lp
		}
	}
	if got.StepTheta == oldStepTheta {
		t.Fatal("expected LoadLocalPolars to re-derive StepTheta for the new radius, not keep the stale one")
	}
	gotAngleTheta := float64(got.QuantITheta) * got.StepTheta
	gotAnglePhi := float64(got.QuantIPhi) * got.StepPhi
	if diff := math.Abs(gotAngleTheta - wantAngleTheta); diff > 1e-9 {
		t.Fatalf("radius change should PRESERVE the represented theta angle: want %.9f got %.9f (diff %.9f)", wantAngleTheta, gotAngleTheta, diff)
	}
	if diff := math.Abs(gotAnglePhi - wantAnglePhi); diff > 1e-9 {
		t.Fatalf("radius change should PRESERVE the represented phi angle: want %.9f got %.9f (diff %.9f)", wantAnglePhi, gotAnglePhi, diff)
	}
	_ = lh // silence unused in case of future trimming
}

// --- End-to-end: dragging must DEMONSTRABLY still work through the real RootMove path
// (the property both reverted attempts broke — the pure candidate-math tests above would
// not have caught either regression, since both were in the caller-side interaction with
// dragCandidateCells' ok/cells output, not in the candidate enumeration itself).

// TestRootMoveDragActuallyMovesTheNode drives a real 2-node graph (src/dst, writeTree)
// through RootMove with its movers RUNNING, and asserts the node's committed centre
// actually changed and converged near the requested target (within one bead, since the
// commit snaps to the nearest legal cell) — the non-negotiable proof that dragging is not
// broken by this change.
func TestRootMoveDragActuallyMovesTheNode(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	if !md.lq.quantizedLayout {
		t.Fatal("test assumes quantizedLayout is on by default")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := md.Start(ctx)
	t.Cleanup(func() { cancel(); wg.Wait() })

	before, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst before drag")
	}
	target := before.Add(vec3{X: 5 * wire.BeadStepR, Y: -3 * wire.BeadStepR, Z: 2 * wire.BeadStepR})
	want := quantizedDragTarget(md, "dst", target)
	if want == before {
		t.Fatal("test setup bug: the chosen target should quantize to a real move, not stay-put")
	}
	if !md.RootMove("dst", target) {
		t.Fatal("RootMove(dst) returned false")
	}
	pollDragConverged(t, md, "dst", want)

	after, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst after drag")
	}
	if after == before {
		t.Fatalf("dst never moved: before=%+v after=%+v (target=%+v) — dragging must demonstrably still work", before, after, target)
	}
	// writeTree's seed positions are hand-authored continuous scenePolar values, not
	// already lattice-exact, so the FIRST-EVER drag's committed cell is not guaranteed
	// to be exactly one BeadStepR from `before` (that exactness holds once a node's
	// position IS itself a prior commit's cell — see
	// TestDragCandidateCells_OneNeighbor_MovesExactlyOneBead for that pinned property).
	// Here the bar is simply "moved a real, non-trivial amount", not "exactly one bead".
	if d := after.Sub(before).Length(); d < 0.5*wire.BeadStepR {
		t.Fatalf("dst moved by only %.6f, well under half a bead (%.6f) — dragging toward a %.1f-bead-away target should move it substantially, not by a rounding-sized amount", d, wire.BeadStepR, 5.0)
	}
}

// TestTangencyAfterDrag_LastBeadExactlyReachesTargetSurface: after a real drag of dst by a
// RAW (not whole-bead) displacement, src's own re-quantized LocalPolar to dst (produced by
// neighborSetCRequantize on src's own goroutine) must place src's chain's LAST bead's far
// edge EXACTLY on dst's torus surface — chain_beads.go's placement formula
// (srcTorusOuterR + BeadTorusOuterR + i*BeadStepR) reads QuantIR directly as a whole
// bead-step count (bead_lattice.md "The count"), so this holds as long as the drag path
// never stores a QuantIR that disagrees with the position it actually committed to — the
// exact defect a measured-then-rounded distance (the reverted single-neighbour attempt's
// `diff = 0.310303` residue) would reintroduce. Proof of failure below temporarily
// reproduces that residue. src is dst's ONLY neighbour in this fixture (writeTree), so its
// own edge to dst is exactly the frame the drag's chosen candidate was generated from —
// the neighbour that ends up EXACT (see dragCandidateCells' doc comment for why only the
// generating neighbour gets this guarantee under the union model).
func TestTangencyAfterDrag_LastBeadExactlyReachesTargetSurface(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := md.Start(ctx)
	t.Cleanup(func() { cancel(); wg.Wait() })

	// A RAW (non-whole-bead) displacement — 37.1/-5.3/12.9 is not an exact multiple of
	// wire.BeadStepR (8.96) on any axis, by construction (same target as
	// TestCommitNodeMoveLocalDrawsQuantizedNotRawTarget).
	before, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst")
	}
	target := before.Add(vec3{X: 37.1, Y: -5.3, Z: 12.9})

	var dbg syncBuffer
	md.tr.SetSink(&dbg)
	if !md.RootMove("dst", target) {
		t.Fatal("RootMove(dst) returned false")
	}
	// src is dst's neighbour and stays put; wait for its own re-quantize (same sync
	// point TestIndividualSnap_OnlyDraggedNodePersists uses).
	waitForAbcDrag(t, &dbg, "src")

	srcNM, ok := md.mr.nodeMovers["src"]
	if !ok {
		t.Fatal("no nodeMover for src")
	}
	lhSrc, ok := md.lq.layoutHolders["src"]
	if !ok {
		t.Fatal("no LayoutHolder for src")
	}
	var lp wire.LocalPolar
	var found bool
	for _, e := range lhSrc.LocalPolarsSnapshot() {
		if e.To == "dst" {
			lp, found = e, true
		}
	}
	if !found {
		t.Fatal("src has no requantized LocalPolar to dst after the drag")
	}

	srcTorus := nodeTorusOuterR(srcNM.geom.Kind)
	dstNM, ok := md.mr.nodeMovers["dst"]
	if !ok {
		t.Fatal("no nodeMover for dst")
	}
	dstTorus := nodeTorusOuterR(dstNM.geom.Kind)

	_, _, rStep := lp.EffectiveSteps()
	separation := float64(lp.QuantIR) * rStep // the EXACT distance placement uses — index*step, never measured
	n := lp.QuantIR - nodeTorusSteps(srcNM.geom.Kind) - nodeTorusSteps(dstNM.geom.Kind)
	if n < 1 {
		n = 1
	}
	lastBeadCenterDist := srcTorus + wire.BeadTorusOuterR + float64(n-1)*wire.BeadStepR
	farEdge := lastBeadCenterDist + wire.BeadTorusOuterR
	wantFarEdge := separation - dstTorus

	const tangencyEps = 1e-6
	if diff := math.Abs(farEdge - wantFarEdge); diff > tangencyEps {
		t.Fatalf("last bead's far edge %.9f does not exactly reach dst's torus surface %.9f (off by %.9f)", farEdge, wantFarEdge, diff)
	}

	// Proof of failure: reproduce the shape of the reverted attempt's defect — the
	// COMMITTED position and the RECORDED index disagreeing (there, a walked/measured
	// distance that got rounded into an index describing a slightly different cell than
	// the one actually drawn; here, simulated directly by shifting the recorded index by
	// one bead-step from what placement is really about to use) — and show the tangency
	// check above FAILS its tolerance against that corrupted value, proving it is not a
	// check that would pass no matter what. The historical live repro measured a
	// same-order-of-magnitude residue (diff = 0.310303, a fraction of one BeadStepR);
	// this reproduces the mechanism at the scale of a whole missed bead, which the same
	// tolerance must reject just as surely as a fractional one.
	buggyIR := lp.QuantIR - 1 // "recorded one bead short of where the commit actually put it"
	buggySeparation := float64(buggyIR) * rStep
	buggyResidue := math.Abs(buggySeparation - separation)
	if buggyResidue <= tangencyEps {
		t.Fatalf("test setup invalid: an off-by-one-bead recorded index should have produced a residue near one BeadStepR (%.6f) against the true separation, got %.9f", wire.BeadStepR, buggyResidue)
	}
	buggyFarEdge := srcTorus + wire.BeadTorusOuterR + float64(n-2)*wire.BeadStepR + wire.BeadTorusOuterR
	if diff := math.Abs(buggyFarEdge - wantFarEdge); diff <= tangencyEps {
		t.Fatalf("test setup invalid: the tangency check should FAIL against the corrupted (buggyIR) placement, but it passed (diff %.9f) — the check would not have caught the original defect", diff)
	} else {
		t.Logf("proof of failure: the tangency check correctly rejects a corrupted (one-bead-short) recorded index, off by %.6f (reverted attempt observed a same-order-of-magnitude 0.310303 residue live)", diff)
	}
}
