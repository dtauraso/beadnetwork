package selectright

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/NodeKinds/nodeapi"

	"github.com/dtauraso/wirefold/Categories/NodeKinds/gatecommon"
	Wiring "github.com/dtauraso/wirefold/Categories/NodeKinds/kindapi"
)

type SelectRight struct {
	gatecommon.GateNode
}

func (g *SelectRight) Update(ctx context.Context) {
	gatecommon.RunGateAccept(ctx, &g.GateNode, 0, 1)
}

var Builder = Wiring.BuilderFor("SelectRight",
	func(a Wiring.BuildArgs) (nodeapi.Node, error) {
		n := &SelectRight{}
		n.Fire = a.Fire()
		n.EmitInputBeads = a.EmitInputBeads()
		n.Self = a.ClaimSelfDrive()
		n.Clock = a.Clock()
		n.SpeedCh = a.SpeedCh()
		n.FromLeft = a.In("FromLeft")
		n.FromRight = a.In("FromRight")
		n.ToPassed = a.Out("ToPassed")

		return n, nil
	})
