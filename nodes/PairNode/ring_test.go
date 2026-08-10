package PairNode

// ring_test.go — the ring's own load-time validation. See docs/process/testing-shape.md for what
// a test here may assert.

import (
	"testing"
)

func TestARingMustHaveAWholeQuarterTurn(t *testing.T) {
	// A quarter turn has to be a whole number of states or the coplanar normal and the
	// perpendicular halt name nothing on the ring.
	defer func() {
		if recover() == nil {
			t.Error("a 10-point lattice has no quarter turn and must not build")
		}
	}()
	newRing(10)
}
