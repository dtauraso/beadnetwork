package Wiring

import (
	"math"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// bead_cell_solve_test.go — pure single-goroutine tests of the bead-cell solver
// (bead_cell_solve.go): "a node lives in N lattices, one per neighbour" — the intersection
// of concentric spheres, per neighbour, at whole wire.BeadStepR multiples. Per
// docs/testing-shape.md this is pure geometry, no goroutines: exactly the shape that
// belongs in a unit test.

func assertOnLattice(t *testing.T, p vec3, neighbors []beadCellNeighbor) {
	t.Helper()
	for _, nb := range neighbors {
		d := p.Sub(nb.Center).Length()
		ratio := d / wire.BeadStepR
		if math.Abs(ratio-math.Round(ratio)) > 1e-6 {
			t.Fatalf("point %+v is not an integer number of BeadStepR from neighbor center %+v: dist=%v ratio=%v", p, nb.Center, d, ratio)
		}
	}
}

func TestSolveBeadCells_OneNeighbor_SphereSurface(t *testing.T) {
	nb := beadCellNeighbor{Center: vec3{}, K: 3}
	target := vec3{X: 100, Y: 5, Z: -7}
	cands := solveBeadCells([]beadCellNeighbor{nb}, target)
	if len(cands) == 0 {
		t.Fatal("expected at least one candidate for a single-neighbour node")
	}
	for _, c := range cands {
		assertOnLattice(t, c, []beadCellNeighbor{nb})
	}
}

func TestSolveBeadCells_OneNeighbor_NearestToTargetOnChosenSphere(t *testing.T) {
	nb := beadCellNeighbor{Center: vec3{}, K: 3}
	target := vec3{X: 1000, Y: 0, Z: 0}
	got := snapToBeadCell(vec3{X: 3 * wire.BeadStepR, Y: 0, Z: 0}, []beadCellNeighbor{nb}, target)
	// The enumeration considers K-1, K, K+1 (2, 3, 4); the far +X target is nearest the
	// LARGEST reachable radius, so the winning candidate sits on the K+1=4 sphere.
	want := vec3{X: 4 * wire.BeadStepR, Y: 0, Z: 0}
	if d := got.Sub(want).Length(); d > 1e-6 {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestSolveBeadCells_TwoNeighbors_OnBothSpheres(t *testing.T) {
	c1 := vec3{X: -20, Y: 0, Z: 0}
	c2 := vec3{X: 20, Y: 0, Z: 0}
	nbs := []beadCellNeighbor{{Center: c1, K: 5}, {Center: c2, K: 5}}
	target := vec3{X: 0, Y: 50, Z: 0}
	cands := solveBeadCells(nbs, target)
	if len(cands) == 0 {
		t.Fatal("expected candidates for a two-neighbour node")
	}
	for _, c := range cands {
		assertOnLattice(t, c, nbs)
	}
}

func TestSolveBeadCells_ThreeNeighbors_DiscretePoints(t *testing.T) {
	// Three neighbour centers not colinear, radii chosen so the spheres genuinely meet.
	c1 := vec3{X: -10, Y: 0, Z: 0}
	c2 := vec3{X: 10, Y: 0, Z: 0}
	c3 := vec3{X: 0, Y: 15, Z: 0}
	nbs := []beadCellNeighbor{
		{Center: c1, K: 3},
		{Center: c2, K: 3},
		{Center: c3, K: 3},
	}
	target := vec3{X: 0, Y: 0, Z: 100}
	cands := solveBeadCells(nbs, target)
	if len(cands) == 0 {
		t.Fatal("expected discrete candidates for a three-neighbour node")
	}
	for _, c := range cands {
		assertOnLattice(t, c, nbs)
	}
}

func TestSolveBeadCells_FourthNeighborConstraintFiltersCandidates(t *testing.T) {
	c1 := vec3{X: -10, Y: 0, Z: 0}
	c2 := vec3{X: 10, Y: 0, Z: 0}
	c3 := vec3{X: 0, Y: 15, Z: 0}
	// Solve without the 4th constraint first to find a real solution point, then place a
	// 4th neighbour exactly at that point's distance so it must survive filtering, and a
	// second 4th-neighbour placement far enough away that NOTHING survives.
	base := []beadCellNeighbor{{Center: c1, K: 3}, {Center: c2, K: 3}, {Center: c3, K: 3}}
	cands := solveBeadCells(base, vec3{X: 0, Y: 0, Z: 100})
	if len(cands) == 0 {
		t.Fatal("setup: expected a base 3-neighbour solution to build the 4th constraint from")
	}
	p := cands[0]

	// Place c4 exactly kExact*BeadStepR from p (by construction, not by rounding an
	// arbitrary distance) so it is a REAL 4th constraint p satisfies exactly.
	const kExact = 5
	c4 := p.Add(vec3{X: 1, Y: 0, Z: 0}.Scale(kExact * wire.BeadStepR))
	withGoodFourth := append(append([]beadCellNeighbor{}, base...), beadCellNeighbor{Center: c4, K: kExact})
	got := solveBeadCells(withGoodFourth, vec3{X: 0, Y: 0, Z: 100})
	found := false
	for _, c := range got {
		if c.Sub(p).Length() < 1e-6 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the known-good 3-neighbour solution %+v to survive a 4th constraint set exactly at its own distance", p)
	}

	// Now push the 4th neighbour's K far off from what p actually measures — nothing
	// should survive that includes p, and a config with NO surviving candidate is a
	// legitimate (observable) empty result, not a bug in the solver.
	withBadFourth := append(append([]beadCellNeighbor{}, base...), beadCellNeighbor{Center: c4, K: kExact + 50})
	got2 := solveBeadCells(withBadFourth, vec3{X: 0, Y: 0, Z: 100})
	for _, c := range got2 {
		if c.Sub(p).Length() < 1e-6 {
			t.Fatalf("expected p=%+v to be filtered out by a wildly wrong 4th constraint, but it survived", p)
		}
	}
}

func TestSolveBeadCells_NoIntersection_ReturnsNoCandidateForThatCombo(t *testing.T) {
	// Two neighbours too far apart for their K=1 spheres to ever meet, and no
	// neighbouring K makes them meet either (radii capped far below the gap).
	c1 := vec3{X: 0, Y: 0, Z: 0}
	c2 := vec3{X: 10000, Y: 0, Z: 0}
	nbs := []beadCellNeighbor{{Center: c1, K: 1}, {Center: c2, K: 1}}
	got := solveBeadCells(nbs, vec3{X: 5000, Y: 0, Z: 0})
	if len(got) != 0 {
		t.Fatalf("expected no admissible candidates for two neighbours whose spheres cannot meet at any neighbouring K, got %+v", got)
	}
}

func TestSnapToBeadCell_NoCandidateHoldsPosition(t *testing.T) {
	prev := vec3{X: 1, Y: 2, Z: 3}
	c1 := vec3{X: 0, Y: 0, Z: 0}
	c2 := vec3{X: 10000, Y: 0, Z: 0}
	nbs := []beadCellNeighbor{{Center: c1, K: 1}, {Center: c2, K: 1}}
	got := snapToBeadCell(prev, nbs, vec3{X: 5000, Y: 0, Z: 0})
	if got != prev {
		t.Fatalf("expected snapToBeadCell to hold position with no candidate, got %+v want %+v", got, prev)
	}
}

func TestSolveBeadCells_UnchangedConfigurationAlwaysHasACandidate(t *testing.T) {
	// The all-delta-zero branch of the enumeration must always be able to solve: a node
	// already exactly on its lattice, asked to move nowhere, must find at least its own
	// current position among the K=(unchanged) candidates. This is the guarantee that a
	// normal drag always has somewhere to land (this file's header comment).
	nodePos := vec3{X: 3, Y: 4, Z: 5}
	c1 := vec3{X: -20, Y: 0, Z: 0}
	c2 := vec3{X: 20, Y: 10, Z: 0}
	c3 := vec3{X: 0, Y: -15, Z: 8}
	k := func(c vec3) int { return int(math.Round(nodePos.Sub(c).Length() / wire.BeadStepR)) }
	nbs := []beadCellNeighbor{
		{Center: c1, K: k(c1)},
		{Center: c2, K: k(c2)},
		{Center: c3, K: k(c3)},
	}
	got := solveBeadCells(nbs, nodePos)
	if len(got) == 0 {
		t.Fatal("expected the unchanged-K combination to always yield at least one candidate")
	}
}

func TestSolveBeadCells_MovingByOneCellChangesKByOne(t *testing.T) {
	// A one-cell move changes an edge's bead count by exactly 1: moving from K to K+1 (or
	// K-1) toward one neighbour, holding the rest, is exactly what the enumeration steps
	// by construction — pin that the +1/-1 delta is actually reachable and lands exactly
	// one BeadStepR farther/nearer.
	nb := beadCellNeighbor{Center: vec3{}, K: 4}
	p0 := nearestOnSphere(nb.Center, float64(nb.K)*wire.BeadStepR, vec3{X: 1, Y: 0, Z: 0})
	nbPlus := beadCellNeighbor{Center: nb.Center, K: nb.K + 1}
	p1 := nearestOnSphere(nbPlus.Center, float64(nbPlus.K)*wire.BeadStepR, vec3{X: 1, Y: 0, Z: 0})
	if d := math.Abs(p1.Sub(nb.Center).Length() - p0.Sub(nb.Center).Length() - wire.BeadStepR); d > 1e-6 {
		t.Fatalf("expected a one-cell move to change the distance by exactly one BeadStepR, got delta=%v",
			p1.Sub(nb.Center).Length()-p0.Sub(nb.Center).Length())
	}
}
