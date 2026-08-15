package owners

import "github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"

type Quant struct {
	base quantoffset.QuantizedOffset
	drag quantoffset.QuantizedOffset
}

func (q *Quant) SetBase(off quantoffset.QuantizedOffset) { q.base = off }

func (q *Quant) SetDrag(off quantoffset.QuantizedOffset) { q.drag = off }

func (q *Quant) Base() quantoffset.QuantizedOffset { return q.base }

func (q *Quant) Drag() quantoffset.QuantizedOffset { return q.drag }

func (q *Quant) Composed() quantoffset.QuantizedOffset {
	return quantoffset.Compose(q.base, q.drag)
}
