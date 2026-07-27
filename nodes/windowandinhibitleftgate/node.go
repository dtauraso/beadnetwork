package windowandinhibitleftgate

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/gatecommon"
)

// SelectRight is the "WindowAndInhibitLeftGate" kind (registered as
// "WindowAndInhibitLeftGate" below — the name lives here in the comment,
// describing what its functions do). Its functions: Update runs the shared gate
// loop (gatecommon.RunGate) with invertLeft=true — the LEFT input is NOT-inverted
// on capture (1→0, 0→1), so the gate fires 1 iff (NOT left) AND right. This
// package owns only the struct layout (required for gen-node-defs port discovery)
// and the init registration; GateNode is embedded so its port fields (FromLeft,
// FromRight, ToPassed) are promoted and discovered by reflectPorts.
type SelectRight struct {
	gatecommon.GateNode
}

func (g *SelectRight) Update(ctx context.Context) {
	gatecommon.RunGate(ctx, &g.GateNode, true /* invertLeft */)
}

func init() {
	wire.Register("WindowAndInhibitLeftGate", func() any { return &SelectRight{} })
}
