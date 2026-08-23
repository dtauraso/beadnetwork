package selectright

import (
	"context"
)

type SelectRight struct {
	GateNode
}

func (g *SelectRight) Update(ctx context.Context) {
	RunGateAccept(ctx, &g.GateNode, 0, 1)
}

var Builder = BuilderFor("SelectRight",
	func(a BuildArgs) (any, error) {
		n := &SelectRight{}
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
