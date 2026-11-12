package Builder

import (
	"fmt"
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

const navHint = "↑↓ move · → open / in · ← close / out · Enter go in · F2 rename · drag a row to reorder"

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

func networkRows(t *tree, m Model) {
	t.add(linkRow{
		id:     Act{Op: OpNetFold},
		open:   openOf(m.NetworkOpen),
		fold:   Act{Op: OpNetFold},
		name:   Text("network"),
		detail: []Cell{Text("what runs now · " + linkCount(len(m.allLinks()))).Dimmed(true)},
	})
	if !m.NetworkOpen {
		return
	}
	for _, id := range m.Nodes {
		fold := Act{Op: OpStepFold, Node: id}
		r := linkRow{id: fold, depth: 1, open: openOf(!m.StepFolded(0, id)), fold: fold, name: Text("node " + id), table: "network node " + id}
		if m.StepFolded(0, id) {
			r.detail = []Cell{Text(linkCount(len(m.linksTo(id)))).Dimmed(true)}
		} else {
			r.more = refCells(m, id)
		}
		r.acts = connectRow(m, id, "link to")
		t.add(r)
		if m.StepFolded(0, id) {
			continue
		}
		for _, l := range m.linksTo(id) {
			e := StepEdge{0, l.From, l.To}
			lr := m.cardRow(linkRowOf(l, l.Picked, Act{}, 2), e, l.Picked)
			lr.acts = []Cell{Button("×", Act{Op: OpDisconnect, Node: l.From, To: l.To}, false).Tipped("remove this link")}
			t.add(lr)
			if m.CardOpen(e, l.Picked) {
				t.put(netCard(l, 3))
			}
		}
	}
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

func stepRows(t *tree, m Model, refused string) string {
	ns := m.stepNums()
	for _, n := range ns {
		show := Act{Op: OpStepShow, I: n}
		rename := Act{Op: OpRename, K: renameStep, I: n}
		r := linkRow{
			id:     show,
			open:   openOf(m.stepOpen(n)),
			fold:   show,
			in:     Act{Op: OpStepEnter, I: n},
			rename: rename,
			name:   Field("step "+itoa(n), m.StepNames[n], Act{Op: OpStepName, I: n}, true, false).Plain().Wording(),
			acts: []Cell{
				Button("→", Act{Op: OpStepEnter, I: n}, false).Tipped("go into this step's own steps"),
				Button("✎", rename, false).Tipped("rename this step"),
				Button("+", Act{Op: OpStepAdd, I: n}, false).Tipped("add a step after this one"),
				Button("⧉", Act{Op: OpStepCopy, I: n}, false).Tipped("duplicate this step"),
				Button("×", Act{Op: OpStepDrop, I: n}, false).Tipped("delete this step"),
			},
		}
		if m.StepNames[n] != "" {
			r.detail = []Cell{Text("step " + itoa(n)).Dimmed(true)}
		}
		t.add(r)
		if m.stepOpen(n) {
			refused = stepBody(t, m, n, refused)
		}
	}
	return refused
}

func stepBody(t *tree, m Model, step int, refused string) string {
	t.put(headRow(m, step, 1))
	for _, id := range m.Nodes {
		fold := Act{Op: OpStepFold, Node: id, I: step}
		folded := m.StepFolded(step, id)
		links := m.gridLinks(step, id)
		r := linkRow{id: fold, depth: 1, open: openOf(!folded), fold: fold, name: Text("node " + id)}
		if len(links) == 0 {
			r.open = leaf
		}
		if folded {
			r.detail = []Cell{Text(linkCount(len(links))).Dimmed(true)}
		}
		t.add(r)
		if folded {
			continue
		}
		for _, l := range links {
			picked := m.StepPicked(step, l)
			e := StepEdge{step, l.From, l.To}
			t.add(m.cardRow(linkRowOf(l, picked, Act{Op: OpStepPick, Node: l.From, To: l.To, I: step}, 2), e, picked))
			if m.CardOpen(e, picked) {
				t.put(linkCard(m, step, l, 3))
			}
			if refused != "" && strings.HasPrefix(refused, fmt.Sprintf("step %d %s ", step, l.Name())) {
				t.put(Group{Rows: [][]Cell{{Text(refused)}}, Wraps: true, Depth: 3})
				refused = ""
			}
		}
	}
	return refused
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
	row = append(row, sep, name)
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
	t.put(pathGroup(m))
	refused := m.Refused
	if m.AtSeqs {
		seqRows(&t, m)
	} else {
		networkRows(&t, m)
		refused = stepRows(&t, m, refused)
		ns := m.stepNums()
		t.put(Group{Rows: [][]Cell{{Text("+  add step").Acting(Act{Op: OpStepAdd, I: ns[len(ns)-1]}).Dimmed(true).Tipped("add a step after the last one")}}})
	}
	if m.DragStep != 0 {
		t.put(Group{Rows: [][]Cell{{Text("moving step " + m.StepLabel(m.DragStep) + ": let go on another step's row").Dimmed(true)}}})
	} else {
		t.put(Group{Rows: [][]Cell{{Text(navHint).Dimmed(true)}}, Wraps: true})
	}
	if refused != "" {
		t.put(Group{Rows: [][]Cell{{Text(refused)}}, Wraps: true})
	}
	return t
}

func LinksPill(m Model) Pill {
	return Pill{Label: LinksLabel, Bottom: true, Groups: linksTree(m).groups}
}
