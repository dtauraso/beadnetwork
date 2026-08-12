package nodeactor

func (t *neighborTopology) EdgeIDs() []string { return t.edgeIDs }

func (t *neighborTopology) PartnerCenters() map[string]vec3 { return t.partnerCenters }

func (t *neighborTopology) NeighborKinds() map[string]string { return t.neighborKinds }

func (t *neighborTopology) AddMutualTarget(target string) {
	if t.mutualTargets == nil {
		t.mutualTargets = map[string]bool{}
	}
	t.mutualTargets[target] = true
}

func (t *neighborTopology) SeedPartnerCenter(neighborID string, c vec3) {
	if t.partnerCenters == nil {
		t.partnerCenters = map[string]vec3{}
	}
	t.partnerCenters[neighborID] = c
}

func (t *neighborTopology) AddEdgeID(edgeID string) {
	t.edgeIDs = append(t.edgeIDs, edgeID)
}

func (t *neighborTopology) AddNeighborKind(toID, kind string) {
	if t.neighborKinds == nil {
		t.neighborKinds = map[string]string{}
	}
	t.neighborKinds[toID] = kind
}
