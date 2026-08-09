package main

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"

	B "github.com/dtauraso/wirefold/Buffer"
)

// toStreamEvents converts a nodeMover/edgeMover/interiorStream goroutine's own
// row-resolved events (Wiring.RowEvent, string kind — kept Buffer-independent there) into
// Buffer.StreamEvent (numeric kind, via Buffer.KindID) for packing into that SAME
// goroutine's own frame's trailing EVENTS section (memory/feedback_no_single_writer_bridge.md).
// Pure value conversion — no shared state, safe to call from any owner goroutine.
func toStreamEvents(events []wire.RowEvent) []B.StreamEvent {
	if len(events) == 0 {
		return nil
	}
	out := make([]B.StreamEvent, len(events))
	for i, e := range events {
		out[i] = B.StreamEvent{
			Kind:          B.KindID(e.Kind),
			NodeRow:       e.NodeRow,
			PortRow:       e.PortRow,
			TargetRow:     e.TargetRow,
			TargetPortRow: e.TargetPortRow,
			EdgeRow:       e.EdgeRow,
			Slot:          e.Slot,
			Value:         e.Value,
			Bead:          uint32(e.Bead),
			BeadSteps:     float32(e.BeadSteps),
			SimLatencyMs:  float32(e.SimLatencyMs),
			X:             float32(e.X),
			Y:             float32(e.Y),
			Z:             float32(e.Z),
			F:             float32(e.F),
			Label:         e.Label,
			Debug:         e.Debug,
			Text:          e.Text,
		}
	}
	return out
}
