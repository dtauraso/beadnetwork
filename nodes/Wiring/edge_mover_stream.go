// edge_mover_stream.go — edgeMover's OWN per-fd stream frame write. See edge_mover.go
// for the actor's held state and edge_mover_run.go for the per-goroutine loop that calls
// this every cycle.

package Wiring

import (
	"encoding/binary"
	"fmt"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// writeStreamFrame packs and writes this edge's combined per-fd frame (edge fields +
// this wire's currently live in-flight beads) to its OWN dedicated fd (streamOut). No-op
// when streamOut is nil (the fallback — see its doc comment) or buildFrame was never
// injected (bare test construction). Called only by this edgeMover's own goroutine
// (recomputeGeometry and run's per-cycle loop), reading m.dest's
// live bead state via LiveBeadRows (same single-goroutine-ownership contract PacedWire's
// other methods rely on).
func (m *edgeMover) writeStreamFrame(tick int64, events []wire.RowEvent) {
	if !m.streamOut.Ok() || m.buildFrame == nil {
		return
	}
	// INVARIANT: same per-goroutine bridge rule the nodeMover twin asserts, but the
	// ownership column is EDGEROW here, not NodeRow — do not copy that condition over.
	// This mover's own event sets EdgeRow: m.edgeRow with NodeRow deliberately -1
	// (recomputeGeometry, edge_mover.go), while the events appended BELOW from
	// DrainPendingEvents/DrainBreadcrumbEvents carry NodeRow as a REFERENCE to the
	// source node with EdgeRow: -1. So NodeRow says nothing about ownership on an
	// edge stream, and only the CALLER-SUPPLIED slice is checkable — hence this runs
	// before those appends.
	// -1 is allowed: it is the "no claim" sentinel, not another edge's row. What this
	// forbids is one edge carrying a DIFFERENT edge's row out on its own stream.
	for _, e := range events {
		if e.EdgeRow != -1 && e.EdgeRow != m.edgeRow {
			panic(fmt.Sprintf(
				"edgeMover.writeStreamFrame: edge %q (row %d) is carrying a %s event for edge row %d on its OWN dedicated stream — EdgeRow is the ownership claim on an edge stream (NodeRow is a reference to the source node)",
				m.edgeID, m.edgeRow, e.Kind, e.EdgeRow))
		}
	}
	// Segment endpoints for this frame's own Edge row (SX..EZ) — no port row to
	// resolve any more (docs/channels-not-ports.md), so this is a plain re-derive
	// from the held endpoint geoms, same computation recomputeGeometry uses.
	seg := edgeSegment(m.srcGeom, m.dstGeom)
	selected := m.selected
	if m.dest != nil {
		// NO live-bead read here. This runs on the EDGE goroutine, but the wire is now
		// stepped by its SOURCE NODE's goroutine (nodeMover.run), so reading pw.inflight
		// from here would break the single-goroutine ownership LiveBeadRows/stepAll depend
		// on. Nothing needs it either: the transit bead is not drawn any more — the
		// animation is the LIT bead on the source node's own chain, which that node
		// computes on its own goroutine (docs/beads-are-the-edge.md).
		// Drain this wire's own OWN-goroutine-recorded Position/Arrive events, resolved
		// to rows here (nodeRowFor — the SAME resolver this frame's own edge columns
		// above just used), and fold them in alongside any caller-supplied events (e.g.
		// a Geometry event from recomputeGeometry). PortRow is -1: there is no port row
		// left to reference.
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
		// This wire's own "wire-send-buffer-full" breadcrumbs, buffered on
		// breadcrumbCh from the source node's goroutine (PacedWire.Send) and
		// resolved to rows here, on this edgeMover's own goroutine — mirrors
		// DrainPendingEvents just above.
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
	// Fire-and-forget, same reasoning throughout this bridge: no delivery
	// guarantee on this channel, errors ignored.
	_, _ = m.streamOut.Write(hdr[:])
	_, _ = m.streamOut.Write(frame)
}
