package Topology

import (
	"fmt"
	"os"

	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"

	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgefile"
	"github.com/dtauraso/beadnetwork/Categories/Vector/polarindex"
)

func ResolveEdgeDeltas(spec *TopoSpec) {
	indices := make(map[string]polarindex.Index, len(spec.Nodes))
	for i := range spec.Nodes {
		if n := &spec.Nodes[i]; n.HasPoint() {
			indices[n.ID] = n.Index()
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
		if placed[seed.ID] || !seed.HasPoint() {
			continue
		}
		placed[seed.ID] = true
		for queue := []*Node{seed}; len(queue) > 0; {
			from := queue[0]
			queue = queue[1:]
			for _, e := range out[from.ID] {
				to, known := byID[e.Target]
				if !known || placed[e.Target] || to.HasPoint() || !e.hasDelta() {
					continue
				}
				to.setIndex(polarindex.Compose(from.Index(), e.deltaIndex(), spec.Constants))
				placed[e.Target] = true
				queue = append(queue, to)
			}
		}
	}
}

func (n *Node) HasPoint() bool {
	return n.IndexPhi != nil && n.IndexTheta != nil && n.IndexR != nil
}

func (n *Node) Index() polarindex.Index {
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

func reportEdgeClosure(spec *TopoSpec) {
	if err := checkEdgeClosure(spec); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func checkEdgeClosure(spec *TopoSpec) error {
	idx := make(map[string]polarindex.Index, len(spec.Nodes))
	for i := range spec.Nodes {
		if n := &spec.Nodes[i]; n.HasPoint() {
			idx[n.ID] = polarindex.Canonical(n.Index(), spec.Constants)
		}
	}

	for i := range spec.Edges {
		e := &spec.Edges[i]
		if !e.hasDelta() {
			continue
		}
		src, okS := idx[e.Source]
		dst, okT := idx[e.Target]
		if !okS || !okT {
			continue
		}
		got := polarindex.Compose(src, e.deltaIndex(), spec.Constants)
		if got != dst {
			return fmt.Errorf(
				"loadTree: edge %q does not close: node %s index %+v composed with delta %+v gives %+v, but node %s is at %+v — an edge delta must be the exact index difference of its endpoints (target minus source), so the edge is drawn to where the node actually is",
				e.Label, e.Source, src, e.deltaIndex(), got, e.Target, dst)
		}
	}
	return nil
}
