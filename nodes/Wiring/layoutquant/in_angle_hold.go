package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

// HoldAgainstInNeighbors is the same outgoing-angle constraint an input node
// keeps on its own paths, applied from the far end: a node being moved will
// not COMMIT a position that puts it off the angles an input node upstream of
// it allows.
//
// It exists so the constraint is met before anything is drawn. Held only on
// the input node's side, a drag committed to a position outside the allowed
// angles renders there, the input node hears about it on the next neighbour
// broadcast, and the correction arrives a frame later — the node visibly goes
// somewhere it is not allowed to be and is then yanked back. Resolving the
// destination here means the position that gets committed was never invalid,
// so there is nothing to yank.
//
// The constraint itself is not restated: this calls the same ClampOutAngles
// the input node's own hold calls. What differs is only which node is asking
// and when.
func HoldAgainstInNeighbors(nm *nodeactor.NodeGeometry, pos spatial.Vec3) spatial.Vec3 {
	kinds := nm.NeighborKinds()
	if len(kinds) == 0 {
		return pos
	}
	centers := nm.PartnerCenters()
	for neighborID, kind := range kinds {
		if kind != nodeactor.OutAngleKind {
			continue
		}
		// An out-target is a node this one points AT; its angles are that
		// node's business, not a constraint on where this one may sit.
		if nm.IsOutTarget(neighborID) {
			continue
		}
		center, ok := centers[neighborID]
		if !ok {
			continue
		}
		want := polar.ClampOutAngles(polar.Cart2polar(pos.Sub(center)))
		pos = center.Add(polar.Polar2cart(want))
	}
	return pos
}
