package moverreg

import (
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/Node/Edge/edgetable"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
)

const InboxDepth = 8

type MoverRegistry struct {
	nodeGeoms map[string]*nodeactor.NodeGeometry

	edges map[string]*edgetable.Edge

	edgeOut map[string]*beadanimation.Sender

	centerMirror map[string]Vec3
}

func New() MoverRegistry {
	return MoverRegistry{
		nodeGeoms:    map[string]*nodeactor.NodeGeometry{},
		edges:        map[string]*edgetable.Edge{},
		edgeOut:      map[string]*beadanimation.Sender{},
		centerMirror: map[string]Vec3{},
	}
}

func (mr *MoverRegistry) NodeGeoms() map[string]*nodeactor.NodeGeometry {
	return mr.nodeGeoms
}

func (mr *MoverRegistry) HasNode(id string) bool {
	_, ok := mr.nodeGeoms[id]
	return ok
}

func (mr *MoverRegistry) Edges() map[string]*edgetable.Edge {
	return mr.edges
}

func (mr *MoverRegistry) SeedCenter(id string, c Vec3) {
	mr.centerMirror[id] = c
}
