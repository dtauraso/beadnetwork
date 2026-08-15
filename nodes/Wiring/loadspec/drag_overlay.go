package loadspec

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/dragfile"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
)

func ApplyDragOverlay(root string, spec *TopoSpec) {
	for i := range spec.Nodes {
		n := &spec.Nodes[i]
		if !n.hasPoint() {
			continue
		}
		if drag, ok := dragfile.Read(root, n.ID); ok {
			dragIdx := polarindex.Index{Phi: drag.IndexPhi, Theta: drag.IndexTheta, R: drag.IndexR}
			base := n.point(spec.Constants)
			composed := polarindex.ToPolar(polarindex.Compose(n.index(), dragIdx, spec.Constants), spec.Constants)
			d := polar.Between(base, composed)
			n.DragScenePolarR, n.DragScenePolarPhi, n.DragScenePolarTheta = &d.R, &d.Phi, &d.Theta
			n.DragIndexPhi, n.DragIndexTheta, n.DragIndexR = &drag.IndexPhi, &drag.IndexTheta, &drag.IndexR
		}
	}

	for i := range spec.Edges {
		e := &spec.Edges[i]
		if !e.hasDelta() {
			continue
		}
		if dragIdx, ok := edgefile.ReadEdgeDragIndex(root, e.Source, e.Label); ok {
			base := e.delta(spec.Constants)
			composedIdx := addIndex(e.deltaIndex(), dragIdx)
			composed := polarindex.ToPolar(composedIdx, spec.Constants)
			d := polar.Between(base, composed)
			e.DragDeltaPolarR, e.DragDeltaPolarPhi, e.DragDeltaPolarTheta = &d.R, &d.Phi, &d.Theta
		}
	}
}

func addIndex(a, b polarindex.Index) polarindex.Index {
	return polarindex.Index{Phi: a.Phi + b.Phi, Theta: a.Theta + b.Theta, R: a.R + b.R}
}
