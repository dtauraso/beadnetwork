package PairNode

import (
	clock "github.com/dtauraso/wirefold/src/Clock"
	"github.com/dtauraso/wirefold/src/NodeKinds/nodeapi"
)

func (n *Node) clock() clock.Clock {
	if n.plumb.Clock == nil {
		return clock.NewRealClock()
	}
	return n.plumb.Clock
}

func (n *Node) openingEmit() {
	nodeapi.TryEmit(n.plumb.EmitGeometry)

	n.plumb.Self.EmitGeometryOnce()
	n.syncTiltIndex()
}

func (n *Node) paceOnBeadArrival() {
	if _, ok := n.plumb.In.PollRecv(); ok {
		if n.plumb.Fire != nil {
			n.plumb.Fire()
		}
	}
}
