package Builder

import (
	"fmt"
	"math"
	"strconv"

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

func activeLinksGroup(m Model, step int) Group {
	deltas := []Cell{
		Text("active links"),
		Text("φ+θ+r").Acting(Act{Op: OpAllLocks, I: step}).Outline(m.headOf(step).Both).Stepping(Act{Op: OpAllBoth, I: step}, 0),
	}
	times := []Cell{Text(""), Text("")}
	ones := []Cell{Text(""), Text("")}
	for _, a := range axes {
		deltas = append(deltas, lockHead(m, step, a))
		times = append(times, Text("nΔ"+a.label).Stepping(Act{Op: OpAllInc, K: a.k, I: step}, 0))
		ones = append(ones, Text("1"+a.label).Stepping(Act{Op: OpAllOne, K: a.k, I: step}, 0))
	}
	return Group{Rows: [][]Cell{deltas, times, ones}}
}

func nodeGroups(m Model, step int, id string) []Group {
	groups := []Group{{Rows: [][]Cell{nodeLine(m, step, id)}, Slides: true, Table: fmt.Sprintf("step %d node %s", step, id)}}
	links := m.stepLinksFrom(step, id)
	if m.StepFolded(step, id) || len(links) == 0 {
		return groups
	}
	for _, l := range links {
		groups = append(groups, linkCard(m, step, l))
	}
	return groups
}

func linkCard(m Model, step int, l Links.Link) Group {
	act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: step} }
	head := func(s string) Cell { return Text(s).AboveLastDigit() }
	rows := [][]Cell{
		{Text(l.Name()).Acting(act(OpStepPick, "")).Outline(m.StepPicked(step, l)), Text(""), Text(""), Text(""), Text(""), Button("x", Act{Op: OpDisconnect, Node: l.From, To: l.To}, false)},
		{Text(""), head("start"), head("Δ"), head("nΔ"), head("1"), Text("on a press")},
	}
	lands := map[string]string{
		Links.KPhi:   "→ tick " + strconv.Itoa(l.Start.Tick.Phi+l.Delta.Phi),
		Links.KTheta: "→ tick " + strconv.Itoa(l.Start.Tick.Theta+l.Delta.Theta),
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
			Text(lands[a.k]),
		})
	}
	return Group{Rows: rows, Card: true, Indent: true}
}
