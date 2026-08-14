package PairNode

import (
	"github.com/dtauraso/wirefold/nodes/PairNode/tiltring"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	"github.com/dtauraso/wirefold/nodes/clock"
)

func (n *Node) drainTiltEdit(clk clock.Clock) {
	if n.tilt.TiltEditIn == nil {
		return
	}
	select {
	case edit := <-n.tilt.TiltEditIn:
		placeBead := n.applyTiltEdit(edit)
		n.syncTiltIndex()
		if placeBead && n.plumb.Out != nil {
			n.plumb.Out.PlaceDrivenAt(1, clk.Tick())
		}
	default:
	}
}

func (n *Node) applyTiltEdit(edit movemsg.TiltEditMsg) (placeBead bool) {
	if edit.Reset {
		n.clear()

		tiltvector.SendVectorLatestNonBlocking(n.vec.VectorOut, tiltvector.TiltVectorMsg{Reset: true})
		return false
	}
	if edit.Start {

		if n.plumb.PairID != 1 {
			return false
		}

		tiltvector.SendVectorLatestNonBlocking(n.vec.VectorOut, n.outgoingVector())
		return true
	}

	if edit.Up {
		n.setTop(n.topState().Next)
	} else {
		n.setTop(n.topState().Prev)
	}

	return false
}

func (n *Node) clear() {
	n.setTop(n.ringOf().At(0))

	n.tilt.Machine = tiltring.Setting
	n.syncTiltIndex()
	n.vec.ReceivedPhiIdx = 0
	n.vec.ReceivedSet = false
	n.syncReceivedVector()

	n.rest = restCounters{}
	if n.plumb.Self != nil {
		n.plumb.Self.SetRoundsToParallel(0, 0)
	}
	tiltvector.PollRecvVector(n.vec.VectorIn)
	n.drainIn()
	if n.plumb.ClearOutBeads != nil {
		n.plumb.ClearOutBeads()
	}
}

func (n *Node) drainIn() {
	if n.plumb.In == nil {
		return
	}
	for {
		if _, ok := n.plumb.In.PollRecv(); !ok {
			return
		}
	}
}
