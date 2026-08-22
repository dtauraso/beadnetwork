package PairNode

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/wirefold/src/Clock"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/NodeKinds/PairNode/tiltring"
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

		TiltPanel.SendVectorLatestNonBlocking(n.vec.VectorOut, TiltPanel.TiltVectorMsg{Reset: true})
		return false
	}
	if edit.Start {

		if n.plumb.PairID != 1 {
			return false
		}

		TiltPanel.SendVectorLatestNonBlocking(n.vec.VectorOut, n.outgoingVector())
		return true
	}

	ring := n.ringOf()
	step := int32(-1)
	if edit.Up {
		step = +1
	}
	n.setTop(ring.At(tiltring.Mod(n.topState().Idx+step, ring.Points)))

	return false
}

func (n *Node) clear() {
	n.setTop(n.ringOf().At(0))

	n.tilt.Machine = tiltring.Unset()
	n.syncTiltIndex()

	n.vec.ReceivedPhiIdx = tiltring.Sent(n.topState().Idx, n.ringOf().Points)
	n.vec.ReceivedSet = false
	n.syncReceivedVector()

	n.rest = restCounters{}
	if n.plumb.Self != nil {
		n.plumb.Self.SetRoundsToParallel(0, 0)
	}
	TiltPanel.PollRecvVector(n.vec.VectorIn)
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
