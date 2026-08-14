package kindreg

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/nodeapi"

	T "github.com/dtauraso/wirefold/Trace"
)

type BuildDeps struct {
	LatticePoints int32

	ClaimLatticeIn func(name string) chan int32

	ClaimTiltEditIn func(name string) chan movemsg.TiltEditMsg

	ClaimSelfDriveGeom func(name string) *nodeactor.NodeGeometry
}

type NodeBuilder struct {
	Ports []portwiring.PortSpec
	Build func(ctx context.Context, name string, data *loadspec.NodeData, pb portwiring.PortBindings, tr *T.Trace, geom nodegeom.NodeGeom, tiltPhiIdx int32, deps BuildDeps) (nodeapi.Node, error)
}

var Registry map[string]NodeBuilder

func init() {
	Registry = make(map[string]NodeBuilder)
}

func BuildRegistry() {
	if len(Registry) == 0 {
		panic("kindreg.BuildRegistry: no node kinds registered — kinds_generated.go's blank imports are what run each kind's init(); regenerate with `go run ./tools/gen-node-defs`")
	}
}
