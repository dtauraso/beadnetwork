package beadanimation

import "github.com/dtauraso/wirefold/src/spatial"

type BeadPlacement struct {
	Steps int

	SlotR float64

	Start, End spatial.Vec3
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

	seg     spatial.Segment
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
