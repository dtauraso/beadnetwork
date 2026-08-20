package PairNode

import (
	"github.com/dtauraso/wirefold/src/NodeKinds/PairNode/tiltring"
	"github.com/dtauraso/wirefold/src/Chrome/TiltPanel"
)

func (n *Node) coplanarNormal() TiltPanel.TiltVectorMsg {
	ring := n.ringOf()
	return TiltPanel.TiltVectorMsg{PhiIdx: tiltring.Sent(n.topState().Idx, ring.Points)}
}

func (n *Node) syncTiltIndex() {
	if n.tilt.SyncTiltIndex == nil {
		return
	}
	n.tilt.SyncTiltIndex(n.topState().Idx)
}

func (n *Node) syncReceivedVector() {
	if n.vec.SyncReceivedVector == nil {
		return
	}
	n.vec.SyncReceivedVector(n.vec.ReceivedPhiIdx, n.vec.ReceivedSet)
}

func (n *Node) recordReceived(received TiltPanel.TiltVectorMsg) {
	n.vec.ReceivedPhiIdx = received.PhiIdx
	n.vec.ReceivedSet = true
	n.syncReceivedVector()
}

func (n *Node) reply() {
	n.syncTiltIndex()
	n.rest.msgsSinceOpen++
	TiltPanel.SendVectorLatestNonBlocking(n.vec.VectorOut, n.outgoingVector())
}

func (n *Node) outgoingVector() TiltPanel.TiltVectorMsg {
	v := n.coplanarNormal()

	v.Points = n.ringOf().Points

	v.Machine = n.tilt.Machine.Choice()
	return v
}
