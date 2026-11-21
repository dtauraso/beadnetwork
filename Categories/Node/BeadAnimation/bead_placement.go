package beadanimation

type BeadPlacement struct {
	Steps int

	SlotR float64

	Start, End Vec3
	Node, Port string
}

func (bp BeadPlacement) streams() bool {
	return bp.Node != ""
}

type placeRequest struct {
	val           int
	bp            BeadPlacement
	placementTick int64
}

type inflightBead struct {
	val  int
	slot int

	seg     Segment
	steps   int
	slotR   float64
	node    string
	port    string
	streams bool
	gen     uint64
}

func (bl *BeadLine) arrived(b *inflightBead) bool {
	return b.slot >= b.steps
}

func (b *inflightBead) pos() Vec3 {
	dir := b.seg.End.Sub(b.seg.Start)
	if dir.Length() < 1e-9 {
		return b.seg.Start
	}
	return b.seg.Start.Add(dir.Normalize().Scale(float64(b.slot) * b.slotR))
}
