package Builder

import (
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
)

const (
	OpSeqFind      = "seq-find"
	OpSeqFindClear = "seq-find-clear"

	OpSeqOpen = "seq-open"
	OpSeqIn   = "seq-in"
	OpSeqAdd  = "seq-add"
	OpSeqCopy = "seq-copy"
	OpSeqDrop = "seq-drop"
	OpSeqName = "seq-name"
	OpSeqs    = "seqs"
)

type Seq struct {
	K    int
	ID   int
	Name string
}

func (q Seq) Tag() string { return "#" + itoa(q.ID) }

func (m Model) openSeqTag() string {
	for _, q := range m.Seqs {
		if q.K == m.OpenSeq {
			return q.Tag()
		}
	}
	return ""
}

func IsSeqOp(a Act) bool {
	switch a.Op {
	case OpSeqIn, OpSeqAdd, OpSeqCopy, OpSeqDrop:
		return true
	}
	return false
}

func (m Model) seqName() string {
	for _, q := range m.Seqs {
		if q.K == m.OpenSeq {
			return q.Name
		}
	}
	return "sequence"
}

func (m Model) SeqNamed(k int) string {
	for _, q := range m.Seqs {
		if q.K == k {
			return q.Name
		}
	}
	return ""
}

func (m *Model) ToSeqs() {
	m.AtSeqs = true
	m.Sel = Act{Op: OpSeqOpen, I: m.OpenSeq}
}

func (q Seq) matches(find string) bool {
	find = strings.ToLower(strings.TrimSpace(find))
	return strings.Contains(strings.ToLower(q.Name), find) || strings.Contains(q.Tag(), find) || find == itoa(q.ID)
}

func (m *Model) Find(e Sheet.Edit, key string) {
	switch {
	case e.Active:
		m.SeqFind = e.Draft
	case key == "Escape":
		m.SeqFind = ""
	}
	if r, ok := m.firstSeqRow(); ok && (e.Active || key == "Enter") {
		m.Sel = r
	}
}

func (m Model) firstSeqRow() (Act, bool) {
	for _, r := range m.linkRows() {
		if r.id.Op == OpSeqOpen {
			return r.id, true
		}
	}
	return Act{}, false
}

func (m Model) FirstSeqMatch() (int, bool) {
	for _, q := range m.Seqs {
		if q.matches(m.SeqFind) {
			return q.K, true
		}
	}
	return 0, false
}

func (m Model) seqMenu() *LinksBoard.Menu {
	menu := &LinksBoard.Menu{Find: &Board.Field{Value: m.SeqFind, Fit: "type part of a name……", Act: Act{Op: OpSeqFind}, Words: true, Tip: "type part of a name: Enter opens the first match, Escape clears"},
		Empty: "no sequence name contains \"" + m.SeqFind + "\""}
	for _, q := range m.Seqs {
		if !q.matches(m.SeqFind) {
			continue
		}
		c := Board.Chip{Label: q.Tag() + "  " + q.Name, Act: Act{Op: OpSeqIn, I: q.K}}
		if q.K == m.OpenSeq {
			c.Lit, c.Tip = true, "open"
		}
		menu.Items = append(menu.Items, c)
	}
	menu.Extra = []Board.Chip{
		{Label: "＋ new sequence", Act: Act{Op: OpSeqAdd, I: m.OpenSeq}, Tip: "after this one"},
		{Label: "⧉ duplicate " + m.openSeqTag(), Act: Act{Op: OpSeqCopy, I: m.OpenSeq}},
		{Label: "☰ manage all…", Act: Act{Op: OpSeqs}},
	}
	return menu
}

func findGroup(m Model) Group {
	row := []Cell{Field("find", m.SeqFind, Act{Op: OpSeqFind}, true, false).Wording().Tipped("type part of a name: Enter selects the first match, Escape clears")}
	if m.SeqFind != "" {
		row = append(row, Button("×", Act{Op: OpSeqFindClear}, false).Tipped("clear the search"))
	}
	return Group{Rows: [][]Cell{row}}
}

func seqRows(t *tree, m Model) {
	t.put(findGroup(m))
	shown := 0
	for _, q := range m.Seqs {
		if !q.matches(m.SeqFind) {
			continue
		}
		shown++
		detail := []Cell{Text(q.Tag()).Dimmed(true).Tipped("this sequence's id · it stays the same when sequences are added, copied or deleted")}
		if q.K == m.OpenSeq {
			detail = append(detail, Text("open").Dimmed(true))
		}
		acts := []Cell{
			Button("Open", Act{Op: OpSeqIn, I: q.K}, false).Tipped("go into this sequence"),
			Button("Rename", Act{Op: OpRename, K: renameSeq, I: q.K}, false).Tipped("rename this sequence"),
			Button("Add after", Act{Op: OpSeqAdd, I: q.K}, false).Tipped("add an empty sequence after this one"),
			Button("Duplicate", Act{Op: OpSeqCopy, I: q.K}, false).Tipped("duplicate this sequence"),
		}
		if len(m.Seqs) > 1 {
			acts = append(acts, Text("·").Dimmed(true), Button("Delete", Act{Op: OpSeqDrop, I: q.K}, false).Tipped("delete this sequence"))
		}
		t.add(linkRow{
			id:     Act{Op: OpSeqOpen, I: q.K},
			in:     Act{Op: OpSeqIn, I: q.K},
			rename: Act{Op: OpRename, K: renameSeq, I: q.K},
			name:   Field("sequence "+itoa(q.ID), q.Name, Act{Op: OpSeqName, I: q.K}, true, false).Plain().Wording(),
			detail: detail,
			acts:   acts,
		})
	}
	if shown == 0 {
		t.put(Group{Rows: [][]Cell{{Text("no sequence name contains \"" + m.SeqFind + "\"").Dimmed(true)}}, Wraps: true})
	}
}
