package PairNode

import (
	"context"
	"strings"

	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	interior "github.com/dtauraso/wirefold/Categories/Node/Interior"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/portwiring"
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
)

type BuildArgs struct {
	Ctx  context.Context
	Name string
	Data *loadspec.NodeData
	PB   bindings

	TiltPhiIdx int32

	sourceOuts *[]*beadanimation.Sender

	getEmitter func() *interior.Emitter

	Deps portwiring.BuildDeps

	kind  string
	ports []PortSpec
}

func (a BuildArgs) mustDeclare(portName string, dir PortDir) {
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

func BuilderFor(kind string, build func(BuildArgs) (portwiring.Node, error)) portwiring.NodeBuilder {
	if len(kindPorts) == 0 {
		panic("BuilderFor: kind " + kind + " has no ports — its SPEC.md ## Ports table is the " +
			"only declaration, and this kind's generated table is stale. Run go generate ./...")
	}
	ports := make([]portwiring.PortSpec, len(kindPorts))
	for i, p := range kindPorts {
		ports[i] = portwiring.PortSpec{Name: p.Name, Dir: portwiring.PortDir(p.Dir)}
	}
	return portwiring.NodeBuilder{
		Ports: ports,
		Build: func(ctx context.Context, name string, data *loadspec.NodeData, pb any, tiltPhiIdx int32, deps portwiring.BuildDeps) (portwiring.Node, error) {
			bound, ok := pb.(bindings)
			if !ok {
				panic("BuilderFor: the scene handed " + name + " something that is not this kind's port bindings")
			}
			var sourceOuts []*beadanimation.Sender
			return build(BuildArgs{
				Ctx: ctx, Name: name, Data: data, PB: bound,
				kind:  kind,
				ports: kindPorts,

				sourceOuts: &sourceOuts,
				getEmitter: NewInteriorEmitterGetter(name, bound),
				TiltPhiIdx: tiltPhiIdx,
				Deps:       deps,
			})
		},
	}
}
