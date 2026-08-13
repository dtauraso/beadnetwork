package edgemover

import (
	"encoding/binary"
	"fmt"

	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/nodes/spatial"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

type vec3 = spatial.Vec3

// edgeBeads is this edge's own in-flight beads, as world positions on its
// own segment. They arrive from the goroutine that steps the wire; nothing
// here recomputes where a bead is.
func (m *EdgeMover) edgeBeads() []SF.EdgeBead {
	beads := make([]SF.EdgeBead, 0, len(m.lastBeadRows))
	for _, r := range m.lastBeadRows {
		beads = append(beads, SF.EdgeBead{
			X: float32(r.X), Y: float32(r.Y), Z: float32(r.Z), Value: int32(r.Val),
		})
	}
	return beads
}

// breadcrumb records one control event on this edge's own stream.
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

// noteBeadCount reports the FIRST time each bead is seen, by generation, and
// how far along it already was at that moment.
//
// Reporting on a count change instead was measuring the wrong thing: it named
// whichever bead sat at index 0, so an old bead most of the way across looked
// exactly like a new one starting there. What matters is the earliest
// position a bead is ever observed at — if that is not near zero, the bead is
// being sampled for the first time long after it was placed, and it will
// appear on screen to start partway down the edge.
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
			"src=%s dst=%s gen=%d firstSeenAlong=%.3f inFlight=%d steps=%d",
			m.srcID, m.dstID, b.Gen, p.Dot(d)/l2, len(rows), m.steps))
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

	// The source row rides the frame itself, not just its events: it is
	// what says whose edge this is, and the renderer needs that whether or
	// not this edge has a destination draining events this tick.
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
