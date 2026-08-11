package PairNode

// machine.go — WHICH TILT MACHINE A NODE RUNS: the two functions that answer it. The machine
// itself — the mode, the stopping counts, settled/step/choice — moved to nodes/PairNode/tiltring
// (pure math, no node, no goroutine); what stays here is specific to a NODE deciding which mode
// to run and when that decision is allowed to change.

import (
	"github.com/dtauraso/wirefold/nodes/PairNode/tiltring"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

// machineForGap reads WHICH MACHINE THE PAIR IS FOR out of the gap between the two tilts.
//
// The arrival is the partner's coplanar NORMAL, a quarter turn off its tilt, so backing that
// quarter out gives the partner's own tilt and the gap between the two is a real measurement
// rather than one node's angle against zero — which is what makes this work whether the user
// tilted one node or both.
//
//	the gap is a quarter turn  ->  perpendicular
//	anything else (acute)      ->  parallel
func (n *Node) machineForGap(arrival *tiltring.State) tiltvector.TiltMachine {
	partnerTilt := arrival.Quarter.Opposite // arrival + three quarters = arrival − a quarter
	if n.topState().AngleLength(partnerTilt) == n.ringOf().QuarterTurn {
		return tiltvector.TiltMachinePerpendicular
	}
	return tiltvector.TiltMachineParallel
}

// adoptMachine sets which mode of the tilt machine this node runs. It is the ONE writer of that
// field outside clear(), and the mapping from the pair-wide name to the mode is
// tiltring.MachineFor — the naming lives in Wiring so both ends can say it to each other, and the
// stopping counts live in tiltring, which is the only place that knows what any of them means on
// the ring.
//
// The choice STICKS: a node already running one keeps it, so a second choice crossing the pair
// — or one arriving at an end that has already made its own — cannot switch it mid-run. Only a
// reset clears it, and the next click after that makes a new one.
func (n *Node) adoptMachine(choice tiltvector.TiltMachine) {
	if n.tilt.Machine != tiltring.Setting {
		return
	}
	n.tilt.Machine = tiltring.MachineFor(choice)
}
