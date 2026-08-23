package kindreg

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Node/nodeinbox"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/nodeapi"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/portwiring"
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
)

type BuildDeps struct {
	LatticePoints int32

	ClaimLatticeIn func(name string) chan int32

	ClaimTiltEditIn func(name string) chan nodeinbox.TiltEditMsg

	ClaimSelfDriveGeom func(name string) *nodeactor.NodeGeometry
}

type NodeBuilder struct {
	Ports []portwiring.PortSpec
	Build func(ctx context.Context, name string, data *loadspec.NodeData, pb portwiring.PortBindings, geom nodegeom.NodeGeom, tiltPhiIdx int32, deps BuildDeps) (nodeapi.Node, error)
}

var Registry map[string]NodeBuilder

func init() {
	Registry = make(map[string]NodeBuilder)
}

func BuildRegistry() {
	if len(Registry) == 0 {
		panic("kindreg.BuildRegistry: no node kinds registered — the blank imports in Categories/NodeKinds/kinds_gen.go run each kind's init(), and something must import NodeKinds for them to take effect at all; regenerate with `go generate ./...`")
	}
}
