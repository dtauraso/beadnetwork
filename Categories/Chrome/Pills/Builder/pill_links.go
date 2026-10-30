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
	OpUnfold    = "unfold"
)

func (m Model) allLinks() []Links.Link {
	var out []Links.Link
	for _, id := range m.Nodes {
		out = append(out, m.Reports[id].Links...)
	}
	Links.SortLinks(out)
	return out
}

type linkCol struct {
	label  string
	unfold string
	value  func(Links.Link) float64
	op, k  string
	typed  bool
	steps  bool
	picked bool
}

func (m Model) linkCols() []linkCol {
	chain := func(cols ...linkCol) []linkCol {
		out := []linkCol{cols[0]}
		for i := 1; i < len(cols) && m.Unfold[cols[i-1].unfold]; i++ {
			out = append(out, cols[i])
		}
		return out
	}
	var cols []linkCol
	cols = append(cols, chain(
		linkCol{"φ", "phi", func(l Links.Link) float64 { return l.Start.Phi }, OpLinkStart, Links.KPhi, true, true, false},
		linkCol{"Δφ", "dphi", func(l Links.Link) float64 { return l.Delta.Phi }, OpLinkPress, Links.KPhi, false, true, true},
		linkCol{"nΔφ", "", func(l Links.Link) float64 { return l.Increment.Phi }, OpIncrement, Links.KPhi, true, true, true},
	)...)
	cols = append(cols, chain(
		linkCol{"θ", "theta", func(l Links.Link) float64 { return l.Start.Theta }, OpLinkStart, Links.KTheta, true, true, false},
		linkCol{"Δθ", "dtheta", func(l Links.Link) float64 { return l.Delta.Theta }, OpLinkPress, Links.KTheta, false, true, true},
		linkCol{"nΔθ", "", func(l Links.Link) float64 { return l.Increment.Theta }, OpIncrement, Links.KTheta, true, true, true},
	)...)
	return append(cols, chain(
		linkCol{"r", "r", func(l Links.Link) float64 { return l.Start.R }, OpLinkStart, Links.KR, true, true, false},
		linkCol{"Δr", "dr", func(l Links.Link) float64 { return l.DeltaR }, OpLinkPress, Links.KR, false, true, false},
		linkCol{"nΔr", "", func(l Links.Link) float64 { return l.IncR }, OpIncrement, Links.KR, true, true, false},
	)...)
}

func (c linkCol) head(m Model) Cell {
	if c.unfold == "" {
		return c.mark(Text(c.label)).AboveLastDigit()
	}
	open := m.Unfold[c.unfold]
	mark := " ▸"
	if open {
		mark = " ▾"
	}
	fold := Act{Op: OpUnfold, K: c.unfold}
	if c.op == OpLinkPress {
		lock := map[string]int{Links.KPhi: m.Step.LockAll.Phi, Links.KTheta: m.Step.LockAll.Theta, Links.KR: m.Step.LockAll.R}[c.k]
		return c.mark(LitText(c.label+mark, fold, open)).AboveLastDigit().Stepping(Act{Op: OpAllStep, K: c.k}, lock)
	}
	return c.mark(LitText(c.label+mark, fold, open)).AboveLastDigit()
}

func (c linkCol) mark(cell Cell) Cell {
	if c.picked {
		return cell.Outline(true)
	}
	return cell
}

func LinksPill(m Model) Pill {
	const fit = "0.0"
	cols := m.linkCols()
	var rows [][]Cell
	for _, l := range m.allLinks() {
		act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k} }
		pick := OpPick
		if l.Picked {
			pick = OpUnpick
		}
		row := []Cell{Text(l.Name()).Acting(act(pick, "")).Outline(l.Picked)}
		for _, c := range cols {
			row = append(row, Field("", num1(c.value(l)), act(c.op, c.k), c.typed, c.steps).Fitting(fit).Outline(c.picked && l.Picked))
		}
		rows = append(rows, append(row, Button("×", act(OpDropLink, ""), false)))
	}
	if len(rows) == 0 {
		rows = [][]Cell{{Text("no links")}}
	} else {
		labels := []Cell{Text("both").Outline(true).Stepping(Act{Op: OpAllBoth}, 0)}
		labels[0].Lit = m.Step.LockAll != (Links.Lock{})
		for _, c := range cols {
			labels = append(labels, c.head(m))
		}
		rows = append([][]Cell{labels}, rows...)
	}
	groups := []Group{
		{Title: "links — one per a→b, set at a: its start, and Δ on it (± steps Δ and moves b); a picked link's name is outlined, and clicking a name picks or unpicks it", Rows: rows, Slides: true},
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
