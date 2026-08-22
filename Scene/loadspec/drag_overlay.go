package loadspec

import (
	"github.com/dtauraso/wirefold/Node/Edge/edgefile"
	"github.com/dtauraso/wirefold/Node/dragfile"
)

func ApplyDragOverlay(root string, spec *TopoSpec) {
	for i := range spec.Nodes {
		n := &spec.Nodes[i]
		if !n.hasPoint() {
			continue
		}
		if drag, ok := dragfile.Read(root, n.ID); ok {
			n.DragIndexPhi, n.DragIndexTheta, n.DragIndexR = &drag.IndexPhi, &drag.IndexTheta, &drag.IndexR
		}
	}

	for i := range spec.Edges {
		e := &spec.Edges[i]
		if !e.hasDelta() {
			continue
		}
		if dragIdx, ok := edgefile.ReadEdgeDragIndex(root, e.Source, e.Label); ok {
			e.DragDeltaIndexPhi, e.DragDeltaIndexTheta, e.DragDeltaIndexR = &dragIdx.Phi, &dragIdx.Theta, &dragIdx.R
		}
	}
}
