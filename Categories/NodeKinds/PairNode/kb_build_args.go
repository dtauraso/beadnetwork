package PairNode

import (
	"context"
	"strings"

	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	interior "github.com/dtauraso/wirefold/Categories/Node/Interior"
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
)

type deps interface {
	LatticePointsSeed() int32
	LatticeChan(name string) chan int32
	TiltEditChan(name string) any
	SelfDriveGeom(name string) any
}

type BuildArgs struct {
	Ctx  context.Context
	Name string
	Data *loadspec.NodeData
	PB   bindings

	TiltPhiIdx int32

	sourceOuts *[]*beadanimation.Sender

	getEmitter func() *interior.Emitter

	Deps deps

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

type kindBuilder struct {
	kind  string
	build func(BuildArgs) (any, error)
}

func (b kindBuilder) Ports() []struct {
	Name string
	Dir  int
} {
	out := make([]struct {
		Name string
		Dir  int
	}, len(kindPorts))
	for i, p := range kindPorts {
		out[i].Name, out[i].Dir = p.Name, int(p.Dir)
	}
	return out
}

func (b kindBuilder) Build(ctx context.Context, name string, data *loadspec.NodeData, pb any, tiltPhiIdx int32, bd any) (any, error) {
	bound, ok := pb.(bindings)
	if !ok {
		panic("Build: the scene handed " + name + " something that is not this kind's port bindings")
	}
	dep, okd := bd.(deps)
	if !okd {
		panic("Build: the scene handed " + name + " something that is not this kind's build deps")
	}
	var sourceOuts []*beadanimation.Sender
	return b.build(BuildArgs{
		Ctx: ctx, Name: name, Data: data, PB: bound,
		kind:  b.kind,
		ports: kindPorts,

		sourceOuts: &sourceOuts,
		getEmitter: NewInteriorEmitterGetter(name, bound),
		TiltPhiIdx: tiltPhiIdx,
		Deps:       dep,
	})
}

func BuilderFor(kind string, build func(BuildArgs) (any, error)) kindBuilder {
	if len(kindPorts) == 0 {
		panic("BuilderFor: kind " + kind + " has no ports — its SPEC.md ## Ports table is the " +
			"only declaration, and this kind's generated table is stale. Run go generate ./...")
	}
	return kindBuilder{kind: kind, build: build}
}
