package bead

import "github.com/dtauraso/wirefold/nodes/spatial"

type BeadPlacement struct {
	Steps int

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

	steps   int
	seg     spatial.Segment
	node    string
	port    string
	streams bool
	gen     uint64
}

func (pw *BeadRun) arrived(b *inflightBead) bool {
	return b.slot >= b.steps*pw.slotsPerBead()
}
