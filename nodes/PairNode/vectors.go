package PairNode

import (
	"github.com/dtauraso/wirefold/nodes/PairNode/tiltring"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

func (n *Node) bottomTilt() tiltvector.TiltVectorMsg {
	ring := n.ringOf()
	return tiltvector.TiltVectorMsg{PhiIdx: tiltring.Bottom(n.topState().Idx, ring.Points)}
}

func (n *Node) coplanarNormal() tiltvector.TiltVectorMsg {
	ring := n.ringOf()
	return tiltvector.TiltVectorMsg{PhiIdx: tiltring.Sent(n.topState().Idx, ring.Points)}
}

func (n *Node) syncTiltIndex() {
	if n.tilt.SyncTiltIndex == nil {
		return
	}
	norm := n.coplanarNormal()
	bottom := n.bottomTilt()
	n.tilt.SyncTiltIndex(n.topState().Idx, norm.PhiIdx, bottom.PhiIdx)
}

func (n *Node) syncReceivedVector() {
	if n.vec.SyncReceivedVector == nil {
		return
	}
	n.vec.SyncReceivedVector(n.vec.ReceivedPhiIdx, n.vec.ReceivedSet)
}

func (n *Node) recordReceived(received tiltvector.TiltVectorMsg) {
	n.vec.ReceivedPhiIdx = received.PhiIdx
	n.vec.ReceivedSet = true
	n.syncReceivedVector()
}

func (n *Node) reply() {
	n.syncTiltIndex()
	n.rest.msgsSinceOpen++
	tiltvector.SendVectorLatestNonBlocking(n.vec.VectorOut, n.outgoingVector())
}

func (n *Node) outgoingVector() tiltvector.TiltVectorMsg {
	v := n.coplanarNormal()

	v.Points = n.ringOf().Points

	v.Machine = n.tilt.Machine.Choice()
	return v
}
