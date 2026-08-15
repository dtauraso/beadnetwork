package loadspec

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

func ResolveEdgeDeltas(spec *TopoSpec) {
	points := make(map[string]polar.Polar, len(spec.Nodes))
	for i := range spec.Nodes {
		if n := &spec.Nodes[i]; n.hasPoint() {
			points[n.ID] = n.point()
		}
	}

	for i := range spec.Edges {
		e := &spec.Edges[i]
		if e.hasDelta() {
			continue
		}
		a, okA := points[e.Source]
		b, okB := points[e.Target]
		if !okA || !okB {
			continue
		}
		e.setDelta(polar.Between(a, b))
	}
}

func PlaceFromDeltas(spec *TopoSpec) {
	out := make(map[string][]*specEdge, len(spec.Nodes))
	for i := range spec.Edges {
		e := &spec.Edges[i]
		out[e.Source] = append(out[e.Source], e)
	}

	placed := make(map[string]bool, len(spec.Nodes))
	byID := make(map[string]*specNode, len(spec.Nodes))
	for i := range spec.Nodes {
		byID[spec.Nodes[i].ID] = &spec.Nodes[i]
	}

	for i := range spec.Nodes {
		seed := &spec.Nodes[i]
		if placed[seed.ID] || !seed.hasPoint() {
			continue
		}
		placed[seed.ID] = true
		for queue := []*specNode{seed}; len(queue) > 0; {
			from := queue[0]
			queue = queue[1:]
			for _, e := range out[from.ID] {
				to, known := byID[e.Target]
				if !known || placed[e.Target] || !e.hasDelta() {
					continue
				}
				to.setPoint(polar.Compose(from.point(), e.delta()))
				placed[e.Target] = true
				queue = append(queue, to)
			}
		}
	}
}

func (n *specNode) hasPoint() bool {
	return n.ScenePolarR != nil && n.ScenePolarPhi != nil && n.ScenePolarTheta != nil
}

func (n *specNode) point() polar.Polar {
	return polar.Polar{R: *n.ScenePolarR, Phi: *n.ScenePolarPhi, Theta: *n.ScenePolarTheta}
}

func (n *specNode) setPoint(p polar.Polar) {
	n.ScenePolarR, n.ScenePolarPhi, n.ScenePolarTheta = &p.R, &p.Phi, &p.Theta
}

func (e specEdge) Delta() (polar.Polar, bool) {
	if !e.hasDelta() {
		return polar.Polar{}, false
	}
	return e.delta(), true
}

func (e specEdge) BaseDelta() (polar.Polar, bool) {
	return e.Delta()
}

func (e specEdge) DragDelta() polar.Polar {
	if e.DragDeltaPolarR == nil || e.DragDeltaPolarPhi == nil || e.DragDeltaPolarTheta == nil {
		return polar.Polar{}
	}
	return polar.Polar{R: *e.DragDeltaPolarR, Phi: *e.DragDeltaPolarPhi, Theta: *e.DragDeltaPolarTheta}
}

func (e *specEdge) hasDelta() bool {
	return e.DeltaPolarR != nil && e.DeltaPolarPhi != nil && e.DeltaPolarTheta != nil
}

func (e *specEdge) delta() polar.Polar {
	return polar.Polar{R: *e.DeltaPolarR, Phi: *e.DeltaPolarPhi, Theta: *e.DeltaPolarTheta}
}

func (e *specEdge) setDelta(d polar.Polar) {
	e.DeltaPolarR, e.DeltaPolarPhi, e.DeltaPolarTheta = &d.R, &d.Phi, &d.Theta
}
