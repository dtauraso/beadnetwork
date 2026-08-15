package owners

import "github.com/dtauraso/wirefold/nodes/Wiring/polarindex"

type Quant struct {
	base      polarindex.Index
	drag      polarindex.Index
	constants polarindex.SceneConstants
}

func (q *Quant) SetConstants(sc polarindex.SceneConstants) { q.constants = sc }

func (q *Quant) Constants() polarindex.SceneConstants { return q.constants }

func (q *Quant) SetBase(off polarindex.Index) { q.base = off }

func (q *Quant) SetDrag(off polarindex.Index) { q.drag = off }

func (q *Quant) Base() polarindex.Index { return q.base }

func (q *Quant) Drag() polarindex.Index { return q.drag }

func (q *Quant) Composed() polarindex.Index {
	return polarindex.Compose(q.base, q.drag)
}
