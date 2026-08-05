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

var (
	centerA = vec3{X: 0, Y: 0, Z: 0}
	centerB = vec3{X: 100, Y: 0, Z: 0}
)

func TestMutualPairOffsetsToOppositeSides(t *testing.T) {
	a, okA := parallelChainOffset("1", "2", centerA, centerB)
	b, okB := parallelChainOffset("2", "1", centerB, centerA)
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
	off, ok := parallelChainOffset("1", "2", centerA, centerB)
	if !ok {
		t.Fatal("offset not resolvable")
	}
	edgeDir := centerB.Sub(centerA).Normalize()
	if dot := off.Dot(edgeDir); math.Abs(dot) > 1e-9 {
		t.Fatalf("offset must be perpendicular to the edge, got dot=%v (offset %v)", dot, off)
	}
}

// The two chains land exactly one bead step apart — on the lattice, not at a tuned gap.
func TestTheTwoChainsSitOneBeadStepApart(t *testing.T) {
	a, _ := parallelChainOffset("1", "2", centerA, centerB)
	b, _ := parallelChainOffset("2", "1", centerB, centerA)
	if got, want := a.Sub(b).Length(), wire.BeadStepR; math.Abs(got-want) > 1e-9 {
		t.Fatalf("chain separation = %v, want one bead step (%v)", got, want)
	}
}

// Node ids are NUMBERS that are strings only because they are directory names. A string
// compare puts "10" before "2", which would hand both ends of that pair the same sign and
// collapse them back onto one line — the failure this orders numerically to avoid.
func TestOrderingIsNumericNotLexicographic(t *testing.T) {
	a, _ := parallelChainOffset("2", "10", centerA, centerB)
	b, _ := parallelChainOffset("10", "2", centerB, centerA)
	if math.Abs(a.X+b.X) > 1e-9 || math.Abs(a.Y+b.Y) > 1e-9 || math.Abs(a.Z+b.Z) > 1e-9 {
		t.Fatalf("ids 2 and 10 must order numerically: got %v and %v, which do not oppose", a, b)
	}
}

// A vertical edge is the one direction the first perpendicular axis cannot produce; it must
// fall through to the second rather than returning a zero vector that silently re-collapses
// the two chains.
func TestVerticalEdgeStillSeparates(t *testing.T) {
	off, ok := parallelChainOffset("1", "2", vec3{}, vec3{X: 0, Y: 100, Z: 0})
	if !ok {
		t.Fatal("a vertical edge must still resolve an offset")
	}
	if off.Length() < 1e-9 {
		t.Fatal("a vertical edge resolved a ZERO offset: the two chains would coincide")
	}
}

// Coincident centres have no direction to be perpendicular to; report that rather than
// dividing by ~0 and emitting a NaN position into the buffer.
func TestCoincidentCentresReportNotResolvable(t *testing.T) {
	if _, ok := parallelChainOffset("1", "2", centerA, centerA); ok {
		t.Fatal("coincident centres must report not-resolvable, not produce an offset")
	}
}
