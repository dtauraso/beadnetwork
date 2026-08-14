package nodeactor

import (
	"context"
)

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
			progressed, cancelled := g.msg.DrainPending(ctx, g.take)
			if cancelled {
				return
			}
			if !progressed {
				break
			}
		}

		g.msg.FlushPending()

		g.writeStreamFrame(g.drainSelfEvents())
		if err := g.clocks.SleepPulse(ctx); err != nil {
			return
		}
	}
}
