package wire

import "github.com/dtauraso/wirefold/nodes/spatial"

// BeadPlacement is exported: it is constructed by nodes/wire/outport.Out and handed
// across the package boundary to PacedWire.Send. Nothing else about it changed.
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
	seg     spatial.WireSegment
	node    string
	port    string
	streams bool
	gen     uint64
}

func (pw *PacedWire) arrived(b *inflightBead) bool {
	return b.slot >= b.steps*pw.slotsPerBead()
}
