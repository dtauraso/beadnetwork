package PairNode

import (
	"github.com/dtauraso/wirefold/nodes/PairNode/tiltring"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

func (n *Node) machineForGap(arrival *tiltring.State) tiltvector.TiltMachine {
	partnerTilt := arrival.Quarter.Opposite
	if n.topState().AngleLength(partnerTilt) == n.ringOf().QuarterTurn {
		return tiltvector.TiltMachinePerpendicular
	}
	return tiltvector.TiltMachineParallel
}

func (n *Node) adoptMachine(choice tiltvector.TiltMachine) {
	if n.tilt.Machine != tiltring.Setting {
		return
	}
	n.tilt.Machine = tiltring.MachineFor(choice)
}
