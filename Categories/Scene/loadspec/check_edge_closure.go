package loadspec

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func reportEdgeClosure(spec *TopoSpec) {
	if err := checkEdgeClosure(spec); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func checkEdgeClosure(spec *TopoSpec) error {
	idx := make(map[string]polarindex.Index, len(spec.Nodes))
	for i := range spec.Nodes {
		if n := &spec.Nodes[i]; n.hasPoint() {
			idx[n.ID] = polarindex.Canonical(n.index(), spec.Constants)
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
