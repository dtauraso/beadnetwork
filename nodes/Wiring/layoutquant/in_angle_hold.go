package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

// TrimDragAgainstInNeighbors keeps only the part of a drag those same
// constraints allow: `from` is where the node is, `target` where the drag is
// asking to put it, and the result is where it actually goes.
//
// A drag is a delta, so it is trimmed rather than overwritten. The component
// a constraint pins contributes nothing to the move — the node does not go
// there and get pulled back, it simply never goes there — while every
// component with room still moves the full amount. Dragging a node held to an
// input's equator slides it around that equator instead of refusing to move.
//
// It is the neighbours whose edge ENTERS this node that constrain it. An
// out-target is a node this one points at, whose angles are that node's own
// business — and the constraining node holds those itself.
func TrimDragAgainstInNeighbors(nm *nodeactor.NodeGeometry, from, target spatial.Vec3) spatial.Vec3 {
	kinds := nm.NeighborKinds()
	if len(kinds) == 0 {
		return target
	}
	centers := nm.PartnerCenters()
	for neighborID, kind := range kinds {
		if kind != nodeactor.OutAngleKind || nm.IsOutTarget(neighborID) {
			continue
		}
		center, ok := centers[neighborID]
		if !ok {
			continue
		}
		have := polar.Cart2polar(from.Sub(center))
		want := polar.Cart2polar(target.Sub(center))
		target = center.Add(polar.Polar2cart(polar.TrimOutAngleDelta(have, want)))
	}
	return target
}
