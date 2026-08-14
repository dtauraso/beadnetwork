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
		g.clocks.ApplySpeed(g.anim.SpeedCh())

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

		tick := g.clocks.Tick()

		g.anim.driveOutWires(ctx, tick)

		g.writeStreamFrame(g.drainSelfEvents())
		g.writeOutEdgeFrames(tick)
		g.writeInteriorFrames()

		if err := g.clocks.SleepPulse(ctx); err != nil {
			return
		}
	}
}
