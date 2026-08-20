package selectleft

import (
	"context"

	"github.com/dtauraso/wirefold/src/NodeKinds/nodeapi"

	Wiring "github.com/dtauraso/wirefold/src/Node/Wiring/kindapi"
	"github.com/dtauraso/wirefold/src/NodeKinds/gatecommon"
)

type SelectLeft struct {
	gatecommon.GateNode
}

func (g *SelectLeft) Update(ctx context.Context) {
	gatecommon.RunGateAccept(ctx, &g.GateNode, 1, 0)
}

func init() {

	Wiring.RegisterBuilder("SelectLeft",
		func(a Wiring.BuildArgs) (nodeapi.Node, error) {
			n := &SelectLeft{}
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
}
