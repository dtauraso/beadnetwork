package kindapi

import (
	"context"

	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/Node/Interior"
	"github.com/dtauraso/wirefold/src/Node/Wiring/kindreg"
	"github.com/dtauraso/wirefold/src/Node/Wiring/loadspec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/src/Node/Wiring/portwiring"
	"github.com/dtauraso/wirefold/src/NodeKinds/nodeapi"
)

type BuildArgs struct {
	ctx  context.Context
	name string
	data *loadspec.NodeData
	pb   portwiring.PortBindings
	geom nodegeom.NodeGeom

	tiltPhiIdx int32

	sourceOuts *[]*beadanimation.Sender

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
		Build: func(ctx context.Context, name string, data *loadspec.NodeData, pb portwiring.PortBindings, geom nodegeom.NodeGeom, tiltPhiIdx int32, deps kindreg.BuildDeps) (nodeapi.Node, error) {
			var sourceOuts []*beadanimation.Sender
			return build(BuildArgs{
				ctx: ctx, name: name, data: data, pb: pb,
				geom:       geom,
				sourceOuts: &sourceOuts,
				getEmitter: portwiring.NewInteriorEmitterGetter(name, pb),
				tiltPhiIdx: tiltPhiIdx,
				deps:       deps,
			})
		},
	}
}
