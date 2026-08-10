package PairNode

// machine_adoption_test.go — how a node picks its machine from the gap it sees, holds it once
// adopted, and drives the end it measured. See docs/process/testing-shape.md for what a test here
// may assert.

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

func TestMachineIsReadFromTheGapNotFromOneTilt(t *testing.T) {
	r := testRing()
	// What arrives is the partner's NORMAL, a quarter turn off its tilt. So for a chosen pair of
	// tilts, the arrival this node sees is partnerTilt + a quarter.
	arrivalFor := func(partnerTilt int32) *tiltState { return r.at(partnerTilt).quarter }

	cases := []struct {
		name             string
		ownTilt, partner int32
		want             tiltvector.TiltMachine
	}{
		// The case the live test runs: one node clicked to a quarter turn, the other left at 0.
		{"a quarter turn apart", 12, 0, tiltvector.TiltMachinePerpendicular},
		{"a quarter turn apart, the other way", 0, 12, tiltvector.TiltMachinePerpendicular},
		// BOTH tilted, still a quarter apart — the gap is the pair's, not one node's angle
		// against zero, which is what reading a single tilt got wrong.
		{"both tilted, a quarter apart", 20, 8, tiltvector.TiltMachinePerpendicular},
		{"the same direction", 7, 7, tiltvector.TiltMachineParallel},
		// One step short of a quarter turn — the gap a click reads mid-setup. Deciding here is
		// what locked a pair to the wrong machine while the tilt was still on its way.
		{"one step off a quarter turn", 11, 0, tiltvector.TiltMachineParallel},
		{"an ordinary acute gap", 3, 0, tiltvector.TiltMachineParallel},
	}
	for _, c := range cases {
		n := &Node{lattice: latticeState{Ring: r}, tilt: tiltHeld{Top: r.at(c.ownTilt)}}
		if got := n.machineForGap(arrivalFor(c.partner)); got != c.want {
			t.Errorf("%s (tilts %d and %d): chose %v, want %v", c.name, c.ownTilt, c.partner, got, c.want)
		}
	}
}

// TestASettingNodeHoldsWhereverItStands is the behaviour that used to be an `if Machine == nil`
// exemption in stepFromVector and is now a consequence of the setting mode's home set: a node
// still deciding which machine it runs is already halted at every angle length, so no arrival can
// move it. A zero-value Node is in that mode without being put there.
func TestASettingNodeHoldsWhereverItStands(t *testing.T) {
	r := testRing()
	for sep := int32(0); sep < r.points; sep++ {
		n := &Node{lattice: latticeState{Ring: r}, tilt: tiltHeld{Top: r.at(5)}}
		if n.tilt.Machine != setting {
			t.Fatalf("a fresh node is not in the setting mode: %v", n.tilt.Machine)
		}
		n.stepFromVector(tiltvector.TiltVectorMsg{ThetaIdx: sep})
		if n.tilt.Top != r.at(5) {
			t.Errorf("arrival at angle length %d moved a node that is still being set up: top %d, want 5",
				sep, n.tilt.Top.idx)
		}
	}
}

// TestASettingNodeTellsTheOtherEndNothing: the mode's own choice is TiltMachineNone, so the wire
// says "no choice carried" without outgoingVector testing for one.
func TestASettingNodeTellsTheOtherEndNothing(t *testing.T) {
	r := testRing()
	n := &Node{lattice: latticeState{Ring: r}, tilt: tiltHeld{Top: r.at(0)}}
	if got := n.outgoingVector().Machine; got != tiltvector.TiltMachineNone {
		t.Errorf("a node still being set up announced machine %v, want none", got)
	}
	n.adoptMachine(tiltvector.TiltMachineParallel)
	if got := n.outgoingVector().Machine; got != tiltvector.TiltMachineParallel {
		t.Errorf("after adopting, announced %v, want parallel", got)
	}
}

