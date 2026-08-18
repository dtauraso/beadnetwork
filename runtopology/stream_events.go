package runtopology

import (
	"github.com/dtauraso/wirefold/nodes/rowevent"

	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
)

func toStreamEvents(events []rowevent.RowEvent) []SF.StreamEvent {
	if len(events) == 0 {
		return nil
	}
	out := make([]SF.StreamEvent, len(events))
	for i, e := range events {
		out[i] = SF.StreamEvent{
			Kind:          SF.KindID(e.Kind),
			NodeRow:       e.NodeRow,
			PortRow:       e.PortRow,
			TargetRow:     e.TargetRow,
			TargetPortRow: e.TargetPortRow,
			EdgeRow:       e.EdgeRow,
			Slot:          e.Slot,
			Value:         e.Value,
			Bead:          uint32(e.Bead),
			BeadSteps:     float32(e.BeadSteps),
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
