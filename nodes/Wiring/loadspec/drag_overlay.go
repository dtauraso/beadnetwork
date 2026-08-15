package loadspec

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/positionfile"
)

func ApplyDragOverlay(root string, spec *TopoSpec) {
	for i := range spec.Nodes {
		n := &spec.Nodes[i]
		if !n.hasPoint() {
			continue
		}
		if drag, ok := positionfile.Read(root, n.ID); ok {
			d := polar.Polar{R: drag.DeltaPolarR, Phi: drag.DeltaPolarPhi, Theta: drag.DeltaPolarTheta}
			n.DragScenePolarR, n.DragScenePolarPhi, n.DragScenePolarTheta = &d.R, &d.Phi, &d.Theta
		}
	}

	for i := range spec.Edges {
		e := &spec.Edges[i]
		if !e.hasDelta() {
			continue
		}
		if d, ok := edgefile.ReadEdgeDragDelta(root, e.Source, e.Label); ok {
			e.DragDeltaPolarR, e.DragDeltaPolarPhi, e.DragDeltaPolarTheta = &d.R, &d.Phi, &d.Theta
		}
	}
}
