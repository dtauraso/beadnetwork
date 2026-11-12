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

func networkGroup(m Model) Group {
	links := m.allLinks()
	g := Group{Title: foldTitle(!m.NetworkOpen, "network · what runs now, "+linkCount(len(links))), Fold: Act{Op: OpNetFold}}
	if !m.NetworkOpen {
		return g
	}
	for _, l := range links {
		g.Rows = append(g.Rows, []Cell{Text(l.Name()).Outline(l.Picked), Text("start " + ang(l.Start.Phi, l.Start.Theta, &l.Start.R)), Text("Δ " + ticks(l.Delta))})
	}
	return g
}

func (m *Model) Touch(a Act) {
	if a.Node != "" && a.To != "" && a.I > 0 {
		m.Touched = StepEdge{Step: a.I, From: a.Node, To: a.To}
	}
}

func (m Model) stepOpen(n int) bool { return n == m.StepShown() && !m.StepClosed }

func stepRow(m Model, n int) Group {
	return Group{Rows: [][]Cell{{
		Text(foldTitle(!m.stepOpen(n), "step "+strconv.Itoa(n))).Acting(Act{Op: OpStepShow, I: n}),
		stepNameCell(m, n),
		Button("+", Act{Op: OpStepAdd, I: n}, false).Tipped("add a step after this one"),
		Button("⧉", Act{Op: OpStepCopy, I: n}, false).Tipped("duplicate this step"),
		Button("×", Act{Op: OpStepDrop, I: n}, false).Tipped("delete this step"),
	}}}
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
	trail := "network only, no step open"
	if m.DragStep != 0 {
		trail = "moving step " + m.StepLabel(m.DragStep) + ": let go on another step's title"
	} else if !m.StepClosed {
		step := m.StepShown()
		trail = "editing step " + m.StepLabel(step)
		if t := m.Touched; t.Step == step && t.From != "" {
			trail += " › " + t.From + "→" + t.To
		}
	}
	for _, q := range m.Seqs {
		if q.K == m.OpenSeq {
			trail = q.Name + " › " + trail
		}
	}
	return Group{Rows: [][]Cell{{Text(trail).Dimmed(true)}}, Wraps: true}
}

func LinksPill(m Model) Pill {
	var groups []Group
	if len(m.Seqs) > 0 {
		groups = append(groups, sequenceGroups(m)...)
	}
	groups = append(groups, breadcrumb(m), networkGroup(m))
	refused := m.Refused
	for _, n := range m.stepNums() {
		groups = append(groups, stepRow(m, n))
		if m.stepOpen(n) {
			var body []Group
			body, refused = stepBody(m, n, refused)
			groups = append(groups, body...)
		}
	}
	if refused != "" {
		groups = append(groups, Group{Rows: [][]Cell{{Text(refused)}}, Wraps: true})
	}
	groups = append(groups, sequenceList(m, 0)...)
	return Pill{Label: LinksLabel, Bottom: true, Groups: groups}
}
