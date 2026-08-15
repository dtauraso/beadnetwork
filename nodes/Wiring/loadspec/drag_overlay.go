package loadspec

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/dragfile"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

func ApplyDragOverlay(root string, spec *TopoSpec) {
	for i := range spec.Nodes {
		n := &spec.Nodes[i]
		if !n.hasPoint() {
			continue
		}
		if drag, ok := dragfile.Read(root, n.ID); ok {
			d := polar.Polar{R: drag.DragPolarR, Phi: drag.DragPolarPhi, Theta: drag.DragPolarTheta}
			n.DragScenePolarR, n.DragScenePolarPhi, n.DragScenePolarTheta = &d.R, &d.Phi, &d.Theta
			n.DragIndexPhi, n.DragIndexTheta, n.DragIndexR = &drag.IndexPhi, &drag.IndexTheta, &drag.IndexR
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
