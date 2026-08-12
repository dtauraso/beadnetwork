package owners

type Topology struct {
	edgeIDs []string

	// neighborPaths is this node's OWN vector to each neighbour — for an
	// outward edge it is that edge's animation path, the line its beads
	// travel. It is a cached path, never a coordinate: nothing derives a
	// node's position from a vector another node stores.
	neighborPaths map[string]vec3
	neighborKinds map[string]string
	mutualTargets map[string]bool
	nodeRowFor    func(id string) (int32, bool)
}

func NewTopology(neighborPaths map[string]vec3) Topology {
	return Topology{neighborPaths: neighborPaths}
}

func (t *Topology) EdgeIDs() []string { return t.edgeIDs }

// PathTo is the stored vector from this node's centre to that neighbour.
func (t *Topology) PathTo(neighborID string) (vec3, bool) {
	p, ok := t.neighborPaths[neighborID]
	return p, ok
}

// PartnerCenters DERIVES each neighbour's world centre from the stored path.
// The path is what is kept; the centre is a sum computed on demand.
func (t *Topology) PartnerCenters(selfCenter vec3) map[string]vec3 {
	out := make(map[string]vec3, len(t.neighborPaths))
	for id, path := range t.neighborPaths {
		out[id] = selfCenter.Add(path)
	}
	return out
}

// SetPathTo rewrites the one path whose destination moved.
func (t *Topology) SetPathTo(neighborID string, selfCenter, neighborCenter vec3) {
	if t.neighborPaths == nil {
		t.neighborPaths = map[string]vec3{}
	}
	t.neighborPaths[neighborID] = neighborCenter.Sub(selfCenter)
}

// RebaseForSelfMove rigidly rebases every path when THIS node moves: the
// neighbours did not move, so each path shifts by the node's own delta. Exact
// arithmetic on the stored vector — never a reconstruction from a live world.
func (t *Topology) RebaseForSelfMove(prevSelf, newSelf vec3) {
	delta := prevSelf.Sub(newSelf)
	for id, path := range t.neighborPaths {
		t.neighborPaths[id] = path.Add(delta)
	}
}

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
