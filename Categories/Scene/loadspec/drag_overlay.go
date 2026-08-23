package loadspec

import (
	NodeBuf "github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Node/Edge/edgefile"
)

func ApplyDragOverlay(root string, spec *TopoSpec) {
	for i := range spec.Nodes {
		n := &spec.Nodes[i]
		if !n.HasPoint() {
			continue
		}
		if phi, theta, r, _, ok := NodeBuf.ReadDragIndex(root, n.ID); ok {
			n.DragIndexPhi, n.DragIndexTheta, n.DragIndexR = &phi, &theta, &r
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
