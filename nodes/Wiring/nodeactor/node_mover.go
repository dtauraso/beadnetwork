package nodeactor

import (
	"context"
)

// NodeMover is the geometry peer: it drains node-move messages and writes
// the node stream, free-running at one raw pulse regardless of bead speed.
// It never calls SleepCycle and never touches the animation peer's outs —
// bead speed must not be able to reach dragging through this loop.
type NodeMover struct {
	geom *NodeGeometry
}

func NewNodeMover(geom *NodeGeometry) *NodeMover {
	return &NodeMover{geom: geom}
}

func (m *NodeMover) Run(ctx context.Context) {
	g := m.geom
	g.clocks.CopyClockSrc()

	if g.tr != nil {
		g.emitGeometry()
	}
	for {
		for {
			progressed, cancelled := g.msg.DrainPending(ctx, g.handle)
			if cancelled {
				return
			}
			if !progressed {
				break
			}
		}

		g.msg.FlushPending()

		g.writeStreamFrame(nil)
		if err := g.clocks.SleepPulse(ctx); err != nil {
			return
		}
	}
}
