package Sheet

const SeqLabel = "Sequences"

const (
	OpSeqOpen = "seq-open"
	OpSeqAdd  = "seq-add"
	OpSeqDrop = "seq-drop"
)

type Seq struct {
	K    int
	Name string
}

func SequencesPill(m Model) Pill {
	var rows [][]Cell
	for _, q := range m.Seqs {
		rows = append(rows, []Cell{LitText(q.Name, Act{Op: OpSeqOpen, I: q.K}, q.K == m.OpenSeq)})
	}
	buttons := []Cell{Button("+ new sequence", Act{Op: OpSeqAdd}, false)}
	if len(m.Seqs) > 1 {
		buttons = append(buttons, Button("delete sequence", Act{Op: OpSeqDrop}, false))
	}
	rows = append(rows, buttons)
	return Pill{Label: SeqLabel, Left: true, Groups: []Group{
		{Title: "sequence — click one to open it", Rows: rows},
	}}
}
