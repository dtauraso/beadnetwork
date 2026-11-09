package Builder

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"

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
	return []Group{{Rows: [][]Cell{row}}}
}

func sequenceList(m Model, at int) []Group {
	if !m.SeqListOpen || len(m.Seqs) == 0 {
		return nil
	}
	list := Group{Card: true, Pop: &Sheet.Anchor{Group: at, Col: 1}}
	for _, q := range m.Seqs {
		mark := "   "
		if q.K == m.OpenSeq {
			mark = "✓ "
		}
		list.Rows = append(list.Rows, []Cell{Button(mark+q.Name, Act{Op: OpSeqOpen, I: q.K}, q.K == m.OpenSeq)})
	}
	return []Group{list}
}
