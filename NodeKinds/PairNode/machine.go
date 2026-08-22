package PairNode

import (
	"github.com/dtauraso/wirefold/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/NodeKinds/PairNode/tiltring"
)

func (n *Node) adoptMachine(choice TiltPanel.TiltMachine) {
	if n.tilt.Machine != tiltring.Unset() {
		return
	}
	n.tilt.Machine = tiltring.MachineFor(choice)
}
