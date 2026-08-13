package owners

import "github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"

type Topology struct {
	edgeIDs []string

	// neighborPaths is this node's OWN polar vector to each neighbour — for
	// an outward edge it is that edge's animation path, the line its beads
	// travel. It is a cached path, never a coordinate: nothing derives a
	// node's position from a vector another node stores.
	//
	// It is POLAR, not cartesian, because every constraint that has ever
	// been placed on a path is a statement about its two angles or its
	// length — the outgoing angles an input node may hold, the step
	// constants a length quantizes to. Holding x/y/z meant each of those
	// read the angles out and wrote them back, which is a round trip
	// through trigonometry to change a number that was already stored.
	// Here the angles ARE the field, and the trig sits only where a world
	// position or a streamed direction is genuinely wanted.
	neighborPaths map[string]polar.Polar

	// sharedOutLen is the ONE length this node's outgoing paths all have,
	// when its kind holds them equal. It is stored rather than derived
	// because the paths are what it decides: asking them what the shared
	// length is, right after setting them from it, would only ever agree
	// with itself and could never pull a path that had drifted.
	//
	// Zero means no length has been established yet.
	sharedOutLen float64

	// outAngleFixes counts consecutive angle corrections per target that
	// have not yet come back in range. It sits with the paths because it
	// is a fact about them: how many times one has been rewritten without
	// the rewrite taking.
	outAngleFixes map[string]int

	neighborKinds map[string]string
	mutualTargets map[string]bool
	nodeRowFor    func(id string) (int32, bool)
}

func NewTopology(neighborPaths map[string]polar.Polar) Topology {
	return Topology{neighborPaths: neighborPaths}
}

func (t *Topology) EdgeIDs() []string { return t.edgeIDs }

// PathTo is the stored path to that neighbour as a cartesian vector — the
// boundary conversion, for the callers that draw or measure it.
func (t *Topology) PathTo(neighborID string) (vec3, bool) {
	p, ok := t.neighborPaths[neighborID]
	if !ok {
		return vec3{}, false
	}
	return polar.Polar2cart(p), true
}

// PolarPathTo is the stored path as it is actually kept, for the callers that
// want to read or constrain its angles without a round trip.
func (t *Topology) PolarPathTo(neighborID string) (polar.Polar, bool) {
	p, ok := t.neighborPaths[neighborID]
	return p, ok
}

// PartnerCenters DERIVES each neighbour's world centre from the stored path.
// The path is what is kept; the centre is a sum computed on demand.
func (t *Topology) PartnerCenters(selfCenter vec3) map[string]vec3 {
	out := make(map[string]vec3, len(t.neighborPaths))
	for id, path := range t.neighborPaths {
		out[id] = selfCenter.Add(polar.Polar2cart(path))
	}
	return out
}

// SetPathTo rewrites the one path whose destination moved.
func (t *Topology) SetPathTo(neighborID string, selfCenter, neighborCenter vec3) {
	t.SetPolarPathTo(neighborID, polar.Cart2polar(neighborCenter.Sub(selfCenter)))
}

// SetPolarPathTo rewrites one path from its angles and length directly, for
// the case where the path itself is what changed and no neighbour centre was
// quoted. The centre this implies is the sum PartnerCenters already computes.
func (t *Topology) SetPolarPathTo(neighborID string, p polar.Polar) {
	if t.neighborPaths == nil {
		t.neighborPaths = map[string]polar.Polar{}
	}
	t.neighborPaths[neighborID] = p
}

// SharedOutLen is the length every outgoing path is held at, and whether one
// has been established.
func (t *Topology) SharedOutLen() (float64, bool) {
	return t.sharedOutLen, t.sharedOutLen != 0
}

// SetSharedOutLen declares the length the outgoing paths are held at from now
// on. The mover that just changed its distance is the one that sets it — the
// node that was dragged states the new length and the others are brought to
// it, rather than the drag being undone.
func (t *Topology) SetSharedOutLen(r float64) {
	t.sharedOutLen = r
}

// BumpOutAngleFix records one more correction to that target and returns the
// running count. ClearOutAngleFix forgets it, which is what a path arriving
// already in range means.
func (t *Topology) BumpOutAngleFix(neighborID string) int {
	if t.outAngleFixes == nil {
		t.outAngleFixes = map[string]int{}
	}
	t.outAngleFixes[neighborID]++
	return t.outAngleFixes[neighborID]
}

// HasPendingOutAngleFix says whether that target has been corrected and has
// not yet reported back in range — that is, whether the position it is about
// to report is one THIS node asked for.
func (t *Topology) HasPendingOutAngleFix(neighborID string) bool {
	return t.outAngleFixes[neighborID] > 0
}

func (t *Topology) ClearOutAngleFix(neighborID string) {
	delete(t.outAngleFixes, neighborID)
}

// RebaseForSelfMove rigidly rebases every path when THIS node moves: the
// neighbours did not move, so each path shifts by the node's own delta.
//
// This is the one operation on a path that is genuinely cartesian — a
// translation is a sum of components, and there is no polar form of it — so
// the conversion out and back is the real cost of the move, not a round trip
// standing in for an assignment.
func (t *Topology) RebaseForSelfMove(prevSelf, newSelf vec3) {
	delta := prevSelf.Sub(newSelf)
	for id, path := range t.neighborPaths {
		t.neighborPaths[id] = polar.Cart2polar(polar.Polar2cart(path).Add(delta))
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
