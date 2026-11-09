package Builder

import (
	"fmt"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const cardFit = "0.0"

type axis struct {
	label, k          string
	at                func(Links.Link) float64
	delta, times, one func(Links.Link) int
}

var axes = []axis{
	{"φ", Links.KPhi, func(l Links.Link) float64 { return l.Start.Phi }, func(l Links.Link) int { return l.Delta.Phi }, func(l Links.Link) int { return l.Increment.Phi }, func(l Links.Link) int { return l.Start.One.Phi }},
	{"θ", Links.KTheta, func(l Links.Link) float64 { return l.Start.Theta }, func(l Links.Link) int { return l.Delta.Theta }, func(l Links.Link) int { return l.Increment.Theta }, func(l Links.Link) int { return l.Start.One.Theta }},
	{"r", Links.KR, func(l Links.Link) float64 { return l.Start.R }, func(l Links.Link) int { return l.Delta.R }, func(l Links.Link) int { return l.Increment.R }, func(l Links.Link) int { return l.Start.One.R }},
}

func lockHead(m Model, id string, step int, a axis) Cell {
	st := m.headOf(step, id)
	way, on := st.At(a.k)
	lock := Act{Op: OpAllLock, Node: id, K: a.k, I: step}
	return Text("Δ"+a.label).Acting(lock).Outline(*on).Stepping(lock, *way).LockedOn(*on)
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
	links := m.stepLinksFrom(step, id)
	row := []Cell{
		Text(foldTitle(folded, "node "+id)).Acting(Act{Op: OpStepFold, Node: id, I: step}),
		Text("s"),
		Field("", strconv.Itoa(m.ScaleOf(id)), Act{Op: OpNodeS, Node: id}, true, false).Fitting(cardFit).Outline(false),
	}
	if folded {
		return append(row, Text(linkCount(len(links))))
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

func allLinksLines(m Model, id string, step int) [][]Cell {
	deltas := []Cell{
		Text("every link"),
		Text("φ+θ+r").Acting(Act{Op: OpAllLocks, Node: id, I: step}).Outline(m.headOf(step, id).Both).Stepping(Act{Op: OpAllBoth, Node: id, I: step}, 0),
	}
	rest := []Cell{Text(""), Text("")}
	for _, a := range axes {
		deltas = append(deltas, lockHead(m, id, step, a))
		rest = append(rest, Text("nΔ"+a.label).Stepping(Act{Op: OpAllInc, Node: id, K: a.k, I: step}, 0))
	}
	for _, a := range axes {
		rest = append(rest, Text("1"+a.label).Stepping(Act{Op: OpAllOne, Node: id, K: a.k, I: step}, 0))
	}
	return [][]Cell{deltas, rest}
}

func nodeGroups(m Model, step int, id string) []Group {
	groups := []Group{{Rows: [][]Cell{nodeLine(m, step, id)}, Slides: true, Table: fmt.Sprintf("step %d node %s", step, id)}}
	links := m.stepLinksFrom(step, id)
	if m.StepFolded(step, id) || len(links) == 0 {
		return groups
	}
	groups = append(groups, Group{Rows: allLinksLines(m, id, step), Indent: true})
	for _, l := range links {
		groups = append(groups, linkCard(m, step, l))
	}
	return groups
}

func linkCard(m Model, step int, l Links.Link) Group {
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: step} }
	head := func(s string) Cell { return Text(s).AboveLastDigit() }
	rows := [][]Cell{
		{Text(l.Name()).Acting(act(OpStepPick, "")).Outline(m.StepPicked(step, l)), Text(""), Text(""), Text(""), Button("x", Act{Op: OpDisconnect, Node: l.From, To: l.To}, false)},
		{Text(""), head("at"), head("Δ"), head("nΔ"), head("1")},
	}
	for _, a := range axes {
		dim := a.k == Links.KR && m.FixedR(l)
		field := func(v, op string, typed bool) Cell {
			return Field("", v, act(op, a.k), typed, true).Fitting(cardFit).Outline(false)
		}
		rows = append(rows, []Cell{
			Text(a.label),
			field(num1(a.at(l)), OpLinkStart, true).Dimmed(dim),
			field(strconv.Itoa(a.delta(l)), OpLinkPress, false).Dimmed(dim),
			field(strconv.Itoa(a.times(l)), OpIncrement, true),
			field(strconv.Itoa(a.one(l)), OpLinkOne, false).Dimmed(dim),
		})
	}
	return Group{Rows: rows, Card: true, Indent: true}
}
