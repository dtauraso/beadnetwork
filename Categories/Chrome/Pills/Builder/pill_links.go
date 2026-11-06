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
	label string
	value func(Links.Link) float64
	op, k string
	typed bool
}

var linkCols = []linkCol{
	{"φ", func(l Links.Link) float64 { return l.Start.Phi }, OpLinkStart, Links.KPhi, true},
	{"Δφ", func(l Links.Link) float64 { return float64(l.Delta.Phi) }, OpLinkPress, Links.KPhi, false},
	{"nΔφ", func(l Links.Link) float64 { return float64(l.Increment.Phi) }, OpIncrement, Links.KPhi, true},
	{"θ", func(l Links.Link) float64 { return l.Start.Theta }, OpLinkStart, Links.KTheta, true},
	{"Δθ", func(l Links.Link) float64 { return float64(l.Delta.Theta) }, OpLinkPress, Links.KTheta, false},
	{"nΔθ", func(l Links.Link) float64 { return float64(l.Increment.Theta) }, OpIncrement, Links.KTheta, true},
	{"r", func(l Links.Link) float64 { return l.Start.R }, OpLinkStart, Links.KR, true},
	{"Δr", func(l Links.Link) float64 { return float64(l.Delta.R) }, OpLinkPress, Links.KR, false},
	{"nΔr", func(l Links.Link) float64 { return float64(l.Increment.R) }, OpIncrement, Links.KR, true},
}

func (c linkCol) head(m Model, id string) Cell {
	if c.op == OpLinkPress {
		st := m.StepOf(id)
		way, on := st.At(c.k)
		lockAct := Act{Op: OpAllLock, Node: id, K: c.k}
		return Text(c.label).Acting(lockAct).Outline(*on).Stepping(lockAct, *way).LockedOn(*on)
	}
	return Text(c.label).AboveLastDigit()
}

func linkHeader(m Model, id string) []Cell {
	labels := []Cell{Text("φ+θ+r").Acting(Act{Op: OpAllLocks, Node: id}).Outline(m.StepOf(id).Both).Stepping(Act{Op: OpAllBoth, Node: id}, 0)}
	for _, c := range linkCols {
		labels = append(labels, c.head(m, id))
	}
	return labels
}

func LinksPill(m Model) Pill {
	const fit = "0.0"
	cols := linkCols
	groups := []Group{{Rows: [][]Cell{{Text("links — one per a→b, set at a: its start, and Δ on it (± steps Δ and moves b); clicking a name picks or unpicks it")}}, Wraps: true}, stepGroup(m)}
	for _, id := range m.Nodes {
		var rows [][]Cell
		for _, l := range m.linksFrom(id) {
			act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k} }
			pick := OpPick
			if l.Picked {
				pick = OpUnpick
			}
			row := []Cell{Text(l.Name()).Acting(act(pick, ""))}
			for _, c := range cols {
				row = append(row, Field("", num1(c.value(l)), act(c.op, c.k), c.typed, true).Fitting(fit).Dimmed(c.k == Links.KR && c.op != OpIncrement && m.FixedR(l)))
			}
			rows = append(rows, append(row, Button("×", act(OpDropLink, ""), false)))
		}
		if len(rows) == 0 {
			rows = [][]Cell{{Text("no links")}}
		} else {
			rows = append([][]Cell{linkHeader(m, id)}, rows...)
		}
		groups = append(groups, Group{Title: "node " + id, Rows: rows, Slides: true})
	}
	if m.Refused != "" {
		groups = append(groups, Group{Rows: [][]Cell{{Text(m.Refused)}}, Wraps: true})
	}
	groups = append(groups, pickGroup(m), pickedGroup(m), recalcGroup(m), landsGroup(m))
	return Pill{Label: LinksLabel, Groups: groups}
}

func stepGroup(m Model) Group {
	rows := [][]Cell{{Text("node"), Text("table")}}
	for _, id := range m.Nodes {
		rows = append(rows, []Cell{Text("node " + id), Text(names(linkNames(m.linksFrom(id))))})
	}
	return Group{Title: "step 1", Rows: rows}
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
