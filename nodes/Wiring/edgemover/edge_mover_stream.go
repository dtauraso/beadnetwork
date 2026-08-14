package edgemover

import (
	"encoding/binary"
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/Trace"
)

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

	nodeRow := int32(-1)
	if m.nodeRowFor != nil {
		if r, ok := m.nodeRowFor(m.srcID); ok {
			nodeRow = r
		}
	}
	frame := m.buildFrame(uint32(tick),
		float32(seg.Start.X), float32(seg.Start.Y), float32(seg.Start.Z),
		float32(seg.End.X), float32(seg.End.Y), float32(seg.End.Z),
		nodeRow, m.edgeID, events)
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))

	_, _ = m.streamOut.Write(hdr[:])
	_, _ = m.streamOut.Write(frame)
}
