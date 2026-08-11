// node_registry.go — the loader-facing kind registry: NodeBuilder is the
// public type consumed by the loader, and BuildRegistry populates Registry
// from wire.KindRegistry as each node package's init() runs.

package kindapi

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// NodeBuilder is the public-facing type consumed by the loader (nodes/Wiring/dispatch).
//
// This type lives in kindapi, not nodes/Wiring/loadspec: Build's signature carries
// BuildDeps (build_args.go) — see that type's own doc comment for why it no longer names
// any dispatch-core type, which is what let this whole package (and NodeBuilder with it)
// move out of the dispatch core in the first place (§24). It still cannot live in
// loadspec: loadspec.ValidateSpec, the one loadspec function that used to read this
// registry, takes the specific values it needs (each kind's port list) as a parameter
// instead — see its own doc comment — rather than this package importing loadspec's own
// importers back.
type NodeBuilder struct {
	Ports []portwiring.PortSpec
	Build func(ctx context.Context, name string, data *loadspec.NodeData, pb portwiring.PortBindings, tr *T.Trace, geom nodegeom.NodeGeom, tiltThetaIdx int32, deps BuildDeps) (wire.Node, error)
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
		panic("kindapi.BuildRegistry: no node kinds registered — kinds_generated.go's blank imports are what run each kind's init(); regenerate with `go run ./tools/gen-node-defs`")
	}
}
