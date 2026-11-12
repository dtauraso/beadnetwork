package Builder

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"

const (
	OpSeqOpen = "seq-open"
	OpSeqAdd  = "seq-add"
	OpSeqDrop = "seq-drop"
	OpSeqFold = "seq-fold"
)

const seqListRows = 6

type Seq struct {
	K    int
	Name string
}

func (m Model) seqName() string {
	for _, q := range m.Seqs {
		if q.K == m.OpenSeq {
			return q.Name
		}
	}
	return "sequence"
}

func sequenceCell(m Model) Cell {
	if len(m.Seqs) == 0 {
		return Button(m.seqName(), Act{Op: OpLevelGo, I: stepIn(m.Trail, 0)}, m.Level == "").Tipped("go to the sequence's own steps")
	}
	caret := " ▾"
	if m.SeqListOpen {
		caret = " ▴"
	}
	return Button(m.seqName()+caret, Act{Op: OpSeqFold}, m.Level == "" || m.SeqListOpen).Tipped("pick a sequence, or pick this one to go to its own steps")
}

func sequenceList(m Model, at int) []Group {
	if !m.SeqListOpen || len(m.Seqs) == 0 {
		return nil
	}
	list := Group{Card: true, Pop: &Sheet.Anchor{Group: at, Col: 0}, MaxRows: seqListRows}
	for _, q := range m.Seqs {
		mark := "   "
		if q.K == m.OpenSeq {
			mark = "✓ "
		}
		list.Rows = append(list.Rows, []Cell{Button(mark+q.Name, Act{Op: OpSeqOpen, I: q.K}, q.K == m.OpenSeq)})
	}
	list.Rows = append(list.Rows, []Cell{Button("+  copy this sequence", Act{Op: OpSeqAdd}, false)})
	if len(m.Seqs) > 1 {
		list.Rows = append(list.Rows, []Cell{Button("−  delete "+m.seqName(), Act{Op: OpSeqDrop}, false)})
	}
	return []Group{list}
}
