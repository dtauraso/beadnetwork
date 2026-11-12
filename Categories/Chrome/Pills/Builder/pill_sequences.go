package Builder

import (
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
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
	Name string
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
	return strings.Contains(strings.ToLower(q.Name), find) || strings.Contains("sequence "+itoa(q.K), find)
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
		detail := []Cell{}
		if q.K == m.OpenSeq {
			detail = append(detail, Text("open").Dimmed(true))
		}
		acts := []Cell{
			Button("→", Act{Op: OpSeqIn, I: q.K}, false).Tipped("go into this sequence"),
			Button("✎", Act{Op: OpRename, K: renameSeq, I: q.K}, false).Tipped("rename this sequence"),
			Button("+", Act{Op: OpSeqAdd, I: q.K}, false).Tipped("add an empty sequence after this one"),
			Button("⧉", Act{Op: OpSeqCopy, I: q.K}, false).Tipped("duplicate this sequence"),
		}
		if len(m.Seqs) > 1 {
			acts = append(acts, Button("×", Act{Op: OpSeqDrop, I: q.K}, false).Tipped("delete this sequence"))
		}
		t.add(linkRow{
			id:     Act{Op: OpSeqOpen, I: q.K},
			in:     Act{Op: OpSeqIn, I: q.K},
			rename: Act{Op: OpRename, K: renameSeq, I: q.K},
			name:   Field("sequence "+itoa(q.K), q.Name, Act{Op: OpSeqName, I: q.K}, true, false).Plain().Wording(),
			detail: detail,
			acts:   acts,
		})
	}
	if shown == 0 {
		t.put(Group{Rows: [][]Cell{{Text("no sequence name contains \"" + m.SeqFind + "\"").Dimmed(true)}}, Wraps: true})
	}
}
