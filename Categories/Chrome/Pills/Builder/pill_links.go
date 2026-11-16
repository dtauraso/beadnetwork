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
	OpPlaceFold = "place-fold"
)

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

func openOf(on bool) int {
	if on {
		return open
	}
	return closed
}

func linkRowOf(l Links.Link, picked bool, pick Act, depth int) linkRow {
	r := linkRow{
		id:     Act{Op: OpLinkRow, Node: l.From, To: l.To, I: pick.I},
		depth:  depth,
		in:     pick,
		name:   Text(pickMark(picked)).Acting(pick).Tipped("pick this link"),
		detail: []Cell{Text(l.Name())},
	}
	if picked {
		r.detail = append(r.detail, Text(pressSummary(l)).Dimmed(true))
	}
	return r
}

func pathGroup(m Model) Group {
	sep := Text("›").Dimmed(true)
	if m.AtSeqs {
		return Group{Rows: [][]Cell{{Text("Sequences").Inked(inkStrong)}}, Slides: true, Table: "where you are"}
	}
	row := []Cell{Text("Sequences").Acting(Act{Op: OpSeqs}).Dimmed(true).Tipped("all sequences")}
	name := Text(m.seqName())
	if len(m.Trail) > 0 {
		name = name.Acting(Act{Op: OpLevelGo, Node: m.Trail[0].Level, I: m.Trail[0].Step}).Dimmed(true).Tipped("go back to this level")
	} else {
		name = name.Inked(inkStrong)
	}
	row = append(row, sep, Text(m.openSeqTag()).Dimmed(true), name)
	for i, s := range m.Trail {
		c := Text(s.Label)
		if i+1 < len(m.Trail) {
			c = c.Acting(Act{Op: OpLevelGo, Node: m.Trail[i+1].Level, I: m.Trail[i+1].Step}).Dimmed(true).Tipped("go back to this level")
		} else {
			c = c.Inked(inkStrong)
		}
		row = append(row, sep, c)
	}
	return Group{Rows: [][]Cell{row}, Slides: true, Table: "where you are"}
}

func linksTree(m Model) tree {
	t := tree{sel: m.Sel}
	if m.AtSeqs {
		t.put(allSeqsGroup(m), pathGroup(m))
		seqRows(&t, m)
		t.put(Group{Rows: [][]Cell{{Text(navHint).Dimmed(true)}}, Wraps: true})
	} else {
		t.put(Group{Rows: [][]Cell{{Button("← back to steps", Act{Op: OpNetFold}, false).Tipped("back to the step cards")}}})
		networkRows(&t, m)
		startRows(&t, m)
	}
	if m.Refused != "" {
		t.put(Group{Rows: [][]Cell{{Text(m.Refused)}}, Wraps: true})
	}
	return t
}

func LinksPill(m Model) Pill {
	if !m.AtSeqs && !m.NetworkOpen {
		d := boardData(m)
		return Pill{Label: LinksLabel, Bottom: true, Board: func(box Sheet.Rect, sx, sy float32) Sheet.Board { return LinksBoard.Lay(box, sx, sy, d) }}
	}
	return Pill{Label: LinksLabel, Bottom: true, Groups: linksTree(m).groups}
}
