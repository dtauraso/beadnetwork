package Builder

const (
	OpSeqOpen = "seq-open"
	OpSeqAdd  = "seq-add"
	OpSeqDrop = "seq-drop"
	OpSeqFold = "seq-fold"
)

type Seq struct {
	K    int
	Name string
}

func sequenceGroup(m Model) Group {
	g := Group{Title: foldTitle(m.SeqFolded, "sequence"), Fold: Act{Op: OpSeqFold}}
	if m.SeqFolded {
		return g
	}
	for _, q := range m.Seqs {
		label := q.Name
		if q.K == m.OpenSeq {
			label += "  (open)"
		}
		g.Rows = append(g.Rows, []Cell{Button(label, Act{Op: OpSeqOpen, I: q.K}, q.K == m.OpenSeq)})
	}
	buttons := []Cell{Button("+ new sequence", Act{Op: OpSeqAdd}, false)}
	if len(m.Seqs) > 1 {
		buttons = append(buttons, Button("delete sequence", Act{Op: OpSeqDrop}, false))
	}
	g.Rows = append(g.Rows, buttons)
	return g
}
