package edgemover

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/clock"
)

func (m *EdgeMover) Run(ctx context.Context) {

	if m.clockSrc != nil {
		m.clk = m.clockSrc.Copy()
	}

	if m.tr != nil {
		m.recomputeGeometry()
	}
	for {

	drain:
		for {
			select {
			case <-ctx.Done():
				return
			case sp := <-m.speedCh:

				if rc, ok := m.clk.(*clock.RealClock); ok {
					rc.SetSpeed(sp)
				}
			case rows := <-m.beadRowsIn:
				m.lastBeadRows = rows
				m.noteBeadCount(rows)
				m.writeStreamFrame(m.clk.Tick(), nil)
			case msg := <-m.extIn:
				m.handle(msg)
				if msg.TestDone != nil {
					close(msg.TestDone)
				}
			case msg := <-m.srcIn:
				m.handle(msg)
				if msg.TestDone != nil {
					close(msg.TestDone)
				}
			case msg := <-m.dstIn:
				m.handle(msg)
				if msg.TestDone != nil {
					close(msg.TestDone)
				}
			default:
				break drain
			}
		}
		if m.dest != nil {

			m.writeStreamFrame(m.clk.Tick(), nil)
		}
		if err := m.clk.SleepPulse(ctx); err != nil {
			return
		}
	}
}
