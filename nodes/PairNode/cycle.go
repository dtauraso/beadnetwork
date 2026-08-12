package PairNode

import (
	"github.com/dtauraso/wirefold/nodes/PairNode/tiltring"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

func (n *Node) stepFromVector(received tiltvector.TiltVectorMsg) bool {
	arrival := n.ringOf().ArrivedState(received.ThetaIdx)
	before := n.topState()

	if !n.tilt.Machine.Settled(before, arrival) {
		if moved, atBottom := n.tilt.Machine.Step(before, arrival); atBottom {
			n.setBottom(moved)
		} else {
			n.setTop(moved)
		}
	} else {

		n.rest.restedThisCycle = true
	}
	return true
}

func (n *Node) handleVectorCycle(tick int64) {
	received, ok := tiltvector.PollRecvVector(n.vec.VectorIn)
	if !ok {
		return
	}
	n.rest.countArrival()

	if received.Reset {
		n.clear()
		return
	}

	n.adoptMachine(received.Machine)
	if n.fromAnotherLattice(received) {
		return
	}
	n.recordReceived(received)

	if n.tilt.Machine == tiltring.Setting {
		n.adoptMachine(n.machineForGap(n.ringOf().ArrivedState(received.ThetaIdx)))
	}
	if !n.stepFromVector(received) {
		return
	}
	n.reply()
	n.reportRest()

	if n.plumb.Out != nil {
		n.plumb.Out.PlaceDrivenAt(1, tick)
	}
}
