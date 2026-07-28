package selectright

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/gatecommon"
)

// SelectRight is the "SelectRight" kind (registered as
// "SelectRight" below — the name lives here in the comment,
// describing what its functions do). Its functions: Update runs the shared gate
// loop (gatecommon.RunGateAccept), which accepts the raw 01 pattern —
// FromLeft==0 AND FromRight==1 — directly, no inversion/NOT gates. Inputs are
// captured raw (no NOT), and the gate fires 1 iff the raw stored values are
// exactly Left==0 && Right==1. This package owns only the struct layout
// (required for gen-node-defs port discovery) and the init registration;
// GateNode is embedded so its port fields (FromLeft, FromRight, ToPassed) are
// promoted and discovered by reflectPorts.
type SelectRight struct {
	gatecommon.GateNode
}

func (g *SelectRight) Update(ctx context.Context) {
	gatecommon.RunGateAccept(ctx, &g.GateNode, 0, 1)
}

func init() {
	wire.Register("SelectRight", func() any { return &SelectRight{} })
}
