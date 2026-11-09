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

func sequenceGroups(m Model) []Group {
	shown := "sequence"
	for _, q := range m.Seqs {
		if q.K == m.OpenSeq {
			shown = q.Name
		}
	}
	caret := " ▾"
	if m.SeqListOpen {
		caret = " ▴"
	}
	row := []Cell{Text("sequence"), Button(shown+caret, Act{Op: OpSeqFold}, m.SeqListOpen), Button("+", Act{Op: OpSeqAdd}, false)}
	if len(m.Seqs) > 1 {
		row = append(row, Button("−", Act{Op: OpSeqDrop}, false))
	}
	groups := []Group{{Rows: [][]Cell{row}}}
	if !m.SeqListOpen {
		return groups
	}
	list := Group{Indent: true}
	for _, q := range m.Seqs {
		label := q.Name
		if q.K == m.OpenSeq {
			label += "  (open)"
		}
		list.Rows = append(list.Rows, []Cell{Button(label, Act{Op: OpSeqOpen, I: q.K}, q.K == m.OpenSeq)})
	}
	return append(groups, list)
}
