package Builder

import (
	"fmt"
	"math"
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

func (m Model) StepShown() int {
	return min(max(m.ShownStep, StepNumbers[0]), StepNumbers[len(StepNumbers)-1])
}

func (m *Model) ShowStep(a Act, sign int) {
	if sign != 0 {
		m.ShownStep = m.StepShown() + sign
	} else {
		m.ShownStep = a.I
	}
	m.ShownStep = m.StepShown()
}

func stepGroup(m Model) Group {
	row := []Cell{Text("step")}
	for _, s := range StepNumbers {
		row = append(row, Button(strconv.Itoa(s), Act{Op: OpStepShow, I: s}, s == m.StepShown()))
	}
	row = append(row, Text("").Stepping(Act{Op: OpStepShow}, 0))
	return Group{Rows: [][]Cell{row}}
}

func LinksPill(m Model) Pill {
	var groups []Group
	if len(m.Seqs) > 0 {
		groups = append(groups, sequenceGroups(m)...)
	}
	step := m.StepShown()
	groups = append(groups, stepGroup(m), activeLinksGroup(m, step))
	refused := m.Refused
	for _, id := range m.Nodes {
		groups = append(groups, nodeGroups(m, step, id)...)
		if m.StepFolded(step, id) || refused == "" {
			continue
		}
		for _, l := range m.stepLinksFrom(step, id) {
			if strings.HasPrefix(refused, fmt.Sprintf("step %d %s ", step, l.Name())) {
				groups = append(groups, Group{Rows: [][]Cell{{Text(refused)}}, Wraps: true, Indent: true})
				refused = ""
				break
			}
		}
	}
	if refused != "" {
		groups = append(groups, Group{Rows: [][]Cell{{Text(refused)}}, Wraps: true})
	}
	groups = append(groups, sequenceList(m, 0)...)
	return Pill{Label: LinksLabel, Bottom: true, Groups: groups}
}

func num1(v float64) string {
	v = math.Round(v*10) / 10
	if v == math.Trunc(v) {
		return strconv.Itoa(int(v))
	}
	return strconv.FormatFloat(v, 'f', 1, 64)
}
