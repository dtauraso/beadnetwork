package NodePhiTheta3

import (
	"context"

	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

type deps interface {
	SelfDriveGeom(name string) any
	LinkInbox(name string) any
	LinkMesh() any
	LinkSeed() any
}

type BuildArgs struct {
	Ctx  context.Context
	Name string
	PB   bindings
	Data *NodeBuf.NodeData

	Deps deps
}

func (a BuildArgs) Inbox() Links.Inbox {
	in, ok := a.Deps.LinkInbox(a.Name).(Links.Inbox)
	if !ok {
		panic("NodePhiTheta3.Inbox: the scene handed node " + a.Name + " something that is not a links inbox — Startup.BuildDeps.ClaimLinkIn should have made one")
	}
	return in
}

func (a BuildArgs) Mesh() *Links.Mesh {
	m, ok := a.Deps.LinkMesh().(*Links.Mesh)
	if !ok || m == nil {
		panic("NodePhiTheta3.Mesh: the scene handed node " + a.Name + " no links mesh — Startup.BuildDeps.Mesh is how a node reaches the others")
	}
	return m
}

func (a BuildArgs) Seed() Links.Seed {
	s, ok := a.Deps.LinkSeed().(Links.Seed)
	if !ok {
		panic("NodePhiTheta3.Seed: the scene handed node " + a.Name + " no links seed — Startup.BuildDeps.Seed holds every node's place and link as loaded")
	}
	return s
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

func (b kindBuilder) Build(ctx context.Context, name string, data *NodeBuf.NodeData, pb any, _ int32, bd any) (any, error) {
	bound, ok := pb.(bindings)
	if !ok {
		panic("Build: the scene handed " + name + " something that is not this kind's port bindings")
	}
	dep, okd := bd.(deps)
	if !okd {
		panic("Build: the scene handed " + name + " something that is not this kind's build deps")
	}
	return b.build(BuildArgs{Ctx: ctx, Name: name, PB: bound, Data: data, Deps: dep})
}

func BuilderFor(kind string, build func(BuildArgs) (any, error)) kindBuilder {
	return kindBuilder{kind: kind, build: build}
}
