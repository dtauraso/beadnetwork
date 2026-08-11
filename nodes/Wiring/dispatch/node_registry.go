// node_registry.go — the loader-facing kind registry: NodeBuilder is the
// public type consumed by the loader, and BuildRegistry populates Registry
// from wire.KindRegistry as each node package's init() runs.

package dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// NodeBuilder is the public-facing type consumed by the loader.
// Ports and Build both come from the kind itself, via RegisterBuilder.
//
// This type stays in nodes/Wiring rather than moving with the rest of the loader/spec
// pieces into nodes/Wiring/loadspec: Build's signature carries buildDeps
// (build_args.go), a Wiring-internal hub type (it holds *nodeInboxes/*moverRegistry) that
// loadspec cannot name without importing nodes/Wiring — the exact cycle "no package under
// nodes/Wiring/ may import nodes/Wiring" forbids. loadspec.ValidateSpec, which is the one
// loadspec function that used to read this registry, takes the specific values it needs
// (each kind's port list) as a parameter instead — see its own doc comment.
type NodeBuilder struct {
	Ports []portwiring.PortSpec
	Build func(ctx context.Context, name string, data *loadspec.NodeData, pb portwiring.PortBindings, tr *T.Trace, geom nodegeom.NodeGeom, tiltThetaIdx int32, deps buildDeps) (wire.Node, error)
}

// Registry is the loader-facing map, populated one kind at a time by
// Register (registry.go) as each node package's init() runs.
var Registry map[string]NodeBuilder

func init() {
	Registry = make(map[string]NodeBuilder)
}

// BuildRegistry is retained as the loader's explicit "registry is ready" call site, but it
// no longer BUILDS anything: every kind now registers itself in its own init() via
// RegisterBuilder (build_args.go), so Registry is fully populated before main runs. It
// stays because the loader and kind_registry_parity_test call it, and because a kind that
// forgets to register is better diagnosed here than at the first unknown-type error.
//
// It panics on an EMPTY registry: that means no kind package's init() ran at all, which in
// practice means kinds_generated.go lost its blank imports — the exact failure the
// primitive landing rule warns about, and one that otherwise surfaces much later as
// `unknown type "X"`.
func BuildRegistry() {
	if len(Registry) == 0 {
		panic("Wiring.BuildRegistry: no node kinds registered — kinds_generated.go's blank imports are what run each kind's init(); regenerate with `go run ./tools/gen-node-defs`")
	}
}
