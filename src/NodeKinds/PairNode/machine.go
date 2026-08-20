package PairNode

import (
	"github.com/dtauraso/wirefold/src/NodeKinds/PairNode/tiltring"
	"github.com/dtauraso/wirefold/src/TiltPanel"
)

func (n *Node) adoptMachine(choice TiltPanel.TiltMachine) {
	if n.tilt.Machine != tiltring.Unset() {
		return
	}
	n.tilt.Machine = tiltring.MachineFor(choice)
}
