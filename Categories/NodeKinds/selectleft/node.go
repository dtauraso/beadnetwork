package selectleft

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/NodeKinds/portwiring"
)

type SelectLeft struct {
	GateNode
}

func (g *SelectLeft) Update(ctx context.Context) {
	RunGateAccept(ctx, &g.GateNode, 1, 0)
}

var Builder = BuilderFor("SelectLeft",
	func(a BuildArgs) (portwiring.Node, error) {
		n := &SelectLeft{}
		n.Fire = a.Fire()
		n.EmitInputBeads = a.EmitInputBeads()
		n.Self = claimSelfDrive(a)
		n.Clock = a.Clock()
		n.SpeedCh = a.SpeedCh()
		n.FromLeft = a.In("FromLeft")
		n.FromRight = a.In("FromRight")
		n.ToPassed = a.Out("ToPassed")

		return n, nil
	})
