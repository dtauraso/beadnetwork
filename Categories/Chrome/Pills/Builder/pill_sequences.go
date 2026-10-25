package Builder

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
		label := q.Name
		if q.K == m.OpenSeq {
			label += "  (open)"
		}
		rows = append(rows, []Cell{Button(label, Act{Op: OpSeqOpen, I: q.K}, q.K == m.OpenSeq)})
	}
	buttons := []Cell{Button("+ new sequence", Act{Op: OpSeqAdd}, false)}
	if len(m.Seqs) > 1 {
		buttons = append(buttons, Button("delete sequence", Act{Op: OpSeqDrop}, false))
	}
	rows = append(rows, buttons)
	groups := []Group{{Title: "sequence — click one to open it; the scene reloads with its nodes and links", Rows: rows}}
	if m.Refused != "" {
		groups = append(groups, Group{Rows: [][]Cell{{Text(m.Refused)}}})
	}
	return Pill{Label: SeqLabel, Left: true, Groups: groups}
}
