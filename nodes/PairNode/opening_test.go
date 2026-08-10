package PairNode

// opening_test.go — a full opening exchange, played by calling the same decision functions the
// live node calls, in the order its own loop calls them, with no goroutines and no channels.
// See docs/process/testing-shape.md for what a test here may assert.

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// openingOutcome is what one opening of the exchange came to: which machine the pair took up,
// how many rounds it took, and how far apart the two TILTS ended.
type openingOutcome struct {
	machine Wiring.TiltMachine
	rounds  int
	tiltGap int32
	settled bool
}

// runOpening plays one exchange from a starting pair of tilts, calling the same decision
// functions the live node calls, in the order its own loop calls them.
//
// It runs no goroutines and opens no channels: it is the RULE being iterated, not the network
// being exercised. The order is the real one — node 1 opens (START belongs to id 1), node 2
// answers, and from then on each end reads the other's last normal and steps once.
func runOpening(r *ring, tiltA, tiltB int32) openingOutcome {
	a := &Node{lattice: latticeState{Ring: r}, tilt: tiltHeld{Top: r.at(tiltA)}}
	b := &Node{lattice: latticeState{Ring: r}, tilt: tiltHeld{Top: r.at(tiltB)}}

	// One end reads an arrival exactly as handleVectorCycle does: adopt what the sender says it
	// is running, then choose from the gap if still running nothing, then step.
	read := func(n *Node, arrival *tiltState, senderRuns Wiring.TiltMachine) bool {
		n.adoptMachine(senderRuns)
		if n.tilt.Machine == setting {
			n.adoptMachine(n.machineForGap(arrival))
		}
		before := n.topState()
		// Written the way stepFromVector writes it: the end that was measured is the end
		// that moves, and the other is read off its opposite in the same statement.
		if !n.tilt.Machine.settled(before, arrival) {
			if moved, atBottom := n.tilt.Machine.step(before, arrival); atBottom {
				n.setBottom(moved)
			} else {
				n.setTop(moved)
			}
		}
		return n.topState() != before
	}
	runs := func(n *Node) Wiring.TiltMachine { return n.tilt.Machine.choice() }
	normal := func(n *Node) *tiltState { return n.topState().quarter }

	out := openingOutcome{}
	// A full turn of rounds is far more than any settling walk needs — the longest is a quarter
	// turn's worth of steps — so reaching the cap means it never settled.
	for out.rounds = 1; out.rounds <= int(r.points); out.rounds++ {
		movedB := read(b, normal(a), runs(a))
		movedA := read(a, normal(b), runs(b))
		if !movedA && !movedB {
			out.settled = true
			break
		}
	}
	out.machine = runs(b)
	out.tiltGap = a.topState().angleLength(b.topState())
	return out
}

func TestEveryOpeningSettlesOnTheMachineItChose(t *testing.T) {
	r := newRing(24)
	// The counts this sweep produces are what docs/pair-node/math/math.html reports, so run it with
	// -v after changing the rule and put the new numbers on the page rather than leaving the
	// old ones there.
	var perp, par, worst int
	var same, reversed int
	for tiltA := int32(0); tiltA < r.points; tiltA++ {
		for tiltB := int32(0); tiltB < r.points; tiltB++ {
			got := runOpening(r, tiltA, tiltB)
			switch {
			case got.machine == Wiring.TiltMachinePerpendicular:
				perp++
			case got.tiltGap == 0:
				par, same = par+1, same+1
			default:
				par, reversed = par+1, reversed+1
			}
			if got.rounds > worst {
				worst = got.rounds
			}
			if !got.settled {
				t.Errorf("opening (%d,%d) never settled: still moving after %d rounds",
					tiltA, tiltB, got.rounds)
				continue
			}
			// Settling means the two TILTS ended in the relationship the chosen machine is for.
			// PARALLEL ACCEPTS TWO OF THEM: the same direction, and a half turn apart, which is
			// the same LINE with the arrows reversed. The halt cannot tell those apart, because
			// angle length folds the long way round into the short one and a reversed partner
			// lands the same distance off. Perpendicular has the one, a quarter turn.
			ok := got.tiltGap == r.quarterTurn
			if got.machine == Wiring.TiltMachineParallel {
				ok = got.tiltGap == 0 || got.tiltGap == r.halfTurn
			}
			if !ok {
				t.Errorf("opening (%d,%d) chose %v but settled with the tilts %d apart",
					tiltA, tiltB, got.machine, got.tiltGap)
			}
		}
	}
	t.Logf("openings=%d perpendicular=%d parallel=%d (same direction=%d, reversed=%d) worst rounds=%d",
		r.points*r.points, perp, par, same, reversed, worst)
}
