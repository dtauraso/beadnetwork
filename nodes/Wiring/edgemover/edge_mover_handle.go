package edgemover

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/Trace"
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

	// How many steps it takes to cross is a fact about THIS edge — its two
	// endpoints and the two node kinds whose tori it runs between — so the
	// edge works it out. It used to be handed over by the source node's
	// chain layout, and when that went the count went to zero with it,
	// which is a bead that never leaves the end it was placed at.
	//
	// The distance is CENTRE to CENTRE, not the drawn segment: EdgeStepCount
	// takes the two tori off itself, and seg has already had them taken off.
	// Passing seg's length subtracts them twice, which shortens the crossing
	// by an amount that depends on the two node KINDS — so every edge runs
	// at its own wrong speed instead of one wrong speed.
	centerDist, _, distOK := edgegeom.EdgeCenterDistAndDir(
		nodegeom.NodeWorldPos(m.srcGeom), nodegeom.NodeWorldPos(m.dstGeom))
	if distOK {
		m.steps = edgegeom.EdgeStepCount(centerDist, m.srcGeom.Kind, m.dstGeom.Kind)
	}

	if m.out != nil {
		m.out.PublishSegment(seg.Start, seg.End)
		m.out.PublishSteps(m.steps)
	}

	// Which way this edge runs and how long it takes to cross, recorded
	// whenever either endpoint moves. It is the segment the beads are placed
	// against, so a bead in the wrong place is either this being wrong or
	// the placement disagreeing with it.
	m.breadcrumb(T.BreadcrumbEdgeGeom, "edge-geom", fmt.Sprintf(
		"src=%s dst=%s steps=%d start=(%.2f,%.2f,%.2f) end=(%.2f,%.2f,%.2f) len=%.2f",
		m.srcID, m.dstID, m.steps,
		seg.Start.X, seg.Start.Y, seg.Start.Z, seg.End.X, seg.End.Y, seg.End.Z,
		seg.End.Sub(seg.Start).Length()))

	if m.dest != nil {
		m.dest.ReviseInFlightGeometry(m.clk.Tick(), m.steps, seg)
	}

	m.writeStreamFrame(m.clk.Tick(), []rowevent.RowEvent{{
		Kind: T.KindGeometry, EdgeRow: m.edgeRow,
		NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1,
	}})
}
