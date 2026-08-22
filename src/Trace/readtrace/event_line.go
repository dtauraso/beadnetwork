package main

import trace "github.com/dtauraso/wirefold/src/Trace"

func LineOf(e trace.RowEvent, nowMs int64) string {
	return Logfmt(fieldsOf(e, nowMs))
}

func fieldsOf(e trace.RowEvent, nowMs int64) []Field {
	head := []Field{{"ts_ms", nowMs}, {"src", "go"}, {"kind", e.Kind}}

	switch e.Kind {
	case trace.KindRecv:
		return append(head,
			Field{"node", trace.NameOf(e.NodeRow)}, Field{"port", ""}, Field{"value", e.Value})

	case trace.KindFire:
		return append(head, Field{"node", trace.NameOf(e.NodeRow)})

	case trace.KindSend:
		out := append(head,
			Field{"node", trace.NameOf(e.NodeRow)}, Field{"port", ""}, Field{"value", e.Value})
		if e.BeadSteps != 0 {
			out = append(out, Field{"beadSteps", e.BeadSteps})
			if target := trace.NameOf(e.TargetRow); target != "" {
				out = append(out, Field{"target", target})
			}
		}
		return out

	case trace.KindArrive:
		out := append(head,
			Field{"node", trace.NameOf(e.NodeRow)}, Field{"port", ""}, Field{"value", e.Value})
		if e.Bead != 0 {
			out = append(out, Field{"bead", e.Bead})
		}
		return out

	case trace.KindBreadcrumb:
		out := []Field{
			{"ts_ms", nowMs}, {"src", "go"}, {"kind", e.Kind},
			{"label", trace.LabelOf(e.Label)}, {"debug", e.Debug == 1},
			{"node", trace.NameOf(e.NodeRow)}, {"port", ""}, {"value", e.Value},
			{"x", e.X}, {"y", e.Y}, {"z", e.Z},
			{"nodeRow", e.NodeRow}, {"portRow", e.PortRow},
			{"targetRow", e.TargetRow}, {"targetPortRow", e.TargetPortRow},
			{"edgeRow", e.EdgeRow}, {"slot", e.Slot},
		}
		if target := trace.NameOf(e.TargetRow); target != "" {
			out = append(out, Field{"target", target})
		}
		if e.Text != "" {
			out = append(out, Field{"text", e.Text})
		}
		return out
	}
	return head
}
