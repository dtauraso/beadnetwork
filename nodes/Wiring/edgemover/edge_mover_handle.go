package edgemover

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

func (m *EdgeMover) handle(msg movemsg.Msg) {
	if msg.Kind == movemsg.KindSelect {
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
	seg := edgegeom.EdgeSegment(m.srcGeom, m.dstGeom)

	m.updateDeltaFromEndpoints()
	m.persistDelta()

	centerDist, _, distOK := edgegeom.EdgeCenterDistAndDir(
		nodegeom.NodeWorldPos(m.srcGeom), nodegeom.NodeWorldPos(m.dstGeom))
	if distOK {
		m.steps = edgegeom.EdgeStepCount(centerDist, m.srcGeom.Kind, m.dstGeom.Kind)
	}

	if m.out != nil {
		m.out.PublishSegment(seg.Start, seg.End)
		m.out.PublishSteps(m.steps)
	}

	m.breadcrumb("edge-geom", fmt.Sprintf(
		"src=%s dst=%s steps=%d start=(%.2f,%.2f,%.2f) end=(%.2f,%.2f,%.2f) len=%.2f",
		m.srcID, m.dstID, m.steps,
		seg.Start.X, seg.Start.Y, seg.Start.Z, seg.End.X, seg.End.Y, seg.End.Z,
		seg.End.Sub(seg.Start).Length()))

	if m.dest != nil {
		m.dest.ReviseInFlightGeometry(m.clk.Tick(), m.steps, seg)
	}
}
