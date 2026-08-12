package selectleft

import (
	"context"
	"github.com/dtauraso/wirefold/nodes/nodeapi"

	Wiring "github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/gatecommon"
)

type SelectLeft struct {
	gatecommon.GateNode
}

func (g *SelectLeft) Update(ctx context.Context) {
	gatecommon.RunGateAccept(ctx, &g.GateNode, 1, 0)
}

func init() {

	Wiring.RegisterBuilder("SelectLeft",
		[]portwiring.PortSpec{
			{Name: "FromLeft", Dir: portwiring.PortIn},
			{Name: "FromRight", Dir: portwiring.PortIn},
			{Name: "ToPassed", Dir: portwiring.PortOut},
		},
		func(a Wiring.BuildArgs) (nodeapi.Node, error) {
			n := &SelectLeft{}
			n.Fire = a.Fire()
			n.EmitInputBeads = a.EmitInputBeads()
			n.Tick = a.Tick()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.FromLeft = a.In("FromLeft")
			n.FromRight = a.In("FromRight")
			n.ToPassed = a.Out("ToPassed")

			return n, nil
		})
}
