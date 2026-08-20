package bead

import (
	"github.com/dtauraso/wirefold/src/Node/spatial"
)

func (b *inflightBead) pos() spatial.Vec3 {
	dir := b.seg.End.Sub(b.seg.Start)
	if dir.Length() < 1e-9 {
		return b.seg.Start
	}
	return b.seg.Start.Add(dir.Normalize().Scale(float64(b.slot) * b.slotR))
}
