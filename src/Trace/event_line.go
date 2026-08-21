package trace

func LineOf(e RowEvent, nowMs int64) string {
	return Logfmt(fieldsOf(e, nowMs))
}

func fieldsOf(e RowEvent, nowMs int64) []Field {
	head := []Field{{"ts_ms", nowMs}, {"src", "go"}, {"kind", e.Kind}}

	switch e.Kind {
	case KindRecv:
		return append(head,
			Field{"node", NameOf(e.NodeRow)}, Field{"port", ""}, Field{"value", e.Value})

	case KindFire:
		return append(head, Field{"node", NameOf(e.NodeRow)})

	case KindSend:
		out := append(head,
			Field{"node", NameOf(e.NodeRow)}, Field{"port", ""}, Field{"value", e.Value})
		if e.BeadSteps != 0 {
			out = append(out, Field{"beadSteps", e.BeadSteps})
			if target := NameOf(e.TargetRow); target != "" {
				out = append(out, Field{"target", target})
			}
		}
		return out

	case KindArrive:
		out := append(head,
			Field{"node", NameOf(e.NodeRow)}, Field{"port", ""}, Field{"value", e.Value})
		if e.Bead != 0 {
			out = append(out, Field{"bead", e.Bead})
		}
		return out

	case KindBreadcrumb:
		out := []Field{
			{"ts_ms", nowMs}, {"src", "go"}, {"kind", e.Kind},
			{"label", LabelOf(e.Label)}, {"debug", e.Debug == 1},
			{"node", NameOf(e.NodeRow)}, {"port", ""}, {"value", e.Value},
			{"x", e.X}, {"y", e.Y}, {"z", e.Z},
			{"nodeRow", e.NodeRow}, {"portRow", e.PortRow},
			{"targetRow", e.TargetRow}, {"targetPortRow", e.TargetPortRow},
			{"edgeRow", e.EdgeRow}, {"slot", e.Slot},
		}
		if target := NameOf(e.TargetRow); target != "" {
			out = append(out, Field{"target", target})
		}
		if e.Text != "" {
			out = append(out, Field{"text", e.Text})
		}
		return out
	}
	return head
}
