package Builder

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const LinksLabel = "Links"

const (
	OpLinkStart = "link-start"
	OpLinkPress = "link-press"
	OpLinkOne   = "link-one"
	OpStepShow  = "step-show"
	OpNetFold   = "network-fold"
)

func (m Model) allLinks() []Links.Link {
	var out []Links.Link
	for _, id := range m.Nodes {
		out = append(out, m.Reports[id].Links...)
	}
	Links.SortLinks(out)
	return out
}

func (m Model) linksTo(id string) []Links.Link {
	var out []Links.Link
	for _, l := range m.allLinks() {
		if l.To == id {
			out = append(out, l)
		}
	}
	return out
}

func (m Model) StepShown() int {
	ns := m.stepNums()
	if slices.Contains(ns, m.ShownStep) {
		return m.ShownStep
	}
	if m.ShownStep > ns[len(ns)-1] {
		return ns[len(ns)-1]
	}
	return ns[0]
}

func (m *Model) ShowStep(a Act, sign int) {
	if sign == 0 {
		if a.I == m.StepShown() && !m.StepClosed {
			m.StepClosed = true
			return
		}
		m.ShownStep, m.StepClosed = a.I, false
		m.ShownStep = m.StepShown()
		return
	}
	ns := m.stepNums()
	i := slices.Index(ns, m.StepShown())
	m.ShownStep = ns[min(max(i+sign, 0), len(ns)-1)]
}

func networkGroups(m Model) []Group {
	groups := []Group{{Rows: [][]Cell{{
		Text(foldTitle(!m.NetworkOpen, "network")).Acting(Act{Op: OpNetFold}).Inked(inkGo),
		Text("what runs now · " + linkCount(len(m.allLinks()))).Dimmed(true),
	}}}}
	if !m.NetworkOpen {
		return groups
	}
	for _, id := range m.Nodes {
		groups = append(groups, Group{Rows: [][]Cell{netNodeLine(m, id)}, Slides: true, Table: "network node " + id, Indent: true})
		if m.StepFolded(0, id) {
			continue
		}
		for _, l := range m.linksTo(id) {
			groups = append(groups, linkLine(l, l.Picked, Act{}, Button("x", Act{Op: OpDisconnect, Node: l.From, To: l.To}, false).Tipped("remove this link")))
			if l.Picked {
				groups = append(groups, netCard(l))
			}
		}
	}
	return groups
}

func netCard(l Links.Link) Group {
	val := func(s string) Cell { return Text(s).Outline(false) }
	head := func(s string) Cell { return Text(s).Inked(inkHead) }
	rows := [][]Cell{{Text(""), head("START"), head("Δ"), head("nΔ"), head("1")}}
	for _, a := range axes {
		rows = append(rows, []Cell{Text(a.label).Dimmed(true), val(a.at(l)), val(strconv.Itoa(a.delta(l))), val(strconv.Itoa(a.times(l))), val(strconv.Itoa(a.one(l)))})
	}
	return Group{Rows: rows, Card: true, Indent: true}
}

func (m *Model) Touch(a Act) {
	if a.Node != "" && a.To != "" && a.I > 0 {
		m.Touched = StepEdge{Step: a.I, From: a.Node, To: a.To}
	}
}

func (m Model) stepOpen(n int) bool { return n == m.StepShown() && !m.StepClosed }

func stepRow(m Model, n int) Group {
	return Group{Rows: [][]Cell{{
		Text(foldTitle(!m.stepOpen(n), "step "+strconv.Itoa(n))).Acting(Act{Op: OpStepShow, I: n}).Inked(inkStrong).Tipped("open or close this step; drag to reorder"),
		stepNameCell(m, n),
		subStepsCell(n),
		Button("+", Act{Op: OpStepAdd, I: n}, false).Tipped("add a step after this one"),
		Button("⧉", Act{Op: OpStepCopy, I: n}, false).Tipped("duplicate this step"),
		Button("×", Act{Op: OpStepDrop, I: n}, false).Inked(inkDanger).Tipped("delete this step"),
	}}, Band: bandRow, Trail: 4}
}

func stepBody(m Model, step int, refused string) ([]Group, string) {
	groups := []Group{headRow(m, step)}
	for _, id := range m.Nodes {
		groups = append(groups, nodeGroups(m, step, id)...)
		if m.StepFolded(step, id) || refused == "" {
			continue
		}
		for _, l := range m.stepLinksTo(step, id) {
			if strings.HasPrefix(refused, fmt.Sprintf("step %d %s ", step, l.Name())) {
				groups = append(groups, Group{Rows: [][]Cell{{Text(refused)}}, Wraps: true, Indent: true})
				refused = ""
				break
			}
		}
	}
	return groups, refused
}

func breadcrumb(m Model) Group {
	row := trailCells(m)
	sep := Text("›").Dimmed(true)
	switch {
	case m.DragStep != 0:
		row = append(row, sep, Text("moving step "+m.StepLabel(m.DragStep)+": let go on another step's row"))
	case m.StepClosed:
		row = append(row, sep, Text("network").Inked(inkStrong))
	default:
		step := m.StepShown()
		row = append(row, sep, Text(stepLabelAt(m.lvl(), step)).Inked(inkStrong))
		if t := m.Touched; t.Step == step && t.From != "" {
			row = append(row, sep, Text(t.From+"→"+t.To).Inked(inkAccent))
		}
	}
	return Group{Rows: [][]Cell{row}, Slides: true, Table: "where you are", Band: bandRow}
}

func LinksPill(m Model) Pill {
	var groups []Group
	groups = append(groups, breadcrumb(m))
	groups = append(groups, networkGroups(m)...)
	refused := m.Refused
	for _, n := range m.stepNums() {
		groups = append(groups, stepRow(m, n))
		if m.stepOpen(n) {
			var body []Group
			body, refused = stepBody(m, n, refused)
			groups = append(groups, body...)
		}
	}
	if ns := m.stepNums(); len(ns) > 0 {
		groups = append(groups,
			Group{Rows: [][]Cell{{Text("+  add step").Acting(Act{Op: OpStepAdd, I: ns[len(ns)-1]}).Inked(inkAccent).Tipped("add a step after the last one")}}},
			Group{Rows: [][]Cell{{Text("drag a step row to reorder").Dimmed(true)}}})
	}
	if refused != "" {
		groups = append(groups, Group{Rows: [][]Cell{{Text(refused)}}, Wraps: true})
	}
	groups = append(groups, sequenceList(m, 0)...)
	return Pill{Label: LinksLabel, Bottom: true, Groups: groups}
}
