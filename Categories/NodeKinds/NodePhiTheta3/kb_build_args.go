package NodePhiTheta3

import (
	"context"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
)

type deps interface {
	SelfDriveGeom(name string) any
	CardEditChan(name string) any
	CardSSeed() int
}

type BuildArgs struct {
	Ctx  context.Context
	Name string
	PB   bindings
	Data *NodeBuf.NodeData

	Deps deps
}

func (a BuildArgs) EditIn() <-chan CardPanel.EditMsg {
	if a.Deps == nil {
		return nil
	}
	ch, ok := a.Deps.CardEditChan(a.Name).(chan CardPanel.EditMsg)
	if !ok {
		panic("NodePhiTheta3.EditIn: the scene handed node " + a.Name + " something that is not a card-edit channel — Startup.BuildDeps.ClaimCardEditIn should have made one")
	}
	return ch
}

func (a BuildArgs) State() map[string]int {
	if a.Data == nil {
		return nil
	}
	return a.Data.State
}

func (a BuildArgs) S() int {
	if a.Deps == nil || a.Deps.CardSSeed() < 1 {
		return CardPanel.DefaultS
	}
	return a.Deps.CardSSeed()
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
