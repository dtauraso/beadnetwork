package PairNode

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

func (n *Node) clock() clock.Clock {
	if n.plumb.Clock == nil {
		return clock.NewRealClock()
	}
	return n.plumb.Clock
}

func (n *Node) openingEmit() {
	wire.TryEmit(n.plumb.EmitGeometry)

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
