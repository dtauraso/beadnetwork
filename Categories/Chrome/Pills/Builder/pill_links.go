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

func (m Model) linksFrom(id string) []Links.Link {
	var out []Links.Link
	for _, l := range m.allLinks() {
		if l.From == id {
			out = append(out, l)
		}
	}
	return out
}

type linkCol struct {
	label  string
	value  func(Links.Link) float64
	op, k  string
	typed  bool
	picked bool
}

var linkCols = []linkCol{
	{"φ", func(l Links.Link) float64 { return l.Start.Phi }, OpLinkStart, Links.KPhi, true, false},
	{"Δφ", func(l Links.Link) float64 { return float64(l.Delta.Phi) }, OpLinkPress, Links.KPhi, false, true},
	{"nΔφ", func(l Links.Link) float64 { return float64(l.Increment.Phi) }, OpIncrement, Links.KPhi, true, true},
	{"θ", func(l Links.Link) float64 { return l.Start.Theta }, OpLinkStart, Links.KTheta, true, false},
	{"Δθ", func(l Links.Link) float64 { return float64(l.Delta.Theta) }, OpLinkPress, Links.KTheta, false, true},
	{"nΔθ", func(l Links.Link) float64 { return float64(l.Increment.Theta) }, OpIncrement, Links.KTheta, true, true},
	{"r", func(l Links.Link) float64 { return l.Start.R }, OpLinkStart, Links.KR, true, false},
	{"Δr", func(l Links.Link) float64 { return float64(l.Delta.R) }, OpLinkPress, Links.KR, false, false},
	{"nΔr", func(l Links.Link) float64 { return float64(l.Increment.R) }, OpIncrement, Links.KR, true, false},
}

func (c linkCol) head(m Model, id string, step int) Cell {
	if c.op == OpLinkPress {
		st := m.headOf(step, id)
		way, on := st.At(c.k)
		lockAct := Act{Op: OpAllLock, Node: id, K: c.k, I: step}
		return Text(c.label).Acting(lockAct).Outline(*on).Stepping(lockAct, *way).LockedOn(*on)
	}
	return Text(c.label).AboveLastDigit()
}

func linkHeader(m Model, id string, step int) []Cell {
	labels := []Cell{Text("φ+θ+r").Acting(Act{Op: OpAllLocks, Node: id, I: step}).Outline(m.headOf(step, id).Both).Stepping(Act{Op: OpAllBoth, Node: id, I: step}, 0)}
	for _, c := range linkCols {
		labels = append(labels, c.head(m, id, step))
	}
	return labels
}

func LinksPill(m Model) Pill {
	groups := []Group{{Rows: [][]Cell{{Text("links — one per a→b, set at a: its start, and Δ on it (± steps Δ and moves b); a picked link's name is outlined; clicking a name in a step picks or unpicks it in that step")}}, Wraps: true}}
	for _, step := range StepNumbers {
		groups = append(groups, Group{Title: "step " + strconv.Itoa(step)})
		for _, g := range nodeTables(m, step) {
			g.Indent = true
			groups = append(groups, g)
		}
	}
	groups = append(groups, nodeTables(m, 0)...)
	if m.Refused != "" {
		groups = append(groups, Group{Rows: [][]Cell{{Text(m.Refused)}}, Wraps: true})
	}
	groups = append(groups, recalcGroup(m), landsGroup(m))
	return Pill{Label: LinksLabel, Groups: groups}
}

func nodeTables(m Model, step int) []Group {
	const fit = "0.0"
	selected := step == 0
	var groups []Group
	for _, id := range m.Nodes {
		var rows [][]Cell
		links := m.linksFrom(id)
		if !selected {
			links = m.stepLinksFrom(step, id)
		}
		for _, l := range links {
			act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: step} }
			name := Text(l.Name()).Outline(l.Picked)
			picked := l.Picked
			if !selected {
				picked = m.StepPicked(step, l)
				name = Text(l.Name()).Acting(Act{Op: OpStepPick, Node: l.From, To: l.To, I: step}).Outline(picked)
			}
			row := []Cell{name}
			for _, c := range linkCols {
				row = append(row, Field("", num1(c.value(l)), act(c.op, c.k), c.typed, true).Fitting(fit).Outline(c.picked && picked).Dimmed(c.k == Links.KR && c.op != OpIncrement && m.FixedR(l)))
			}
			if selected {
				row = append(row, Button("×", act(OpDropLink, ""), false))
			}
			rows = append(rows, row)
		}
		if len(rows) == 0 {
			rows = [][]Cell{{Text("no links")}}
		} else {
			rows = append([][]Cell{linkHeader(m, id, step)}, rows...)
		}
		groups = append(groups, Group{Title: "node " + id, Rows: rows, Slides: true})
	}
	return groups
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
