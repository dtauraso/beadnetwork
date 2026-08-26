package NodePhiTheta

import (
	"context"

	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
)

type deps interface {
	SelfDriveGeom(name string) any
}

type BuildArgs struct {
	Ctx  context.Context
	Name string
	PB   bindings

	Deps deps
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

func (b kindBuilder) Build(ctx context.Context, name string, _ *NodeBuf.NodeData, pb any, _ int32, bd any) (any, error) {
	bound, ok := pb.(bindings)
	if !ok {
		panic("Build: the scene handed " + name + " something that is not this kind's port bindings")
	}
	dep, okd := bd.(deps)
	if !okd {
		panic("Build: the scene handed " + name + " something that is not this kind's build deps")
	}
	return b.build(BuildArgs{Ctx: ctx, Name: name, PB: bound, Deps: dep})
}

func BuilderFor(kind string, build func(BuildArgs) (any, error)) kindBuilder {
	return kindBuilder{kind: kind, build: build}
}
