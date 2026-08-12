package owners

type Topology struct {
	edgeIDs []string

	partnerCenters map[string]vec3
	neighborKinds  map[string]string
	mutualTargets  map[string]bool
	nodeRowFor     func(id string) (int32, bool)
}

func NewTopology(partnerCenters map[string]vec3) Topology {
	return Topology{partnerCenters: partnerCenters}
}

func (t *Topology) EdgeIDs() []string { return t.edgeIDs }

func (t *Topology) PartnerCenters() map[string]vec3 { return t.partnerCenters }

func (t *Topology) NeighborKinds() map[string]string { return t.neighborKinds }

func (t *Topology) AddMutualTarget(target string) {
	if t.mutualTargets == nil {
		t.mutualTargets = map[string]bool{}
	}
	t.mutualTargets[target] = true
}

func (t *Topology) SeedPartnerCenter(neighborID string, c vec3) {
	if t.partnerCenters == nil {
		t.partnerCenters = map[string]vec3{}
	}
	t.partnerCenters[neighborID] = c
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
