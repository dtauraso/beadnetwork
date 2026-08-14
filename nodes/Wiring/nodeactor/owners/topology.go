package owners

import "github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"

type Topology struct {
	edgeIDs       []string
	neighborKinds map[string]string
	mutualTargets map[string]bool
	nodeRowFor    func(id string) (int32, bool)

	// orbitRule is THIS node's own statement of how it may sit about the node
	// it hangs from, loaded from its own meta.json by id. It belongs here
	// rather than beside a neighbour's entry because it is a fact about this
	// node, true of every edge it hangs from at once. nil means free, which is
	// what a node that says nothing about it is.
	orbitRule *polar.OrbitRule
}

func NewTopology() Topology {
	return Topology{}
}

func (t *Topology) EdgeIDs() []string { return t.edgeIDs }

func (t *Topology) NeighborKinds() map[string]string { return t.neighborKinds }

func (t *Topology) AddMutualTarget(target string) {
	if t.mutualTargets == nil {
		t.mutualTargets = map[string]bool{}
	}
	t.mutualTargets[target] = true
}

func (t *Topology) AddEdgeID(edgeID string) {
	t.edgeIDs = append(t.edgeIDs, edgeID)
}

func (t *Topology) AddNeighborKind(toID, kind string) {
	if t.neighborKinds == nil {
		t.neighborKinds = map[string]string{}
	}
	t.neighborKinds[toID] = kind
}

func (t *Topology) SetOrbitRule(rule *polar.OrbitRule) {
	t.orbitRule = rule
}

func (t *Topology) OrbitRule() *polar.OrbitRule { return t.orbitRule }

func (t *Topology) NeighborKind(toID string) string {
	return t.neighborKinds[toID]
}

func (t *Topology) IsMutualTarget(toID string) bool {
	return t.mutualTargets[toID]
}

func (t *Topology) SetNodeRowFor(fn func(id string) (int32, bool)) {
	t.nodeRowFor = fn
}

func (t *Topology) NodeRowFor(id string) (int32, bool) {
	if t.nodeRowFor == nil {
		return -1, false
	}
	return t.nodeRowFor(id)
}
