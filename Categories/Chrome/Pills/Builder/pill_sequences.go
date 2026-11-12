package Builder

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"

const (
	OpSeqOpen = "seq-open"
	OpSeqAdd  = "seq-add"
	OpSeqDrop = "seq-drop"
	OpSeqFold = "seq-fold"
	OpSeqName = "seq-name"
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

func sequenceCells(m Model) []Cell {
	if len(m.Seqs) == 0 {
		return []Cell{Text(m.seqName()).Dimmed(true)}
	}
	return []Cell{Text(m.seqName()).Acting(Act{Op: OpSeqFold}).Dimmed(true).Tipped("pick, rename, copy or delete a sequence")}
}

func sequenceList(m Model, at int) []Group {
	if !m.SeqListOpen || len(m.Seqs) == 0 {
		return nil
	}
	list := Group{Card: true, Pop: &Sheet.Anchor{Group: at, Col: 0}, MaxRows: seqListRows}
	for _, q := range m.Seqs {
		if q.K == m.OpenSeq {
			list.Rows = append(list.Rows, []Cell{Field("", q.Name, Act{Op: OpSeqName, I: q.K}, true, false).Plain().Wording().Tipped("the open sequence; click to rename it")})
			continue
		}
		list.Rows = append(list.Rows, []Cell{LitText(q.Name, Act{Op: OpSeqOpen, I: q.K}, false).Tipped("open this sequence")})
	}
	list.Rows = append(list.Rows, []Cell{Button("+  copy this sequence", Act{Op: OpSeqAdd}, false)})
	if len(m.Seqs) > 1 {
		list.Rows = append(list.Rows, []Cell{Button("−  delete "+m.seqName(), Act{Op: OpSeqDrop}, false)})
	}
	return []Group{list}
}
