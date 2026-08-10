package PairNode

// node_helpers_test.go — shared fixtures and helpers used by more than one of this
// package's test files. See docs/process/testing-shape.md for what a test here may assert.

import (
	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// offBy is how far a count sits from its mode's stop, the short way round the count-ring.
//
// THE MACHINE DOES NOT COMPUTE THIS, and that is the point of it living here. One arrival asks two
// things — am I there, and which way — and neither is a length: settled is a comparison and step is
// a subtraction against a quarter turn. The rule used to produce a magnitude because its second
// question scored both neighbours and compared the results; when that went, the magnitude had one
// consumer left, a test for zero.
//
// The tests below still want it, because "each step lands nearer than it started" is a claim about
// a distance and cannot be stated without one. So it is computed here, by the test that needs it,
// and the production rule stays free of a number nothing reads.
func offBy(m tiltMachine, from, arrival *tiltState) int32 {
	stop := m.stopping()
	if stop.anywhere {
		return 0
	}
	h := from.ring.halfTurn
	c, _ := from.nearerEndCount(arrival)
	up := ((stop.at(from.ring)-c)%h + h) % h
	return min(up, h-up)
}

// steppedTop is this node's new TOP after one step, whichever end the rule actually drove. step
// names the end it moved (machine.go) and node.go writes that one and derives the other; this is
// that same conversion, so a test whose subject is the top can go on naming the top.
func steppedTop(m tiltMachine, top, arrival *tiltState) *tiltState {
	moved, atBottom := m.step(top, arrival)
	if atBottom {
		return moved.opposite
	}
	return moved
}

// a 48-point ring: a quarter turn is 12, a half turn 24 — the lattice the live scene runs at,
// so the numbers here read the same as the rows in the probe log.
func testRing() *ring { return newRing(48) }

// The two chosen modes, named once here so the tests read as the modes rather than as calls.
// `setting` is the zero value and lives in machine.go, since production code names it too.
var (
	perpendicular = machineFor(Wiring.TiltMachinePerpendicular)
	parallel      = machineFor(Wiring.TiltMachineParallel)
)