func TestAdoptedMachineSticksUntilCleared(t *testing.T) {
	r := testRing()
	n := &Node{lattice: latticeState{Ring: r}, tilt: tiltHeld{Top: r.at(0)}}

	n.adoptMachine(tiltvector.TiltMachinePerpendicular)
	if n.tilt.Machine != perpendicular {
		t.Fatalf("adopt did not take: running %v", n.tilt.Machine)
	}
	// A second choice — a click landing mid-run, or the partner's own answer arriving — must not
	// switch a running machine. Re-deciding on a jitter click switched a started perpendicular
	// pair to parallel one step after START.
	n.adoptMachine(tiltvector.TiltMachineParallel)
	if n.tilt.Machine != perpendicular {
		t.Errorf("a later choice switched a running machine: now %v", n.tilt.Machine)
	}
	// RESET is the one thing that releases it.
	n.clear()
	if n.tilt.Machine != setting {
		t.Errorf("reset left a machine running: %v", n.tilt.Machine)
	}
}

func TestNoMachineMeansNoMovement(t *testing.T) {
	r := testRing()
	n := &Node{lattice: latticeState{Ring: r}, tilt: tiltHeld{Top: r.at(5)}}
	// Before any start, and after a reset, an arrival moves nothing. The node used to infer a
	// machine from the arrival here, and that inference always answered perpendicular, because
	// closing on the arrival IS the perpendicular measure.
	before := n.topState()
	for _, sep := range []int32{0, 1, 7, 12, 24, 40} {
		n.stepFromVector(tiltvector.TiltVectorMsg{ThetaIdx: sep, Points: r.points})
		if got := n.topState(); got != before {
			t.Fatalf("arrival at %d moved the tilt with no machine running: %d -> %d",
				sep, before.idx, got.idx)
		}
	}
}

// TestAnUpdateDrivesTheEndItMeasured is the claim behind storing both ends: an arrival is measured
// at whichever end is nearer, and THAT is the end the update moves — the page's two halves, each
// moving the end it read.
//
// Two things are asserted on every (tilt, arrival) pair of both lattices, on ONE node, with no
// exchange running:
//
//	the driven end is the measured end     step's bit is nearerEndCount's bit
//	the two ends stay a half turn apart    whichever was written, the other followed
//
// The second is the reason setTop/setBottom exist rather than two assignments at each site. It
// cannot drift into a delivery test: nothing here sends anything, and one goroutine does all of it.
func TestAnUpdateDrivesTheEndItMeasured(t *testing.T) {
	for _, points := range []int32{24, 48} {
		r := newRing(points)
		for _, m := range []tiltMachine{perpendicular, parallel} {
			for tilt := int32(0); tilt < points; tilt++ {
				for arr := int32(0); arr < points; arr++ {
					a, before := r.at(arr), r.at(tilt)
					n := &Node{lattice: latticeState{Ring: r}}
					n.setTop(before)
					n.tilt.Machine = m
					if m.settled(before, a) {
						continue
					}
					moved, atBottom := m.step(before, a)
					if _, measuredAtBottom := before.nearerEndCount(a); atBottom != measuredAtBottom {
						t.Fatalf("points=%d %v t=%d a=%d: measured at bottom=%v but drove bottom=%v",
							points, m, tilt, arr, measuredAtBottom, atBottom)
					}
					if atBottom {
						n.setBottom(moved)
					} else {
						n.setTop(moved)
					}
					if n.tilt.Top.opposite != n.tilt.Bottom {
						t.Fatalf("points=%d %v t=%d a=%d: top %d and bottom %d are not a half turn apart",
							points, m, tilt, arr, n.tilt.Top.idx, n.tilt.Bottom.idx)
					}
					// And the line moved by exactly one slot, in the direction the count
					// went — the behaviour storing the second end was not allowed to change.
					if n.tilt.Top != before.next && n.tilt.Top != before.prev {
						t.Fatalf("points=%d %v t=%d a=%d: top went %d -> %d, not one slot",
							points, m, tilt, arr, before.idx, n.tilt.Top.idx)
					}
				}
			}
		}
	}
}
