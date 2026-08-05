// parallel_chain_offset_test.go — two nodes that point at each other must not draw their
// chains on the same line.
//
// The property that matters is not "an offset exists" but that the two ends, deciding
// ALONE on their own goroutines with no message between them, land on OPPOSITE sides. That
// is the one thing a local decision can get wrong, and the reason canonical id order exists.
package Wiring

import (
	"math"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// Two nodes at the SAME radius about the scene centre, 90 degrees apart — the ordinary
// case. A radial pair (one node directly between the other and the scene centre) is the
// degenerate one, covered separately below.
var (
	sceneOrigin = vec3{X: 0, Y: 0, Z: 0}
	centerA     = vec3{X: 100, Y: 0, Z: 0}
	centerB     = vec3{X: 0, Y: 0, Z: 100}
)

func TestMutualPairOffsetsToOppositeSides(t *testing.T) {
	a, okA := parallelChainOffset("1", "2", centerA, centerB, sceneOrigin)
	b, okB := parallelChainOffset("2", "1", centerB, centerA, sceneOrigin)
	if !okA || !okB {
		t.Fatalf("offset not resolvable for a well-separated pair: okA=%v okB=%v", okA, okB)
	}
	// Opposite sides means the two offsets are negatives of each other. Deriving the
	// perpendicular from each node's OWN outgoing direction would instead make them EQUAL
	// (both directions negate, so both perpendiculars negate), putting the two chains back
	// on the same line — the exact bug this ordering prevents.
	if math.Abs(a.X+b.X) > 1e-9 || math.Abs(a.Y+b.Y) > 1e-9 || math.Abs(a.Z+b.Z) > 1e-9 {
		t.Fatalf("the two ends must offset to OPPOSITE sides: got %v and %v (sum is not zero)", a, b)
	}
}

func TestOffsetIsPerpendicularToTheEdge(t *testing.T) {
	off, ok := parallelChainOffset("1", "2", centerA, centerB, sceneOrigin)
	if !ok {
		t.Fatal("offset not resolvable")
	}
	edgeDir := centerB.Sub(centerA).Normalize()
	if dot := off.Dot(edgeDir); math.Abs(dot) > 1e-9 {
		t.Fatalf("offset must be perpendicular to the edge, got dot=%v (offset %v)", dot, off)
	}
}

// The two chains land exactly TWO bead steps apart — on the lattice, not at a tuned gap.
// One step each way, so a full bead-sized gap sits between them; at half this (the chains
// exactly touching) the pair read as a single thick wire rather than as two edges.
func TestTheTwoChainsSitTwoBeadStepsApart(t *testing.T) {
	a, _ := parallelChainOffset("1", "2", centerA, centerB, sceneOrigin)
	b, _ := parallelChainOffset("2", "1", centerB, centerA, sceneOrigin)
	if got, want := a.Sub(b).Length(), 2*wire.BeadStepR; math.Abs(got-want) > 1e-9 {
		t.Fatalf("chain separation = %v, want two bead steps (%v)", got, want)
	}
}

// Node ids are NUMBERS that are strings only because they are directory names. A string
// compare puts "10" before "2", which would hand both ends of that pair the same sign and
// collapse them back onto one line — the failure this orders numerically to avoid.
func TestOrderingIsNumericNotLexicographic(t *testing.T) {
	a, _ := parallelChainOffset("2", "10", centerA, centerB, sceneOrigin)
	b, _ := parallelChainOffset("10", "2", centerB, centerA, sceneOrigin)
	if math.Abs(a.X+b.X) > 1e-9 || math.Abs(a.Y+b.Y) > 1e-9 || math.Abs(a.Z+b.Z) > 1e-9 {
		t.Fatalf("ids 2 and 10 must order numerically: got %v and %v, which do not oppose", a, b)
	}
}

// A RADIAL edge — one node directly between the other and the scene centre — runs ALONG
// the pole, so cross(pole, dir) vanishes. Every perpendicular then lies in the ring plane
// and any is admissible, but it must still be a real vector rather than a silent zero that
// re-collapses the two chains onto one line.
func TestRadialEdgeStillSeparates(t *testing.T) {
	near := vec3{X: 50, Y: 0, Z: 0}
	far := vec3{X: 150, Y: 0, Z: 0}
	a, okA := parallelChainOffset("1", "2", near, far, sceneOrigin)
	b, okB := parallelChainOffset("2", "1", far, near, sceneOrigin)
	if !okA || !okB {
		t.Fatalf("a radial edge must still resolve an offset: okA=%v okB=%v", okA, okB)
	}
	if a.Length() < 1e-9 {
		t.Fatal("a radial edge resolved a ZERO offset: the two chains would coincide")
	}
	if math.Abs(a.X+b.X) > 1e-9 || math.Abs(a.Y+b.Y) > 1e-9 || math.Abs(a.Z+b.Z) > 1e-9 {
		t.Fatalf("even in the degenerate case both ends must oppose: got %v and %v", a, b)
	}
}

// The offset must lie IN the node's ring plane — perpendicular to the inward pole — so the
// chain stays coplanar with the tori it runs between.
func TestOffsetLiesInTheRingPlane(t *testing.T) {
	off, ok := parallelChainOffset("1", "2", centerA, centerB, sceneOrigin)
	if !ok {
		t.Fatal("offset not resolvable")
	}
	// Canonical order takes the pole from node "1" (centerA): inward, toward the scene centre.
	pole := sceneOrigin.Sub(centerA).Normalize()
	if dot := off.Dot(pole); math.Abs(dot) > 1e-9 {
		t.Fatalf("offset must be perpendicular to the inward pole (in the ring plane), got dot=%v", dot)
	}
}

// Coincident centres have no direction to be perpendicular to; report that rather than
// dividing by ~0 and emitting a NaN position into the buffer.
func TestCoincidentCentresReportNotResolvable(t *testing.T) {
	if _, ok := parallelChainOffset("1", "2", centerA, centerA, sceneOrigin); ok {
		t.Fatal("coincident centres must report not-resolvable, not produce an offset")
	}
}
