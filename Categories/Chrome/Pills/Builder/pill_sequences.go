package Builder

const (
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

func seqRows(t *tree, m Model) {
	for _, q := range m.Seqs {
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
}
