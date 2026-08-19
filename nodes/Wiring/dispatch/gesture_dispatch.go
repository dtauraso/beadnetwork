package dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/gesture"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	T "github.com/dtauraso/wirefold/tools/topology-vscode/src/Trace"
)

func (md *MoveDispatch) HandleRawInput(ctx context.Context, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
	gesture.HandleRawInput(gesture.Deps{MR: &md.MR, UI: &md.UI, Mover: &md.Mover, RT: &md.RT, Ctx: ctx}, ev, slotReg, tr)
}
