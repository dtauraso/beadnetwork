package moverreg

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/edgetable"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

const InboxDepth = 8

type MoverRegistry struct {
	nodeGeoms map[string]*nodeactor.NodeGeometry

	edges map[string]*edgetable.Edge

	edgeOut map[string]*outport.Out

	centerMirror map[string]vec3
}

func New() MoverRegistry {
	return MoverRegistry{
		nodeGeoms:    map[string]*nodeactor.NodeGeometry{},
		edges:        map[string]*edgetable.Edge{},
		edgeOut:      map[string]*outport.Out{},
		centerMirror: map[string]vec3{},
	}
}

func (mr *MoverRegistry) NodeGeoms() map[string]*nodeactor.NodeGeometry {
	return mr.nodeGeoms
}

func (mr *MoverRegistry) Edges() map[string]*edgetable.Edge {
	return mr.edges
}

func (mr *MoverRegistry) SeedCenter(id string, c vec3) {
	mr.centerMirror[id] = c
}
