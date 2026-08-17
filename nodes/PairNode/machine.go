package PairNode

import (
	"github.com/dtauraso/wirefold/nodes/PairNode/tiltring"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

func (n *Node) adoptMachine(choice tiltvector.TiltMachine) {
	if n.tilt.Machine != tiltring.Setting {
		return
	}
	n.tilt.Machine = tiltring.MachineFor(choice)
}
