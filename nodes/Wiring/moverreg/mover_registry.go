package moverreg

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

const InboxDepth = 8

type MoverRegistry struct {
	nodeGeoms map[string]*nodeactor.NodeGeometry

	nodeMovers map[string]*nodeactor.NodeMover

	selfDriveClaimed map[string]bool
	edgeMovers       map[string]*edgemover.EdgeMover

	edgeOut map[string]*wire.Out

	centerMirror map[string]vec3
}

func New() MoverRegistry {
	return MoverRegistry{
		nodeGeoms:    map[string]*nodeactor.NodeGeometry{},
		edgeMovers:   map[string]*edgemover.EdgeMover{},
		edgeOut:      map[string]*wire.Out{},
		centerMirror: map[string]vec3{},
	}
}

func (mr *MoverRegistry) NodeGeoms() map[string]*nodeactor.NodeGeometry {
	return mr.nodeGeoms
}

func (mr *MoverRegistry) EdgeMovers() map[string]*edgemover.EdgeMover {
	return mr.edgeMovers
}

func (mr *MoverRegistry) ClaimSelfDrive(id string) {
	if mr.selfDriveClaimed == nil {
		mr.selfDriveClaimed = map[string]bool{}
	}
	mr.selfDriveClaimed[id] = true
}

func (mr *MoverRegistry) SeedCenter(id string, c vec3) {
	mr.centerMirror[id] = c
}
