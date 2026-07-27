package windowandinhibitrightgate

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/gatecommon"
)

// SelectLeft is the "WindowAndInhibitRightGate" kind (registered as
// "WindowAndInhibitRightGate" below — the name lives here in the comment,
// describing what its functions do). Its functions: Update runs the shared gate
// loop (gatecommon.RunGate) with invertLeft=false — the RIGHT input is NOT-inverted
// on capture (1→0, 0→1), so the gate fires 1 iff left AND (NOT right). This
// package owns only the struct layout (required for gen-node-defs port discovery)
// and the init registration; GateNode is embedded so its port fields (FromLeft,
// FromRight, ToPassed) are promoted and discovered by reflectPorts.
type SelectLeft struct {
	gatecommon.GateNode
}

func (g *SelectLeft) Update(ctx context.Context) {
	gatecommon.RunGate(ctx, &g.GateNode, false /* invertLeft */)
}

func init() {
	wire.Register("WindowAndInhibitRightGate", func() any { return &SelectLeft{} })
}
