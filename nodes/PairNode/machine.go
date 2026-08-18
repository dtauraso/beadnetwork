package PairNode

import (
	"github.com/dtauraso/wirefold/nodes/PairNode/tiltring"
	"github.com/dtauraso/wirefold/tools/topology-vscode/TiltPanel"
)

func (n *Node) adoptMachine(choice TiltPanel.TiltMachine) {
	if n.tilt.Machine != tiltring.Unset() {
		return
	}
	n.tilt.Machine = tiltring.MachineFor(choice)
}
