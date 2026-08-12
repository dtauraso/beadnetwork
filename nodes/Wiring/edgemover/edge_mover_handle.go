package edgemover

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

func (m *EdgeMover) handle(msg movemsg.Msg) {
	if msg.Kind == movemsg.KindSelect {
		if msg.Bool {
			m.selected = 1
		} else {
			m.selected = 0
		}
		return
	}
	if msg.Kind == movemsg.KindCenter {

		if msg.Center == nil {
			return
		}
		switch msg.NodeID {
		case m.srcID:
			nodegeom.SetNodeWorld(&m.srcGeom, *msg.Center)
		case m.dstID:
			nodegeom.SetNodeWorld(&m.dstGeom, *msg.Center)
		default:
			return
		}
		m.recomputeGeometry()
		return
	}
	if msg.Kind == movemsg.KindCenters {

		moved := false
		if c, ok := msg.Centers[m.srcID]; ok {
			nodegeom.SetNodeWorld(&m.srcGeom, c)
			moved = true
		}
		if c, ok := msg.Centers[m.dstID]; ok {
			nodegeom.SetNodeWorld(&m.dstGeom, c)
			moved = true
		}
		if moved {
			m.recomputeGeometry()
		}
		return
	}

	_ = msg
}

func (m *EdgeMover) recomputeGeometry() {
	seg := nodegeom.EdgeSegment(m.srcGeom, m.dstGeom)

	if m.out != nil {
		m.out.PublishSegment(seg.Start, seg.End)
	}

	if m.dest != nil {
		m.dest.ReviseInFlightGeometry(m.clk.Tick(), m.steps, seg)
	}

	m.writeStreamFrame(m.clk.Tick(), []wire.RowEvent{{
		Kind: T.KindGeometry, EdgeRow: m.edgeRow,
		NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1,
	}})
}
