package PairNode

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/src/NodeKinds/PairNode/tiltring"
)

func (n *Node) adoptMachine(choice TiltPanel.TiltMachine) {
	if n.tilt.Machine != tiltring.Unset() {
		return
	}
	n.tilt.Machine = tiltring.MachineFor(choice)
}
