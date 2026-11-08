package Builder

import (
	"math"
	"strconv"

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

func (m Model) SetS(id string, v int) Links.SetS {
	return Links.SetS{V: v}
}

type linkCol struct {
	label  string
	value  func(Links.Link) string
	op, k  string
	typed  bool
	picked bool
}

var linkCols = []linkCol{
	{"φ", func(l Links.Link) string { return num1(l.Start.Phi) }, OpLinkStart, Links.KPhi, true, false},
	{"Δφ", func(l Links.Link) string { return strconv.Itoa(l.Delta.Phi) }, OpLinkPress, Links.KPhi, false, true},
	{"nΔφ", func(l Links.Link) string { return strconv.Itoa(l.Increment.Phi) }, OpIncrement, Links.KPhi, true, true},
	{"θ", func(l Links.Link) string { return num1(l.Start.Theta) }, OpLinkStart, Links.KTheta, true, false},
	{"Δθ", func(l Links.Link) string { return strconv.Itoa(l.Delta.Theta) }, OpLinkPress, Links.KTheta, false, true},
	{"nΔθ", func(l Links.Link) string { return strconv.Itoa(l.Increment.Theta) }, OpIncrement, Links.KTheta, true, true},
	{"r", func(l Links.Link) string { return num1(l.Start.R) }, OpLinkStart, Links.KR, true, false},
	{"Δr", func(l Links.Link) string { return strconv.Itoa(l.Delta.R) }, OpLinkPress, Links.KR, false, false},
	{"nΔr", func(l Links.Link) string { return strconv.Itoa(l.Increment.R) }, OpIncrement, Links.KR, true, false},
}

func (c linkCol) head(m Model, id string, step int) Cell {
	if c.op == OpLinkPress {
		st := m.headOf(step, id)
		way, on := st.At(c.k)
		lockAct := Act{Op: OpAllLock, Node: id, K: c.k, I: step}
		return Text(c.label).Acting(lockAct).Outline(*on).Stepping(lockAct, *way).LockedOn(*on)
	}
	if c.op == OpIncrement {
		return Text(c.label).Stepping(Act{Op: OpAllInc, Node: id, K: c.k, I: step}, 0)
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
	var groups []Group
	if len(m.Seqs) > 0 {
		groups = append(groups, sequenceGroup(m))
	}
	for _, step := range StepNumbers {
		groups = append(groups, Group{Title: "step " + strconv.Itoa(step)})
		for _, g := range nodeTables(m, step) {
			g.Indent = true
			groups = append(groups, g)
		}
	}
	if m.Refused != "" {
		groups = append(groups, Group{Rows: [][]Cell{{Text(m.Refused)}}, Wraps: true})
	}
	return Pill{Label: LinksLabel, Bottom: true, Groups: groups}
}

func nodeTables(m Model, step int) []Group {
	const fit = "0.0"
	var groups []Group
	for _, id := range m.Nodes {
		fold := Act{Op: OpStepFold, Node: id, I: step}
		if m.StepFolded(step, id) {
			groups = append(groups, Group{Title: foldTitle(true, "node "+id), Fold: fold})
			continue
		}
		var rows [][]Cell
		for _, l := range m.stepLinksFrom(step, id) {
			l = l.Shown()
			act := func(op, k string) Act { return Act{Op: op, Node: l.From, To: l.To, K: k, I: step} }
			picked := m.StepPicked(step, l)
			row := []Cell{Text(l.Name()).Acting(act(OpStepPick, "")).Outline(picked)}
			for _, c := range linkCols {
				row = append(row, Field("", c.value(l), act(c.op, c.k), c.typed, true).Fitting(fit).Outline(c.picked && picked).Dimmed(c.k == Links.KR && c.op != OpIncrement && m.FixedR(l)))
			}
			row = append(row, Button("x", Act{Op: OpDisconnect, Node: l.From, To: l.To}, false))
			rows = append(rows, row)
		}
		if len(rows) > 0 {
			rows = append([][]Cell{linkHeader(m, id, step)}, rows...)
		}
		if row := connectRow(m, id, "link to"); row != nil {
			rows = append(rows, row)
		} else if len(rows) == 0 {
			rows = [][]Cell{{Text("no links")}}
		}
		groups = append(groups, Group{Title: foldTitle(false, "node "+id), Fold: fold}, Group{Rows: refRows(m, id, fit), Slides: true}, Group{Rows: rows, Slides: true})
	}
	return groups
}

func refRows(m Model, id, fit string) [][]Cell {
	s := float64(m.ScaleOf(id))
	ref := m.Reports[id].Ref
	act := func(k string) Act { return Act{Op: OpNodeRef, Node: id, K: k} }
	return [][]Cell{
		{Text("reference"), Text("Δφ").AboveLastDigit(), Text("Δθ").AboveLastDigit(), Text("s").AboveLastDigit()},
		{Text("ref " + id).Outline(false), Field("", num1(ref.Phi*s), act(Links.KPhi), true, true).Fitting(fit).Outline(false), Field("", num1(ref.Theta*s), act(Links.KTheta), true, true).Fitting(fit).Outline(false), Field("", strconv.Itoa(m.ScaleOf(id)), Act{Op: OpNodeS, Node: id}, true, false).Fitting(fit).Outline(false)},
	}
}

func num1(v float64) string {
	v = math.Round(v*10) / 10
	if v == math.Trunc(v) {
		return strconv.Itoa(int(v))
	}
	return strconv.FormatFloat(v, 'f', 1, 64)
}
