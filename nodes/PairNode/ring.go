package PairNode

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/PairNode/tiltring"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

var defaultRing = tiltring.NewRing(tiltvector.FullTurnPhiIdx)

func (n *Node) ringOf() *tiltring.Ring {
	if n.lattice.Ring == nil {
		return defaultRing
	}
	return n.lattice.Ring
}

func (n *Node) topState() *tiltring.State {
	if n.tilt.Top == nil {
		return n.ringOf().At(0)
	}
	return n.tilt.Top
}

func (n *Node) bottomState() *tiltring.State {
	if n.tilt.Bottom == nil {
		return n.topState().Opposite
	}
	return n.tilt.Bottom
}

func (n *Node) setTop(top *tiltring.State) { n.tilt.Top, n.tilt.Bottom = top, top.Opposite }

func (n *Node) setBottom(bottom *tiltring.State) { n.tilt.Bottom, n.tilt.Top = bottom, bottom.Opposite }

func (n *Node) fromAnotherLattice(received tiltvector.TiltVectorMsg) bool {
	return received.Points != 0 && received.Points != n.ringOf().Points
}

func (n *Node) drainLattice() {
	if n.lattice.LatticeIn == nil {
		return
	}
	select {
	case points := <-n.lattice.LatticeIn:
		n.adoptLattice(points)
	default:
	}
}

func (n *Node) adoptLattice(points int32) {
	if points == n.ringOf().Points {
		return
	}
	keptIdx := n.topState().Idx
	n.lattice.Ring = tiltring.NewRing(points)
	top, unknown := n.lattice.Ring.SeedState(keptIdx)
	n.setTop(top)
	if unknown && n.plumb.Self != nil {
		n.plumb.Self.Breadcrumb("pair-lattice-adopt", fmt.Sprintf(
			"points=%d keptIdx=%d unknown=true loaded=%d", points, keptIdx, top.Idx))
	}
	n.vec.ReceivedPhiIdx = 0
	n.vec.ReceivedSet = false
	n.syncReceivedVector()
	tiltvector.PollRecvVector(n.vec.VectorIn)
	if n.lattice.SyncLatticePoints != nil {
		n.lattice.SyncLatticePoints(points)
	}
	n.syncTiltIndex()
}
