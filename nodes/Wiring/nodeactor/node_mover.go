package nodeactor

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/clock"
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
	if g.clocks.clockSrc != nil {
		g.clocks.clk = g.clocks.clockSrc.Copy()
	}

	if g.tr != nil {
		g.emitGeometry()
	}
	for {
		clock.ApplySpeedNonBlocking(g.clocks.clk, m.speedCh)

		for {
			progressed := false
			select {
			case <-ctx.Done():
				return
			case msg := <-g.msg.extIn:
				g.handle(msg)
				if msg.TestDone != nil {
					close(msg.TestDone)
				}
				progressed = true
			default:
			}
			for _, ch := range g.msg.neighborIn {
				select {
				case msg := <-ch:
					g.handle(msg)
					if msg.TestDone != nil {
						close(msg.TestDone)
					}
					progressed = true
				default:
				}
			}
			if !progressed {
				break
			}
		}

		outTick := g.clocks.clk.Tick()
		for _, pw := range g.outs.outWires {
			pw.DriveOneCycle(ctx, outTick)
		}

		g.msg.flushPending()

		g.writeStreamFrame(nil)
		if err := g.clocks.clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}
