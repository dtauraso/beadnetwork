package PairNode

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

func (n *Node) bottomTilt() tiltvector.TiltVectorMsg {
	return tiltvector.TiltVectorMsg{ThetaIdx: n.bottomState().Idx}
}

func (n *Node) coplanarNormal() tiltvector.TiltVectorMsg {
	return tiltvector.TiltVectorMsg{ThetaIdx: n.topState().Quarter.Idx}
}

func (n *Node) syncTiltIndex() {
	if n.tilt.SyncTiltIndex == nil {
		return
	}
	norm := n.coplanarNormal()
	bottom := n.bottomTilt()
	n.tilt.SyncTiltIndex(n.topState().Idx, norm.ThetaIdx, bottom.ThetaIdx)
}

func (n *Node) syncReceivedVector() {
	if n.vec.SyncReceivedVector == nil {
		return
	}
	n.vec.SyncReceivedVector(n.vec.ReceivedThetaIdx, n.vec.ReceivedSet)
}

func (n *Node) recordReceived(received tiltvector.TiltVectorMsg) {
	n.vec.ReceivedThetaIdx = received.ThetaIdx
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
