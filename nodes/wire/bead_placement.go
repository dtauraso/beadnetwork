package wire

type beadPlacement struct {
	Steps int

	Start, End Vec3
	Node, Port string
}

func (bp beadPlacement) streams() bool {
	return bp.Node != ""
}

type placeRequest struct {
	val           int
	bp            beadPlacement
	placementTick int64
}

type inflightBead struct {
	val           int
	placementTick float64
	steps         int
	seg           WireSegment
	node          string
	port          string
	streams       bool
	gen           uint64

	finalPending bool
}
