// node_registry.go — the loader-facing kind registry: NodeBuilder is the
// public type consumed by the loader, and BuildRegistry populates Registry
// from wire.KindRegistry as each node package's init() runs.

package Wiring

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// NodeBuilder is the public-facing type consumed by the loader.
// Ports is derived lazily from reflection; Build delegates to reflectBuild.
type NodeBuilder struct {
	Ports []PortSpec
	Build func(ctx context.Context, name string, data *NodeData, pb PortBindings, tr *T.Trace, geom nodeGeom, partnerCenter partnerCenterFn) (wire.Node, error)
}

// Registry is the loader-facing map, populated one kind at a time by
// Register (registry.go) as each node package's init() runs.
var Registry map[string]NodeBuilder

func init() {
	Registry = make(map[string]NodeBuilder)
}

// BuildRegistry populates Registry from KindRegistry for any kind not yet
// built. KindRegistry is filled by wire.Register as each node package's
// init() runs; building the NodeBuilder (reflectPorts + reflectBuild closure) is
// deferred to here, the loader's entry point, so wire.Register itself has no
// dependency on the build pipeline. Idempotent — safe to call on every load. Must
// run before any code reads Registry (validateSpec, buildFromSpec). Exported so
// package-main tests (which never call LoadTopology) can force population, e.g.
// kind_registry_parity_test.go.
func BuildRegistry() {
	for kind, newNode := range wire.KindRegistry {
		if _, ok := Registry[kind]; ok {
			continue
		}
		newNode := newNode // capture for closure
		sample := newNode()
		ports := reflectPorts(sample)
		Registry[kind] = NodeBuilder{
			Ports: ports,
			Build: func(ctx context.Context, name string, data *NodeData, pb PortBindings, tr *T.Trace, geom nodeGeom, partnerCenter partnerCenterFn) (wire.Node, error) {
				return reflectBuild(ctx, name, data, pb, newNode, tr, geom, partnerCenter)
			},
		}
	}
}
