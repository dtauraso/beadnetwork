package Builder

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const LinksLabel = "Links"

const (
	OpLinkStart = "link-start"
	OpLinkPress = "link-press"
	OpLinkOne   = "link-one"
	OpStepShow  = "step-show"
	OpNetFold   = "network-fold"
	OpMenu      = "menu"
)

const (
	menuStep = "step"
	menuSeqs = "seqs"
	menuMore = "more"
	menuNode = "node"
)

func (m *Model) PressMenu(a Act) {
	if a.Op != OpMenu {
		if a.Op != OpSeqFind && a.Op != OpSeqFindClear && a.Op != OpFrame {
			m.Menu = Act{}
		}
		return
	}
	if m.Menu == a {
		m.Menu = Act{}
		return
	}
	m.Menu = a
}

const navHint = "click a row to select · double-click to open · two-finger scroll to pan"

func (m Model) allLinks() []Links.Link {
	var out []Links.Link
	for _, id := range m.Nodes {
		out = append(out, m.Col(id).Links()...)
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

func pathGroup(m Model) Group {
	sep := Text("›").Dimmed(true)
	if m.AtSeqs {
		return Group{Rows: [][]Cell{{Text("Sequences").Inked(inkStrong)}}, Slides: true, Table: "where you are"}
	}
	row := []Cell{Text("Sequences").Acting(Act{Op: OpSeqs}).Dimmed(true)}
	name := Text(m.seqName())
	if len(m.Trail) > 0 {
		name = name.Acting(Act{Op: OpLevelGo, Node: m.Trail[0].Level, I: m.Trail[0].Step}).Dimmed(true)
	} else {
		name = name.Inked(inkStrong)
	}
	row = append(row, sep, Text(m.openSeqTag()).Dimmed(true), name)
	for i, s := range m.Trail {
		c := Text(s.Label)
		if i+1 < len(m.Trail) {
			c = c.Acting(Act{Op: OpLevelGo, Node: m.Trail[i+1].Level, I: m.Trail[i+1].Step}).Dimmed(true)
		} else {
			c = c.Inked(inkStrong)
		}
		row = append(row, sep, c)
	}
	return Group{Rows: [][]Cell{row}, Slides: true, Table: "where you are"}
}

func linksTree(m Model) tree {
	t := tree{sel: m.Sel}
	t.put(allSeqsGroup(m), pathGroup(m))
	seqRows(&t, m)
	t.put(Group{Rows: [][]Cell{{Text(navHint).Dimmed(true)}}, Wraps: true})
	if m.Refused != "" {
		t.put(Group{Rows: [][]Cell{{Text(m.Refused)}}, Wraps: true})
	}
	return t
}

func LinksPill(m Model) Pill {
	if !m.AtSeqs {
		d := boardData(m)
		if m.NetworkOpen {
			d.Net = netData(m)
		}
		return Pill{Label: LinksLabel, Bottom: true, Board: func(box Sheet.Rect, sx, sy float32) Sheet.Board { return LinksBoard.Lay(box, sx, sy, d) }}
	}
	return Pill{Label: LinksLabel, Bottom: true, Groups: linksTree(m).groups}
}
