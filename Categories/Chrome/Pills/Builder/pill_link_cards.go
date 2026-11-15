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
	ref := m.Col(id).Ref
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
	term := func(k string, start, total, d int, end string) {
		if total != 0 || d != 0 {
			s := k + " " + strconv.Itoa(start)
			if total != 0 {
				s += " + total " + strconv.Itoa(total)
			}
			parts = append(parts, s+" + "+strconv.Itoa(d)+" → "+end)
		}
	}
	sum := l.Sum()
	term("φ", l.Start.Tick.Phi, l.Total.Phi, l.Delta.Phi, "tick "+strconv.Itoa(l.Start.Tick.Phi+sum.Phi))
	term("θ", l.Start.Tick.Theta, l.Total.Theta, l.Delta.Theta, "tick "+strconv.Itoa(l.Start.Tick.Theta+sum.Theta))
	if sum.R != 0 {
		parts = append(parts, "r "+tenths(l.Start.R)+" + "+strconv.Itoa(sum.R)+" → "+tenths(l.Start.R+float64(sum.R)))
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
