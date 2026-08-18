package PairNode

import (
	"github.com/dtauraso/wirefold/nodes/PairNode/tiltring"
	"github.com/dtauraso/wirefold/tools/topology-vscode/TiltPanel"
)

func (n *Node) stepFromVector(received TiltPanel.TiltVectorMsg) bool {
	ring := n.ringOf()
	ring.ArrivedState(received.PhiIdx)
	top := n.topState().Idx

	offset := n.tilt.Machine.OffsetOn(ring, top, received.PhiIdx)
	if offset == 0 {
		n.rest.restedThisCycle = true
		return true
	}
	n.setTop(ring.At(tiltring.Mod(top+offset, ring.Points)))
	return true
}

func (n *Node) handleVectorCycle(tick int64) {
	received, ok := TiltPanel.PollRecvVector(n.vec.VectorIn)
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

	n.adoptMachine(TiltPanel.TiltMachineParallel)
	if !n.stepFromVector(received) {
		return
	}
	n.reply()
	n.reportRest()

	if n.plumb.Out != nil {
		n.plumb.Out.PlaceDrivenAt(1, tick)
	}
}
