package edgemover

import (
	"encoding/binary"
	"fmt"

	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/framegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/nodes/spatial"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

type vec3 = spatial.Vec3

func (m *EdgeMover) edgeBeads() []SF.EdgeBead {
	axisPhi, axisTheta := framegeom.TorusDefaultAxisAngles()
	beads := make([]SF.EdgeBead, 0, len(m.lastBeadRows))
	for _, r := range m.lastBeadRows {
		pos := vec3{X: r.X, Y: r.Y, Z: r.Z}
		beads = append(beads, SF.EdgeBead{
			X: float32(r.X), Y: float32(r.Y), Z: float32(r.Z), Value: int32(r.Val),
			RingMatrix: framegeom.RingInstanceMatrixColumnMajor(
				pos, nodegeom.ShadingParamBeadRadius, axisPhi, axisTheta),
		})
	}
	return beads
}

func (m *EdgeMover) breadcrumb(label uint8, name, value string) {
	if m.tr == nil {
		return
	}
	m.tr.Breadcrumb(name, m.edgeID, "", value)
	m.streamBreadcrumb(label, value)
}

func (m *EdgeMover) streamBreadcrumb(label uint8, value string) {
	if !m.streamOut.Ok() || m.buildFrame == nil {
		return
	}
	m.writeStreamFrame(m.clk.Tick(), []rowevent.RowEvent{{
		Kind: T.KindBreadcrumb, Label: label, Debug: 1,
		EdgeRow: m.edgeRow, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, Slot: -1,
		Text: value,
	}})
}

func (m *EdgeMover) noteBeadCount(rows []wire.LiveBeadRow) {
	seg := edgegeom.EdgeSegment(m.srcGeom, m.dstGeom)
	d := seg.End.Sub(seg.Start)
	l2 := d.Dot(d)
	if l2 <= 0 {
		return
	}
	for _, b := range rows {
		if m.seenBeadGens[b.Gen] {
			continue
		}
		if m.seenBeadGens == nil {
			m.seenBeadGens = map[uint64]bool{}
		}
		m.seenBeadGens[b.Gen] = true
		p := vec3{X: b.X, Y: b.Y, Z: b.Z}.Sub(seg.Start)
		m.breadcrumb(T.BreadcrumbEdgeBeads, "edge-beads", fmt.Sprintf(
			"src=%s dst=%s gen=%d firstSeenAlong=%.3f beadSteps=%d age=%.1f edgeSteps=%d inFlight=%d",
			m.srcID, m.dstID, b.Gen, p.Dot(d)/l2, b.Steps, b.Age, m.steps, len(rows)))
	}
}

func (m *EdgeMover) writeStreamFrame(tick int64, events []rowevent.RowEvent) {
	if !m.streamOut.Ok() || m.buildFrame == nil {
		return
	}

	for _, e := range events {
		if e.EdgeRow != -1 && e.EdgeRow != m.edgeRow {
			panic(fmt.Sprintf(
				"edgeMover.writeStreamFrame: edge %q (row %d) is carrying a %s event for edge row %d on its OWN dedicated stream — EdgeRow is the ownership claim on an edge stream (NodeRow is a reference to the source node)",
				m.edgeID, m.edgeRow, e.Kind, e.EdgeRow))
		}
	}

	seg := edgegeom.EdgeSegment(m.srcGeom, m.dstGeom)

	nodeRow, targetRow := int32(-1), int32(-1)
	if m.nodeRowFor != nil {
		if r, ok := m.nodeRowFor(m.srcID); ok {
			nodeRow = r
		}
		if r, ok := m.nodeRowFor(m.dstID); ok {
			targetRow = r
		}
	}
	if m.dest != nil {
		for _, pe := range m.dest.DrainPendingEvents() {
			events = append(events, rowevent.RowEvent{
				Kind: pe.Kind, NodeRow: nodeRow, PortRow: -1,
				TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				Value: int32(pe.Value), Bead: pe.Gen,
				X: pe.X, Y: pe.Y, Z: pe.Z, F: pe.T,
			})
		}

		for _, ev := range m.dest.DrainBreadcrumbEvents() {
			ev.NodeRow = nodeRow
			ev.PortRow = -1
			ev.TargetRow = targetRow
			events = append(events, ev)
		}
	}
	frame := m.buildFrame(uint32(tick),
		float32(seg.Start.X), float32(seg.Start.Y), float32(seg.Start.Z),
		float32(seg.End.X), float32(seg.End.Y), float32(seg.End.Z),
		nodeRow, m.edgeID, m.edgeBeads(), events)
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))

	_, _ = m.streamOut.Write(hdr[:])
	_, _ = m.streamOut.Write(frame)
}
