package loadspec

import (
	"github.com/dtauraso/wirefold/Polar/polarindex"
)

func ResolveEdgeDeltas(spec *TopoSpec) {
	indices := make(map[string]polarindex.Index, len(spec.Nodes))
	for i := range spec.Nodes {
		if n := &spec.Nodes[i]; n.hasPoint() {
			indices[n.ID] = n.index()
		}
	}

	for i := range spec.Edges {
		e := &spec.Edges[i]
		if e.hasDelta() {
			continue
		}
		a, okA := indices[e.Source]
		b, okB := indices[e.Target]
		if !okA || !okB {
			continue
		}
		e.setDeltaIndex(polarindex.Delta(b, a))
	}
}

func PlaceFromDeltas(spec *TopoSpec) {
	out := make(map[string][]*Edge, len(spec.Nodes))
	for i := range spec.Edges {
		e := &spec.Edges[i]
		out[e.Source] = append(out[e.Source], e)
	}

	placed := make(map[string]bool, len(spec.Nodes))
	byID := make(map[string]*Node, len(spec.Nodes))
	for i := range spec.Nodes {
		byID[spec.Nodes[i].ID] = &spec.Nodes[i]
	}

	for i := range spec.Nodes {
		seed := &spec.Nodes[i]
		if placed[seed.ID] || !seed.hasPoint() {
			continue
		}
		placed[seed.ID] = true
		for queue := []*Node{seed}; len(queue) > 0; {
			from := queue[0]
			queue = queue[1:]
			for _, e := range out[from.ID] {
				to, known := byID[e.Target]
				if !known || placed[e.Target] || to.hasPoint() || !e.hasDelta() {
					continue
				}
				to.setIndex(polarindex.Compose(from.index(), e.deltaIndex(), spec.Constants))
				placed[e.Target] = true
				queue = append(queue, to)
			}
		}
	}
}

func (n *Node) hasPoint() bool {
	return n.IndexPhi != nil && n.IndexTheta != nil && n.IndexR != nil
}

func (n *Node) index() polarindex.Index {
	return polarindex.Index{Phi: *n.IndexPhi, Theta: *n.IndexTheta, R: *n.IndexR}
}

func (n *Node) setIndex(idx polarindex.Index) {
	n.IndexPhi, n.IndexTheta, n.IndexR = &idx.Phi, &idx.Theta, &idx.R
}

func (e Edge) BaseDeltaIndex() (polarindex.Offset, bool) {
	if !e.hasDelta() {
		return polarindex.Offset{}, false
	}
	return e.deltaIndex(), true
}

func (e Edge) DragDeltaIndex() polarindex.Offset {
	if e.DragDeltaIndexR == nil || e.DragDeltaIndexPhi == nil || e.DragDeltaIndexTheta == nil {
		return polarindex.Offset{}
	}
	return polarindex.Offset{Phi: *e.DragDeltaIndexPhi, Theta: *e.DragDeltaIndexTheta, R: *e.DragDeltaIndexR}
}

func (e *Edge) hasDelta() bool {
	return e.DeltaIndexR != nil && e.DeltaIndexPhi != nil && e.DeltaIndexTheta != nil
}

func (e *Edge) deltaIndex() polarindex.Offset {
	return polarindex.Offset{Phi: *e.DeltaIndexPhi, Theta: *e.DeltaIndexTheta, R: *e.DeltaIndexR}
}

func (e *Edge) setDeltaIndex(off polarindex.Offset) {
	e.DeltaIndexPhi, e.DeltaIndexTheta, e.DeltaIndexR = &off.Phi, &off.Theta, &off.R
}
