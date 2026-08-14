package kindapi

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	"github.com/dtauraso/wirefold/nodes/Wiring/kindreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/nodeapi"
	"github.com/dtauraso/wirefold/nodes/wire/outport"

	T "github.com/dtauraso/wirefold/Trace"
)

type BuildArgs struct {
	ctx  context.Context
	name string
	data *loadspec.NodeData
	pb   portwiring.PortBindings
	tr   *T.Trace
	geom nodegeom.NodeGeom

	tiltPhiIdx int32

	sourceOuts *[]*outport.Out

	getEmitter func() *interior.Emitter

	deps kindreg.BuildDeps
}

func (a BuildArgs) Name() string { return a.name }

func (a BuildArgs) Ctx() context.Context { return a.ctx }

func RegisterBuilder(kind string, ports []portwiring.PortSpec, build func(BuildArgs) (nodeapi.Node, error)) {
	if _, exists := kindreg.Registry[kind]; exists {
		panic("kindapi.RegisterBuilder: kind already registered: " + kind)
	}
	kindreg.Registry[kind] = kindreg.NodeBuilder{
		Ports: ports,
		Build: func(ctx context.Context, name string, data *loadspec.NodeData, pb portwiring.PortBindings, tr *T.Trace, geom nodegeom.NodeGeom, tiltPhiIdx int32, deps kindreg.BuildDeps) (nodeapi.Node, error) {
			var sourceOuts []*outport.Out
			return build(BuildArgs{
				ctx: ctx, name: name, data: data, pb: pb, tr: tr,
				geom:       geom,
				sourceOuts: &sourceOuts,
				getEmitter: portwiring.NewInteriorEmitterGetter(name, pb),
				tiltPhiIdx: tiltPhiIdx,
				deps:       deps,
			})
		},
	}
}
