package kindapi

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/nodeapi"

	T "github.com/dtauraso/wirefold/Trace"
)

type NodeBuilder struct {
	Ports []portwiring.PortSpec
	Build func(ctx context.Context, name string, data *loadspec.NodeData, pb portwiring.PortBindings, tr *T.Trace, geom nodegeom.NodeGeom, tiltThetaIdx int32, deps BuildDeps) (nodeapi.Node, error)
}

var Registry map[string]NodeBuilder

func init() {
	Registry = make(map[string]NodeBuilder)
}

func BuildRegistry() {
	if len(Registry) == 0 {
		panic("kindapi.BuildRegistry: no node kinds registered — kinds_generated.go's blank imports are what run each kind's init(); regenerate with `go run ./tools/gen-node-defs`")
	}
}
