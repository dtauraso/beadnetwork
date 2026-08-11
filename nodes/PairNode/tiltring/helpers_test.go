package tiltring

// helpers_test.go — shared fixtures and helpers used by more than one of this package's test
// files. See docs/process/testing-shape.md for what a test here may assert. (Was
// PairNode/node_helpers_test.go.)

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

// offBy is how far a count sits from its mode's stop, the short way round the count-ring.
//
// THE MACHINE DOES NOT COMPUTE THIS, and that is the point of it living here. One arrival asks two
// things — am I there, and which way — and neither is a length: Settled is a comparison and Step is
// a subtraction against a quarter turn. The rule used to produce a magnitude because its second
// question scored both neighbours and compared the results; when that went, the magnitude had one
// consumer left, a test for zero.
//
// The tests below still want it, because "each step lands nearer than it started" is a claim about
// a distance and cannot be stated without one. So it is computed here, by the test that needs it,
// and the production rule stays free of a number nothing reads.
func offBy(m Machine, from, arrival *State) int32 {
	stop := m.Stopping()
	if stop.Anywhere {
		return 0
	}
	h := from.R.HalfTurn
	c, _ := from.NearerEndCount(arrival)
	up := ((stop.At(from.R)-c)%h + h) % h
	return min(up, h-up)
}

// steppedTop is a node's new TOP after one step, whichever end the rule actually drove. Step
// names the end it moved and PairNode.Node writes that one and derives the other; this is that
// same conversion, so a test whose subject is the top can go on naming the top.
func steppedTop(m Machine, top, arrival *State) *State {
	moved, atBottom := m.Step(top, arrival)
	if atBottom {
		return moved.Opposite
	}
	return moved
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

// testRing is a 48-point ring: a quarter turn is 12, a half turn 24 — the lattice the live scene
// runs at, so the numbers here read the same as the rows in the probe log.
func testRing() *Ring { return NewRing(48) }

// The two chosen modes, named once here so the tests read as the modes rather than as calls.
// Setting is the zero value and lives in machine.go, since production code names it too.
var (
	perpendicular = MachineFor(tiltvector.TiltMachinePerpendicular)
	parallel      = MachineFor(tiltvector.TiltMachineParallel)
)
