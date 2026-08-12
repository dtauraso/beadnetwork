package edgemover

import (
	"encoding/binary"
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

func (m *EdgeMover) writeStreamFrame(tick int64, events []wire.RowEvent) {
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

	seg := nodegeom.EdgeSegment(m.srcGeom, m.dstGeom)
	selected := m.selected
	if m.dest != nil {

		nodeRow, targetRow := int32(-1), int32(-1)
		if m.nodeRowFor != nil {
			if r, ok := m.nodeRowFor(m.srcID); ok {
				nodeRow = r
			}
			if r, ok := m.nodeRowFor(m.dstID); ok {
				targetRow = r
			}
		}
		for _, pe := range m.dest.DrainPendingEvents() {
			events = append(events, wire.RowEvent{
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
		selected, m.edgeID, events)
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))

	_, _ = m.streamOut.Write(hdr[:])
	_, _ = m.streamOut.Write(frame)
}
