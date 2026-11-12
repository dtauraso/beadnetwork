package Builder

import (
	"fmt"
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

func nodeLine(m Model, step int, id string) []Cell {
	folded := m.StepFolded(step, id)
	links := m.stepLinksTo(step, id)
	row := []Cell{
		Text(foldTitle(folded, "node "+id)).Acting(Act{Op: OpStepFold, Node: id, I: step}),
	}
	if folded {
		return append(row, Text(linkCount(len(links))).Dimmed(true))
	}
	return row
}

func netNodeLine(m Model, id string) []Cell {
	row := nodeLine(m, 0, id)
	if m.StepFolded(0, id) {
		return row
	}
	ref := m.Reports[id].Ref
	refAct := func(k string) Act { return Act{Op: OpNodeRef, Node: id, K: k} }
	row = append(row,
		Text("ref Δφ"),
		Field("", strconv.Itoa(ref.Phi), refAct(Links.KPhi), true, true).Fitting(cardFit).Outline(false),
		Text("Δθ"),
		Field("", strconv.Itoa(ref.Theta), refAct(Links.KTheta), true, true).Fitting(cardFit).Outline(false),
	)
	if to := connectRow(m, id, "link to"); to != nil {
		row = append(row, to...)
	}
	return row
}

func headRow(m Model, step int) Group {
	row := []Cell{Text("HEAD")}
	for _, a := range axes {
		row = append(row, lockHead(m, step, a))
	}
	return Group{Rows: [][]Cell{row}, Indent: true}
}

func nodeGroups(m Model, step int, id string) []Group {
	groups := []Group{{Rows: [][]Cell{nodeLine(m, step, id)}, Slides: true, Table: fmt.Sprintf("step %d node %s", step, id)}}
	links := m.stepLinksTo(step, id)
	if m.StepFolded(step, id) || len(links) == 0 {
		return groups
	}
	for _, l := range links {
		picked := m.StepPicked(step, l)
		groups = append(groups, linkLine(l, picked, Act{Op: OpStepPick, Node: l.From, To: l.To, I: step}))
		if picked {
			groups = append(groups, linkCard(m, step, l))
		}
	}
	return groups
}

func pickMark(on bool) string {
	if on {
		return "● "
	}
	return "○ "
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

func linkLine(l Links.Link, picked bool, pick Act, more ...Cell) Group {
	row := []Cell{Text(pickMark(picked) + l.Name()).Acting(pick)}
	if picked {
		row = append(row, Text(pressSummary(l)).Dimmed(true))
	}
	return Group{Rows: [][]Cell{append(row, more...)}, Indent: true}
}

func linkCard(m Model, step int, l Links.Link) Group {
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: step} }
	head := func(s string) Cell { return Text(s).AboveLastDigit() }
	rows := [][]Cell{
		{Text(""), head("START"), head("Δ"), head("nΔ"), head("1")},
	}
	for _, a := range axes {
		dim := a.k == Links.KR && m.FixedR(l)
		field := func(v, op string, typed bool) Cell {
			return Field("", v, act(op, a.k), typed, true).Fitting(cardFit).Outline(false)
		}
		rows = append(rows, []Cell{
			Text(a.label),
			field(a.at(l), OpLinkStart, true).Dimmed(dim),
			field(strconv.Itoa(a.delta(l)), OpLinkPress, false).Dimmed(dim),
			field(strconv.Itoa(a.times(l)), OpIncrement, true),
			field(strconv.Itoa(a.one(l)), OpLinkOne, false).Dimmed(dim),
		})
	}
	return Group{Rows: rows, Card: true, Indent: true}
}
