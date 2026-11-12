package Builder

import (
	"math"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const cardFit = "0.0"

func tenths(v float64) string {
	t := math.Round(v * 10)
	s := strconv.FormatInt(int64(math.Abs(t))/10, 10)
	if d := int64(math.Abs(t)) % 10; d != 0 {
		s += "." + strconv.FormatInt(d, 10)
	}
	if t < 0 {
		s = "-" + s
	}
	return s
}

type axis struct {
	label, k          string
	at                func(Links.Link) string
	delta, times, one func(Links.Link) int
}

var axes = []axis{
	{"φ", Links.KPhi, func(l Links.Link) string { return strconv.Itoa(l.Start.Tick.Phi) }, func(l Links.Link) int { return l.Delta.Phi }, func(l Links.Link) int { return l.Increment.Phi }, func(l Links.Link) int { return l.Start.One.Phi }},
	{"θ", Links.KTheta, func(l Links.Link) string { return strconv.Itoa(l.Start.Tick.Theta) }, func(l Links.Link) int { return l.Delta.Theta }, func(l Links.Link) int { return l.Increment.Theta }, func(l Links.Link) int { return l.Start.One.Theta }},
	{"r", Links.KR, func(l Links.Link) string { return tenths(l.Start.R) }, func(l Links.Link) int { return l.Delta.R }, func(l Links.Link) int { return l.Increment.R }, func(l Links.Link) int { return l.Start.One.R }},
}

func lockHead(m Model, step int, a axis) Cell {
	st := m.headOf(step)
	way, on := st.At(a.k)
	lock := Act{Op: OpAllLock, K: a.k, I: step}
	glyph := "·"
	if *on && *way > 0 {
		glyph = "↑"
	} else if *on && *way < 0 {
		glyph = "↓"
	}
	return Text(a.label+" "+glyph).Acting(lock).Outline(*on).Stepping(lock, *way).LockedOn(*on)
}

func linkCount(n int) string {
	switch n {
	case 0:
		return "no links"
	case 1:
		return "1 link"
	}
	return strconv.Itoa(n) + " links"
}

func refCells(m Model, id string) []Cell {
	ref := m.Reports[id].Ref
	refAct := func(k string) Act { return Act{Op: OpNodeRef, Node: id, K: k} }
	return []Cell{
		Text("ref Δφ").Dimmed(true),
		Field("", strconv.Itoa(ref.Phi), refAct(Links.KPhi), true, true).Fitting(cardFit).Outline(false),
		Text("Δθ").Dimmed(true),
		Field("", strconv.Itoa(ref.Theta), refAct(Links.KTheta), true, true).Fitting(cardFit).Outline(false),
	}
}

func headRow(m Model, step, depth int) Group {
	row := []Cell{Text("HEAD").Inked(inkHead)}
	for _, a := range axes {
		row = append(row, lockHead(m, step, a))
	}
	return Group{Rows: [][]Cell{row}, Depth: depth}
}

func pickMark(on bool) string {
	if on {
		return "●"
	}
	return "○"
}

func pressSummary(l Links.Link) string {
	var parts []string
	if l.Delta.Phi != 0 {
		parts = append(parts, "φ "+strconv.Itoa(l.Start.Tick.Phi)+" + "+strconv.Itoa(l.Delta.Phi)+" → tick "+strconv.Itoa(l.Start.Tick.Phi+l.Delta.Phi))
	}
	if l.Delta.Theta != 0 {
		parts = append(parts, "θ "+strconv.Itoa(l.Start.Tick.Theta)+" + "+strconv.Itoa(l.Delta.Theta)+" → tick "+strconv.Itoa(l.Start.Tick.Theta+l.Delta.Theta))
	}
	if l.Delta.R != 0 {
		parts = append(parts, "r "+tenths(l.Start.R)+" + "+strconv.Itoa(l.Delta.R)+" → "+tenths(l.Start.R+float64(l.Delta.R)))
	}
	if len(parts) == 0 {
		return "no move"
	}
	return strings.Join(parts, " · ")
}

func netCard(l Links.Link, depth int) Group {
	val := func(s string) Cell { return Text(s).Outline(false) }
	head := func(s string) Cell { return Text(s).Inked(inkHead) }
	rows := [][]Cell{{Text(""), head("START"), head("Δ"), head("nΔ"), head("1")}}
	for _, a := range axes {
		rows = append(rows, []Cell{Text(a.label).Dimmed(true), val(a.at(l)), val(strconv.Itoa(a.delta(l))), val(strconv.Itoa(a.times(l))), val(strconv.Itoa(a.one(l)))})
	}
	return Group{Rows: rows, Card: true, Depth: depth}
}

const GridLabel = "Node grid"

func nodeGrid(m Model, id string) Group {
	ns := m.stepNums()
	head := []Cell{Text("")}
	for _, n := range ns {
		head = append(head, Text(m.StepLabel(n)).Inked(inkHead))
	}
	same := func(name string, v int) []Cell {
		row := []Cell{Text(name).Dimmed(true)}
		for range ns {
			row = append(row, Text(strconv.Itoa(v)).Outline(false))
		}
		return row
	}
	ref := m.Reports[id].Ref
	rows := [][]Cell{head, same("ref Δφ", ref.Phi), same("ref Δθ", ref.Theta)}
	links := m.linksTo(id)
	if len(links) == 0 {
		rows = append(rows, []Cell{Text("no links into this node").Dimmed(true)})
	}
	for _, l := range links {
		if len(links) > 1 {
			rows = append(rows, []Cell{Text(l.Name()).Inked(inkStrong)})
		}
		for _, a := range axes {
			for _, v := range []struct {
				name string
				of   func(Links.Link) string
			}{
				{"start", a.at},
				{"Δ", func(l Links.Link) string { return strconv.Itoa(a.delta(l)) }},
				{"nΔ", func(l Links.Link) string { return strconv.Itoa(a.times(l)) }},
				{"1 step", func(l Links.Link) string { return strconv.Itoa(a.one(l)) }},
			} {
				row := []Cell{Text(a.label + " " + v.name).Dimmed(true)}
				for _, n := range ns {
					c := l
					if s := m.StepLinks[StepEdge{n, l.From, l.To}]; s != nil {
						c = *s
					}
					cell := Text(v.of(c)).Outline(false)
					if m.StepPicked(n, l) {
						cell = cell.Inked(inkStrong)
					}
					row = append(row, cell)
				}
				rows = append(rows, row)
			}
		}
	}
	rows = append(rows, same("s", m.SceneS()))
	return Group{Title: "node " + id, Rows: rows, Slides: true, Table: "grid node " + id}
}

func GridPill(m Model) Pill {
	var groups []Group
	for _, id := range m.Nodes {
		groups = append(groups, nodeGrid(m, id))
	}
	return Pill{Label: GridLabel, Bottom: true, Groups: groups}
}

func linkCard(m Model, step int, l Links.Link, depth int) Group {
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: step} }
	head := func(s string) Cell { return Text(s).AboveLastDigit().Inked(inkHead) }
	rows := [][]Cell{
		{Text(""), head("START"), head("Δ"), head("nΔ"), head("1")},
	}
	for _, a := range axes {
		dim := a.k == Links.KR && m.FixedR(l)
		field := func(v, op string, typed bool) Cell {
			return Field("", v, act(op, a.k), typed, true).Fitting(cardFit).Outline(false)
		}
		rows = append(rows, []Cell{
			Text(a.label).Dimmed(true),
			field(a.at(l), OpLinkStart, true).Dimmed(dim),
			field(strconv.Itoa(a.delta(l)), OpLinkPress, false).Dimmed(dim),
			field(strconv.Itoa(a.times(l)), OpIncrement, true),
			field(strconv.Itoa(a.one(l)), OpLinkOne, false).Dimmed(dim),
		})
	}
	return Group{Rows: rows, Card: true, Depth: depth}
}
