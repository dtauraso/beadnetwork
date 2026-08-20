package kindapi

import (
	"context"
	"strings"

	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/Node/Interior"
	"github.com/dtauraso/wirefold/src/NodeKinds/kindreg"
	"github.com/dtauraso/wirefold/src/runtopology/loadspec"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
	"github.com/dtauraso/wirefold/src/NodeKinds/portwiring"
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

	kind  string
	ports []portwiring.PortSpec
}

func (a BuildArgs) mustDeclare(portName string, dir portwiring.PortDir) {
	for _, p := range a.ports {
		if p.Name == portName && p.Dir == dir {
			return
		}
	}
	names := make([]string, 0, len(a.ports))
	for _, p := range a.ports {
		names = append(names, p.Name)
	}

	panic("kindapi.mustDeclare: this kind binds a port its SPEC.md ## Ports table does not declare with that direction, which would silently bind a dead-end channel: kind " +
		a.kind + " asked for " + portName + ", table declares " + strings.Join(names, ", "))
}

func (a BuildArgs) Name() string { return a.name }

func (a BuildArgs) Ctx() context.Context { return a.ctx }

func RegisterBuilder(kind string, build func(BuildArgs) (nodeapi.Node, error)) {
	if _, exists := kindreg.Registry[kind]; exists {
		panic("kindapi.RegisterBuilder: kind already registered: " + kind)
	}

	ports, declared := portwiring.KindPorts[kind]
	if !declared {

		panic("kindapi.RegisterBuilder: kind " + kind + " has no ports in portwiring.KindPorts — " +
			"its SPEC.md ## Ports table is the only declaration, and the generated table is stale. " +
			"Run go generate ./...")
	}
	kindreg.Registry[kind] = kindreg.NodeBuilder{
		Ports: ports,
		Build: func(ctx context.Context, name string, data *loadspec.NodeData, pb portwiring.PortBindings, geom nodegeom.NodeGeom, tiltPhiIdx int32, deps kindreg.BuildDeps) (nodeapi.Node, error) {
			var sourceOuts []*beadanimation.Sender
			return build(BuildArgs{
				ctx: ctx, name: name, data: data, pb: pb,
				kind:       kind,
				ports:      ports,
				geom:       geom,
				sourceOuts: &sourceOuts,
				getEmitter: portwiring.NewInteriorEmitterGetter(name, pb),
				tiltPhiIdx: tiltPhiIdx,
				deps:       deps,
			})
		},
	}
}
