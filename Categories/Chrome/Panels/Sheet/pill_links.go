package Sheet

import "github.com/dtauraso/beadnetwork/Categories/Node/Links"

const LinksLabel = "Links"

const (
	OpLinkStart = "link-start"
	OpLinkPress = "link-press"
)

func LinksPill(m Model) Pill {
	var rows [][]Cell
	for _, id := range m.Nodes {
		s := m.ScaleOf(id)
		for _, l := range m.Reports[id].Links {
			act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k} }
			d := l.Dir(s)
			rows = append(rows, []Cell{
				Text(l.Name()),
				Field("φ", Num(l.Start.Phi), act(OpLinkStart, Links.KPhi), true, true),
				Field("Δφ", Num(l.Delta.Phi), act(OpLinkPress, Links.KPhi), false, true),
				Field("θ", Num(l.Start.Theta), act(OpLinkStart, Links.KTheta), true, true),
				Field("Δθ", Num(l.Delta.Theta), act(OpLinkPress, Links.KTheta), false, true),
				Field("r", Num(l.Start.R), act(OpLinkStart, Links.KR), true, false),
				Text("→ φ " + Num(d.Phi) + "  θ " + Num(d.Theta)),
			})
		}
	}
	if len(rows) == 0 {
		rows = [][]Cell{{Text("no links")}}
	}
	groups := []Group{{Title: "links — start, and Δ on it (± steps Δ and moves the target)", Rows: rows}}
	if m.Refused != "" {
		groups = append(groups, Group{Rows: [][]Cell{{Text(m.Refused)}}})
	}
	return Pill{Label: LinksLabel, Groups: groups}
}
