package windowandinhibitrightgate

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/gatecommon"
)

// SelectLeft is the "WindowAndInhibitRightGate" kind (registered as
// "WindowAndInhibitRightGate" below — the name lives here in the comment,
// describing what its functions do). Its functions: Update runs the shared gate
// loop (gatecommon.RunGateAccept), which collects both inputs in the coincidence
// window and accepts the raw 10 pattern — FromLeft==1 AND FromRight==0 — DIRECTLY,
// no inversion/NOT. This package owns only the struct layout (required for
// gen-node-defs port discovery) and the init registration; GateNode is embedded so
// its port fields (FromLeft, FromRight, ToPassed) are promoted and discovered by
// reflectPorts.
type SelectLeft struct {
	gatecommon.GateNode
}

func (g *SelectLeft) Update(ctx context.Context) {
	gatecommon.RunGateAccept(ctx, &g.GateNode, 1, 0)
}

func init() {
	wire.Register("WindowAndInhibitRightGate", func() any { return &SelectLeft{} })
}
