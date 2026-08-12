package kindapi

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/nodeapi"
	"github.com/dtauraso/wirefold/nodes/wire/outport"

	T "github.com/dtauraso/wirefold/Trace"
)

type BuildDeps struct {
	LatticePoints int32

	ClaimLatticeIn func(name string) chan int32

	ClaimTiltEditIn func(name string) chan movemsg.TiltEditMsg

	ClaimSelfDriveGeom func(name string) *nodeactor.NodeGeometry
}

type BuildArgs struct {
	ctx  context.Context
	name string
	data *loadspec.NodeData
	pb   portwiring.PortBindings
	tr   *T.Trace
	geom nodegeom.NodeGeom

	tiltThetaIdx int32

	sourceOuts *[]*outport.Out

	getStream func() *interior.InteriorStream

	driveSlotClaims map[int]string

	deps BuildDeps
}

func (a BuildArgs) Name() string { return a.name }

func (a BuildArgs) Ctx() context.Context { return a.ctx }

func RegisterBuilder(kind string, ports []portwiring.PortSpec, build func(BuildArgs) (nodeapi.Node, error)) {
	if _, exists := Registry[kind]; exists {
		panic("kindapi.RegisterBuilder: kind already registered: " + kind)
	}
	Registry[kind] = NodeBuilder{
		Ports: ports,
		Build: func(ctx context.Context, name string, data *loadspec.NodeData, pb portwiring.PortBindings, tr *T.Trace, geom nodegeom.NodeGeom, tiltThetaIdx int32, deps BuildDeps) (nodeapi.Node, error) {
			var sourceOuts []*outport.Out
			return build(BuildArgs{
				ctx: ctx, name: name, data: data, pb: pb, tr: tr,
				geom:            geom,
				sourceOuts:      &sourceOuts,
				getStream:       portwiring.NewInteriorStreamGetter(name, pb),
				driveSlotClaims: map[int]string{},
				tiltThetaIdx:    tiltThetaIdx,
				deps:            deps,
			})
		},
	}
}
