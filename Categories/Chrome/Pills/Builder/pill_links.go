package Builder

import (
	"math"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const LinksLabel = "Links"

const (
	OpLinkStart = "link-start"
	OpLinkPress = "link-press"
)

func (m Model) allLinks() []Links.Link {
	var out []Links.Link
	for _, id := range m.Nodes {
		out = append(out, m.Reports[id].Links...)
	}
	Links.SortLinks(out)
	return out
}

func LinksPill(m Model) Pill {
	const fit = "0.0"
	var rows [][]Cell
	for _, l := range m.allLinks() {
		act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k} }
		rows = append(rows, []Cell{
			Text(l.Name()),
			Field("", num1(l.Start.Phi), act(OpLinkStart, Links.KPhi), true, true).Fitting(fit),
			Field("", num1(l.Delta.Phi), act(OpLinkPress, Links.KPhi), false, true).Fitting(fit),
			Field("", num1(l.Start.Theta), act(OpLinkStart, Links.KTheta), true, true).Fitting(fit),
			Field("", num1(l.Delta.Theta), act(OpLinkPress, Links.KTheta), false, true).Fitting(fit),
			Field("", num1(l.Start.R), act(OpLinkStart, Links.KR), true, false).Fitting(fit),
			Button("remove", act(OpDropLink, ""), false),
		})
	}
	if len(rows) == 0 {
		rows = [][]Cell{{Text("no links")}}
	} else {
		labels := []Cell{Text(""), Text("φ"), Text("Δφ"), Text("θ"), Text("Δθ"), Text("r")}
		rows = append([][]Cell{labels}, rows...)
	}
	groups := []Group{
		{Title: "links — one per a→b, set at a: its start, and Δ on it (± steps Δ and moves b)", Rows: rows},
		pickGroup(m),
		pickedGroup(m),
		recalcGroup(m),
		landsGroup(m),
	}
	if m.Refused != "" {
		groups = append(groups, Group{Rows: [][]Cell{{Text(m.Refused)}}})
	}
	return Pill{Label: LinksLabel, Groups: groups}
}

func num1(v float64) string {
	v = math.Round(v*10) / 10
	if v == math.Trunc(v) {
		return strconv.Itoa(int(v))
	}
	return strconv.FormatFloat(v, 'f', 1, 64)
}

func recalcGroup(m Model) Group {
	var names []string
	for _, l := range m.allLinks() {
		if !l.Picked {
			names = append(names, l.Name())
		}
	}
	text := "no links chosen for recalculation"
	if len(names) > 0 {
		text = strings.Join(names, "  ")
	}
	return Group{Title: "chosen for vector recalculation — remeasured when either end moves", Rows: [][]Cell{{Text(text)}}}
}

func landsGroup(m Model) Group {
	var rows [][]Cell
	for _, l := range m.allLinks() {
		d := l.Dir(m.ScaleOf(l.To))
		rows = append(rows, []Cell{
			Text(l.Name()),
			Text("start φ " + Num(l.Start.Phi) + "  θ " + Num(l.Start.Theta) + "  r " + Num(l.Start.R)),
			Text("start + Δ φ " + Num(d.Phi) + "  θ " + Num(d.Theta) + "  r " + Num(d.R)),
		})
	}
	if len(rows) == 0 {
		rows = [][]Cell{{Text("no rows")}}
	}
	return Group{Title: "read only: each row's start, and start + Δ — the direction from its first node to its second", Rows: rows}
}
