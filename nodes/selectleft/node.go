package selectleft

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
	"github.com/dtauraso/wirefold/nodes/gatecommon"
)

// SelectLeft is the "SelectLeft" kind (registered as
// "SelectLeft" below — the name lives here in the comment,
// describing what its functions do). Its functions: Update runs the shared gate
// loop (gatecommon.RunGateAccept), which collects both inputs in the coincidence
// window and accepts the raw 10 pattern — FromLeft==1 AND FromRight==0 — DIRECTLY,
// no inversion/NOT. This package owns only the struct layout (required for
// gen-node-defs port discovery) and the init registration; GateNode is embedded so
// its port fields (FromLeft, FromRight, ToPassed) are promoted, and gen-node-defs'
// parseEmbeddedPorts follows the embedding to find them at BUILD time. (This was
// runtime reflectPorts before kinds constructed themselves; the struct still has to
// carry the fields, but nothing reads them reflectively any more — the ports the
// loader binds are the ones declared in the RegisterBuilder call below.)
type SelectLeft struct {
	gatecommon.GateNode
}

func (g *SelectLeft) Update(ctx context.Context) {
	gatecommon.RunGateAccept(ctx, &g.GateNode, 1, 0)
}

func init() {
	// SelectLeft CONSTRUCTS ITSELF. Every assignment below was previously performed by
	// Wiring.reflectBuild via reflection over the embedded gatecommon.GateNode fields —
	// a rename here is now a compile error instead of a silently-nil field.
	Wiring.RegisterBuilder("SelectLeft",
		[]Wiring.PortSpec{
			{Name: "FromLeft", Dir: Wiring.PortIn},
			{Name: "FromRight", Dir: Wiring.PortIn},
			{Name: "ToPassed", Dir: Wiring.PortOut},
		},
		func(a Wiring.BuildArgs) (wire.Node, error) {
			n := &SelectLeft{}
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
