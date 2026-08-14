package owners

import "github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"

type Topology struct {
	edgeIDs       []string
	neighborKinds map[string]string
	mutualTargets map[string]bool
	nodeRowFor    func(id string) (int32, bool)

	orbitRule   *polar.OrbitRule
	orbitActive bool
}

func NewTopology() Topology {
	return Topology{orbitActive: true}
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

func (t *Topology) SetOrbitActive(active bool) {
	t.orbitActive = active
}

func (t *Topology) OrbitActive() bool { return t.orbitActive }

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
