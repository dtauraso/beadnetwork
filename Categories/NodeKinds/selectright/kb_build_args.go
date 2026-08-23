package selectright

import (
	"context"
	"strings"

	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	interior "github.com/dtauraso/wirefold/Categories/Node/Interior"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/kindreg"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/nodeapi"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/portwiring"
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
)

type BuildArgs struct {
	Ctx  context.Context
	Name string
	Data *loadspec.NodeData
	PB   portwiring.PortBindings

	TiltPhiIdx int32

	sourceOuts *[]*beadanimation.Sender

	getEmitter func() *interior.Emitter

	Deps kindreg.BuildDeps

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

	panic("mustDeclare: this kind binds a port its SPEC.md ## Ports table does not declare with that direction, which would silently bind a dead-end channel: kind " +
		a.kind + " asked for " + portName + ", table declares " + strings.Join(names, ", "))
}

func BuilderFor(kind string, build func(BuildArgs) (nodeapi.Node, error)) kindreg.NodeBuilder {
	ports, declared := portwiring.KindPorts[kind]
	if !declared {

		panic("BuilderFor: kind " + kind + " has no ports in portwiring.KindPorts — " +
			"its SPEC.md ## Ports table is the only declaration, and the generated table is stale. " +
			"Run go generate ./...")
	}
	return kindreg.NodeBuilder{
		Ports: ports,
		Build: func(ctx context.Context, name string, data *loadspec.NodeData, pb portwiring.PortBindings, tiltPhiIdx int32, deps kindreg.BuildDeps) (nodeapi.Node, error) {
			var sourceOuts []*beadanimation.Sender
			return build(BuildArgs{
				Ctx: ctx, Name: name, Data: data, PB: pb,
				kind:  kind,
				ports: ports,

				sourceOuts: &sourceOuts,
				getEmitter: portwiring.NewInteriorEmitterGetter(name, pb),
				TiltPhiIdx: tiltPhiIdx,
				Deps:       deps,
			})
		},
	}
}
