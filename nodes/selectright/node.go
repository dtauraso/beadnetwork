package selectright

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
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
	// SelectRight CONSTRUCTS ITSELF. Every assignment below was previously performed by
	// Wiring.reflectBuild via reflection over the embedded gatecommon.GateNode fields —
	// a rename here is now a compile error instead of a silently-nil field.
	Wiring.RegisterBuilder("SelectRight",
		[]Wiring.PortSpec{
			{Name: "FromLeft", Dir: Wiring.PortIn},
			{Name: "FromRight", Dir: Wiring.PortIn},
			{Name: "ToPassed", Dir: Wiring.PortOut},
		},
		func(a Wiring.BuildArgs) (wire.Node, error) {
			n := &SelectRight{}
			n.Fire = a.Fire()
			n.EmitInputBeads = a.EmitInputBeads()
			n.Tick = a.Tick()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.FromLeft = a.In("FromLeft")
			n.FromRight = a.In("FromRight")
			n.ToPassed = a.Out("ToPassed")
			// EmitGeometry stays nil deliberately — nodeMover/edgeMover emit the same
			// geometry from their own goroutine start (see builders.go's note).
			// Left/HasLeft/Right/HasRight are runtime capture state, not injected —
			// they start at their Go zero-values (0/false) exactly as the retired reflection
			// left them (no matching tag/type for reflection to populate).
			return n, nil
		})
}
