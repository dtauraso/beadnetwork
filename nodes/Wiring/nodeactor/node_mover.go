package nodeactor

import (
	"context"
)

type NodeMover struct {
	geom *NodeGeometry

	speedCh chan float64
}

func NewNodeMover(geom *NodeGeometry) *NodeMover {
	return &NodeMover{geom: geom}
}

func (m *NodeMover) SetSpeedCh(ch chan float64) {
	m.speedCh = ch
}

func (m *NodeMover) Run(ctx context.Context) {
	g := m.geom
	g.clocks.CopyClockSrc()

	if g.tr != nil {
		g.emitGeometry()
	}
	for {
		g.clocks.ApplySpeed(m.speedCh)

		for {
			progressed, cancelled := g.msg.DrainPending(ctx, g.handle)
			if cancelled {
				return
			}
			if !progressed {
				break
			}
		}

		outTick := g.clocks.Tick()
		g.outs.DriveOutWires(ctx, outTick)

		g.msg.FlushPending()

		g.writeStreamFrame(nil)
		if err := g.clocks.SleepCycle(ctx); err != nil {
			return
		}
	}
}
