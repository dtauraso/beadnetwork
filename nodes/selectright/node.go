package selectright

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/nodeapi"

	Wiring "github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/gatecommon"
)

type SelectRight struct {
	gatecommon.GateNode
}

func (g *SelectRight) Update(ctx context.Context) {
	gatecommon.RunGateAccept(ctx, &g.GateNode, 0, 1)
}

func init() {

	Wiring.RegisterBuilder("SelectRight",
		[]portwiring.PortSpec{
			{Name: "FromLeft", Dir: portwiring.PortIn},
			{Name: "FromRight", Dir: portwiring.PortIn},
			{Name: "ToPassed", Dir: portwiring.PortOut},
		},
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
}
