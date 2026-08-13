package loadspec

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

// A node is a point and an edge is the polar vector that closes the triangle:
//
//	A = scene centre -> source     the source's own point
//	D = source -> target           the edge's vector
//	B = scene centre -> target     the target's point
//
//	A + D = B
//
// D is the side that closes it, and storing it is what lets a node reach its
// out-neighbour without reading anything the neighbour owns.
//
// The three are VECTORS, each with its own owner: A belongs to the node, B to
// the neighbour, D to the EDGE. D is a polar vector about its source, on the
// same +y pole as everything else — its phi the bearing from the source's own
// top pole, its theta the turn around it, its r the distance — which is what
// makes the out-angle constraints statements about D itself.
//
// A + D = B is not a rule anything enforces or a formula anything computes from.
// It is what vector addition MAKES TRUE, and it is why local and global are the
// same thing: the node's own account of where its neighbour is IS the
// neighbour's own point.

// ResolveEdgeDeltas gives D to an edge whose file does not carry one yet.
//
// It is a MIGRATION, and only that: a scene authored before edges held their own
// vector has the value nowhere else, so it is taken once from the two endpoints
// and belongs to the edge from then on. An edge that already carries D is left
// exactly as it is — the file is the authority, and nothing recomputes it.
//
// There is nothing here to check. A + D = B is not a rule to enforce, it is what
// vector addition makes true: a node's own statement about where its neighbour
// is and the neighbour's own point are the same value, not two values that could
// disagree. This function used to derive D from the endpoints on EVERY load and
// then assert the triangle closed to within 1e-6 — computing state the edge owns,
// and then checking an identity. The assertion could only ever fire on a bug
// somewhere else, while naming the edge file as the culprit.
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

// PlaceFromDeltas gives every node its point by walking outward along the edges:
// a source's point plus its edge's D IS its target's point. The stored point of a
// node no edge reaches — a node nothing points at, or the first node of a ring —
// is what seeds its component, and is the only point read as final.
//
// A node's stored point and the point the walk arrives at are the same value —
// that is the identity, not a coincidence to be verified — so this pass is how
// the edge's own vector becomes what places its neighbour.
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

// Delta is this edge's D — the vector from its source to its target — once the
// loader has read or derived it.
func (e specEdge) Delta() (polar.Polar, bool) {
	if !e.hasDelta() {
		return polar.Polar{}, false
	}
	return e.delta(), true
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
