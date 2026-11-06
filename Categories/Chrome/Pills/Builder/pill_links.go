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

func (c linkCol) head(m Model) Cell {
	if c.op == OpLinkPress {
		way, on := m.Step.At(c.k)
		lockAct := Act{Op: OpAllLock, K: c.k}
		return Text(c.label).Acting(lockAct).Outline(*on).Stepping(lockAct, *way).LockedOn(*on)
	}
	return Text(c.label).AboveLastDigit()
}

func LinksPill(m Model) Pill {
	const fit = "0.0"
	cols := linkCols
	var rows [][]Cell
	for _, l := range m.allLinks() {
		act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k} }
		pick := OpPick
		if l.Picked {
			pick = OpUnpick
		}
		row := []Cell{Text(l.Name()).Acting(act(pick, "")).Outline(l.Picked)}
		for _, c := range cols {
			row = append(row, Field("", num1(c.value(l)), act(c.op, c.k), c.typed, true).Fitting(fit).Outline(c.picked && l.Picked).Dimmed(c.k == Links.KR && c.op != OpIncrement && m.FixedR(l)))
		}
		rows = append(rows, append(row, Button("×", act(OpDropLink, ""), false)))
	}
	if len(rows) == 0 {
		rows = [][]Cell{{Text("no links")}}
	} else {
		labels := []Cell{Text("φ+θ+r").Acting(Act{Op: OpAllLocks}).Outline(m.Step.Both).Stepping(Act{Op: OpAllBoth}, 0)}
		for _, c := range cols {
			labels = append(labels, c.head(m))
		}
		rows = append([][]Cell{labels}, rows...)
	}
	groups := []Group{
		{Title: "links — one per a→b, set at a: its start, and Δ on it (± steps Δ and moves b); a picked link's name is outlined, and clicking a name picks or unpicks it", Rows: rows, Slides: true},
	}
	if m.Refused != "" {
		groups = append(groups, Group{Rows: [][]Cell{{Text(m.Refused)}}, Wraps: true})
	}
	groups = append(groups, pickGroup(m), pickedGroup(m), recalcGroup(m), landsGroup(m))
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
